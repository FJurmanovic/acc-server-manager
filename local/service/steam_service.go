package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/command"
	"acc-server-manager/local/utl/env"
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/security"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ACCServerAppID = "1430110"
)

type SteamService struct {
	executor         *command.CommandExecutor
	repository       *repository.SteamCredentialsRepository
	tfaManager       *model.Steam2FAManager
	pathValidator    *security.PathValidator
	downloadVerifier *security.DownloadVerifier
}

func NewSteamService(repository *repository.SteamCredentialsRepository, tfaManager *model.Steam2FAManager) *SteamService {
	baseExecutor := &command.CommandExecutor{
		ExePath:   "powershell",
		LogOutput: true,
	}

	return &SteamService{
		executor:         baseExecutor,
		repository:       repository,
		tfaManager:       tfaManager,
		pathValidator:    security.NewPathValidator(),
		downloadVerifier: security.NewDownloadVerifier(),
	}
}

func (s *SteamService) GetCredentials(ctx context.Context) (*model.SteamCredentials, error) {
	return s.repository.GetCurrent(ctx)
}

func (s *SteamService) SaveCredentials(ctx context.Context, creds *model.SteamCredentials) error {
	if err := creds.Validate(); err != nil {
		return err
	}
	return s.repository.Save(ctx, creds)
}

func (s *SteamService) ensureSteamCMD(_ context.Context) error {
	steamCMDPath := env.GetSteamCMDPath()
	steamCMDDir := filepath.Dir(steamCMDPath)

	if _, err := os.Stat(steamCMDPath); !os.IsNotExist(err) {
		return nil
	}

	if err := os.MkdirAll(steamCMDDir, 0755); err != nil {
		return fmt.Errorf("failed to create SteamCMD directory: %v", err)
	}

	logging.Info("Downloading SteamCMD...")
	steamCMDZip := filepath.Join(steamCMDDir, "steamcmd.zip")
	if err := s.downloadVerifier.VerifyAndDownload(
		"https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip",
		steamCMDZip,
		""); err != nil {
		return fmt.Errorf("failed to download SteamCMD: %v", err)
	}

	logging.Info("Extracting SteamCMD...")
	if err := s.executor.Execute("-Command",
		fmt.Sprintf("Expand-Archive -Path 'steamcmd.zip' -DestinationPath '%s'", steamCMDDir)); err != nil {
		return fmt.Errorf("failed to extract SteamCMD: %v", err)
	}

	os.Remove("steamcmd.zip")
	return nil
}

func (s *SteamService) InstallServerWithWebSocket(ctx context.Context, installPath string, serverID *uuid.UUID, wsService *WebSocketService) error {
	if err := s.ensureSteamCMD(ctx); err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Error ensuring SteamCMD: %v", err), true)
		return err
	}

	if err := s.pathValidator.ValidateInstallPath(installPath); err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Invalid installation path: %v", err), true)
		return fmt.Errorf("invalid installation path: %v", err)
	}

	absPath, err := filepath.Abs(installPath)
	if err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Failed to get absolute path: %v", err), true)
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	absPath = filepath.Clean(absPath)

	if err := os.MkdirAll(absPath, 0755); err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Failed to create install directory: %v", err), true)
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Installation directory prepared: %s", absPath), false)

	creds, err := s.GetCredentials(ctx)
	if err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Failed to get Steam credentials: %v", err), true)
		return fmt.Errorf("failed to get Steam credentials: %v", err)
	}

	steamCMDPath := env.GetSteamCMDPath()

	steamCMDArgs := []string{
		"+force_install_dir", absPath,
		"+login",
	}

	if creds != nil && creds.Username != "" {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Using Steam credentials for user: %s", creds.Username), false)
		steamCMDArgs = append(steamCMDArgs, creds.Username)
		if creds.Password != "" {
			steamCMDArgs = append(steamCMDArgs, creds.Password)
		}
	} else {
		wsService.BroadcastSteamOutput(*serverID, "Using anonymous Steam login", false)
		steamCMDArgs = append(steamCMDArgs, "anonymous")
	}

	steamCMDArgs = append(steamCMDArgs,
		"+app_update", ACCServerAppID,
		"validate",
		"+quit",
	)

	args := steamCMDArgs

	wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Starting SteamCMD: %s %s", steamCMDPath, strings.Join(args, " ")), false)

	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	callbackConfig := &command.CallbackConfig{
		OnOutput: func(serverID uuid.UUID, output string, isError bool) {
			wsService.BroadcastSteamOutput(serverID, output, isError)
		},
		OnCommand: func(serverID uuid.UUID, command string, args []string, completed bool, success bool, error string) {
			if completed {
				if success {
					wsService.BroadcastSteamOutput(serverID, "Command completed successfully", false)
				} else {
					wsService.BroadcastSteamOutput(serverID, fmt.Sprintf("Command failed: %s", error), true)
				}
			}
		},
	}

	callbackInteractiveExecutor := command.NewCallbackInteractiveCommandExecutor(s.executor, s.tfaManager, callbackConfig, *serverID)
	callbackInteractiveExecutor.ExePath = steamCMDPath

	if err := callbackInteractiveExecutor.ExecuteInteractive(timeoutCtx, serverID, args...); err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("SteamCMD execution failed: %v", err), true)
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SteamCMD operation timed out after 15 minutes - this usually means Steam Guard confirmation is required")
		}
		return fmt.Errorf("failed to run SteamCMD: %v", err)
	}

	wsService.BroadcastSteamOutput(*serverID, "SteamCMD execution completed successfully, proceeding with verification...", false)

	wsService.BroadcastSteamOutput(*serverID, "Waiting for Steam operations to complete...", false)
	time.Sleep(5 * time.Second)

	exePath := filepath.Join(absPath, "server", "accServer.exe")
	wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Checking for ACC server executable at: %s", exePath), false)

	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		wsService.BroadcastSteamOutput(*serverID, "accServer.exe not found, checking directory contents...", false)

		if entries, dirErr := os.ReadDir(absPath); dirErr == nil {
			wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Contents of %s:", absPath), false)
			for _, entry := range entries {
				wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("  - %s (dir: %v)", entry.Name(), entry.IsDir()), false)
			}
		}

		serverDir := filepath.Join(absPath, "server")
		if entries, dirErr := os.ReadDir(serverDir); dirErr == nil {
			wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Contents of %s:", serverDir), false)
			for _, entry := range entries {
				wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("  - %s (dir: %v)", entry.Name(), entry.IsDir()), false)
			}
		} else {
			wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Server directory %s does not exist or cannot be read: %v", serverDir, dirErr), true)
		}

		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Server installation failed: accServer.exe not found in %s", exePath), true)
		return fmt.Errorf("server installation failed: accServer.exe not found in %s", exePath)
	}

	wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Server installation completed successfully - accServer.exe found at %s", exePath), false)
	return nil
}

func (s *SteamService) InstallServerWithCallbacks(ctx context.Context, installPath string, serverID *uuid.UUID, outputCallback command.OutputCallback) error {
	if err := s.ensureSteamCMD(ctx); err != nil {
		outputCallback(*serverID, fmt.Sprintf("Error ensuring SteamCMD: %v", err), true)
		return err
	}

	if err := s.pathValidator.ValidateInstallPath(installPath); err != nil {
		outputCallback(*serverID, fmt.Sprintf("Invalid installation path: %v", err), true)
		return fmt.Errorf("invalid installation path: %v", err)
	}

	absPath, err := filepath.Abs(installPath)
	if err != nil {
		outputCallback(*serverID, fmt.Sprintf("Failed to get absolute path: %v", err), true)
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	absPath = filepath.Clean(absPath)

	if err := os.MkdirAll(absPath, 0755); err != nil {
		outputCallback(*serverID, fmt.Sprintf("Failed to create install directory: %v", err), true)
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	outputCallback(*serverID, fmt.Sprintf("Installation directory prepared: %s", absPath), false)

	creds, err := s.GetCredentials(ctx)
	if err != nil {
		outputCallback(*serverID, fmt.Sprintf("Failed to get Steam credentials: %v", err), true)
		return fmt.Errorf("failed to get Steam credentials: %v", err)
	}

	steamCMDPath := env.GetSteamCMDPath()

	steamCMDArgs := []string{
		"+force_install_dir", absPath,
		"+login",
	}

	if creds != nil && creds.Username != "" {
		outputCallback(*serverID, fmt.Sprintf("Using Steam credentials for user: %s", creds.Username), false)
		steamCMDArgs = append(steamCMDArgs, creds.Username)
		if creds.Password != "" {
			steamCMDArgs = append(steamCMDArgs, creds.Password)
		}
	} else {
		outputCallback(*serverID, "Using anonymous Steam login", false)
		steamCMDArgs = append(steamCMDArgs, "anonymous")
	}

	steamCMDArgs = append(steamCMDArgs,
		"+app_update", ACCServerAppID,
		"validate",
		"+quit",
	)

	args := steamCMDArgs

	outputCallback(*serverID, fmt.Sprintf("Starting SteamCMD: %s %s", steamCMDPath, strings.Join(args, " ")), false)

	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	callbacks := &command.CallbackConfig{
		OnOutput: outputCallback,
	}

	callbackExecutor := command.NewCallbackInteractiveCommandExecutor(s.executor, s.tfaManager, callbacks, *serverID)
	callbackExecutor.ExePath = steamCMDPath

	if err := callbackExecutor.ExecuteInteractive(timeoutCtx, serverID, args...); err != nil {
		outputCallback(*serverID, fmt.Sprintf("SteamCMD execution failed: %v", err), true)
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SteamCMD operation timed out after 15 minutes - this usually means Steam Guard confirmation is required")
		}
		return fmt.Errorf("failed to run SteamCMD: %v", err)
	}

	outputCallback(*serverID, "SteamCMD execution completed successfully, proceeding with verification...", false)

	outputCallback(*serverID, "Waiting for Steam operations to complete...", false)
	time.Sleep(5 * time.Second)

	exePath := filepath.Join(absPath, "server", "accServer.exe")
	outputCallback(*serverID, fmt.Sprintf("Checking for ACC server executable at: %s", exePath), false)

	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		outputCallback(*serverID, "accServer.exe not found, checking directory contents...", false)

		if entries, dirErr := os.ReadDir(absPath); dirErr == nil {
			outputCallback(*serverID, fmt.Sprintf("Contents of %s:", absPath), false)
			for _, entry := range entries {
				outputCallback(*serverID, fmt.Sprintf("  - %s (dir: %v)", entry.Name(), entry.IsDir()), false)
			}
		}

		serverDir := filepath.Join(absPath, "server")
		if entries, dirErr := os.ReadDir(serverDir); dirErr == nil {
			outputCallback(*serverID, fmt.Sprintf("Contents of %s:", serverDir), false)
			for _, entry := range entries {
				outputCallback(*serverID, fmt.Sprintf("  - %s (dir: %v)", entry.Name(), entry.IsDir()), false)
			}
		} else {
			outputCallback(*serverID, fmt.Sprintf("Server directory %s does not exist or cannot be read: %v", serverDir, dirErr), true)
		}

		outputCallback(*serverID, fmt.Sprintf("Server installation failed: accServer.exe not found in %s", exePath), true)
		return fmt.Errorf("server installation failed: accServer.exe not found in %s", exePath)
	}

	outputCallback(*serverID, fmt.Sprintf("Server installation completed successfully - accServer.exe found at %s", exePath), false)
	return nil
}

func (s *SteamService) UninstallServer(installPath string) error {
	return os.RemoveAll(installPath)
}
