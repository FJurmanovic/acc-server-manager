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
	executor            *command.CommandExecutor
	interactiveExecutor *command.InteractiveCommandExecutor
	repository          *repository.SteamCredentialsRepository
	tfaManager          *model.Steam2FAManager
	pathValidator       *security.PathValidator
	downloadVerifier    *security.DownloadVerifier
}

func NewSteamService(repository *repository.SteamCredentialsRepository, tfaManager *model.Steam2FAManager) *SteamService {
	baseExecutor := &command.CommandExecutor{
		ExePath:   "powershell",
		LogOutput: true,
	}

	return &SteamService{
		executor:            baseExecutor,
		interactiveExecutor: command.NewInteractiveCommandExecutor(baseExecutor, tfaManager),
		repository:          repository,
		tfaManager:          tfaManager,
		pathValidator:       security.NewPathValidator(),
		downloadVerifier:    security.NewDownloadVerifier(),
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
	// Get SteamCMD path from environment variable
	steamCMDPath := env.GetSteamCMDPath()
	steamCMDDir := filepath.Dir(steamCMDPath)

	// Check if SteamCMD exists
	if _, err := os.Stat(steamCMDPath); !os.IsNotExist(err) {
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(steamCMDDir, 0755); err != nil {
		return fmt.Errorf("failed to create SteamCMD directory: %v", err)
	}

	// Download and install SteamCMD securely
	logging.Info("Downloading SteamCMD...")
	steamCMDZip := filepath.Join(steamCMDDir, "steamcmd.zip")
	if err := s.downloadVerifier.VerifyAndDownload(
		"https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip",
		steamCMDZip,
		""); err != nil {
		return fmt.Errorf("failed to download SteamCMD: %v", err)
	}

	// Extract SteamCMD
	logging.Info("Extracting SteamCMD...")
	if err := s.executor.Execute("-Command",
		fmt.Sprintf("Expand-Archive -Path 'steamcmd.zip' -DestinationPath '%s'", steamCMDDir)); err != nil {
		return fmt.Errorf("failed to extract SteamCMD: %v", err)
	}

	// Clean up zip file
	os.Remove("steamcmd.zip")
	return nil
}

func (s *SteamService) InstallServer(ctx context.Context, installPath string, serverID *uuid.UUID) error {
	if err := s.ensureSteamCMD(ctx); err != nil {
		return err
	}

	// Validate installation path for security
	if err := s.pathValidator.ValidateInstallPath(installPath); err != nil {
		return fmt.Errorf("invalid installation path: %v", err)
	}

	// Convert to absolute path and ensure proper Windows path format
	absPath, err := filepath.Abs(installPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	absPath = filepath.Clean(absPath)

	// Ensure install path exists
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	// Get Steam credentials
	creds, err := s.GetCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Steam credentials: %v", err)
	}

	// Get SteamCMD path from environment variable
	steamCMDPath := env.GetSteamCMDPath()

	// Build SteamCMD command arguments
	steamCMDArgs := []string{
		"+force_install_dir", absPath,
		"+login",
	}

	if creds != nil && creds.Username != "" {
		logging.Info("Using Steam credentials for user: %s", creds.Username)
		steamCMDArgs = append(steamCMDArgs, creds.Username)
		if creds.Password != "" {
			steamCMDArgs = append(steamCMDArgs, creds.Password)
		}
	} else {
		logging.Info("Using anonymous Steam login")
		steamCMDArgs = append(steamCMDArgs, "anonymous")
	}

	steamCMDArgs = append(steamCMDArgs,
		"+app_update", ACCServerAppID,
		"validate",
		"+quit",
	)

	// Execute SteamCMD directly without PowerShell wrapper to get better output capture
	args := steamCMDArgs

	// Use interactive executor to handle potential 2FA prompts with timeout
	logging.Info("Installing ACC server to %s...", absPath)
	logging.Info("SteamCMD command: %s %s", steamCMDPath, strings.Join(args, " "))

	// Create a context with timeout to prevent hanging indefinitely
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Minute) // Increased timeout
	defer cancel()

	// Update the executor to use SteamCMD directly
	originalExePath := s.interactiveExecutor.ExePath
	s.interactiveExecutor.ExePath = steamCMDPath
	defer func() {
		s.interactiveExecutor.ExePath = originalExePath
	}()

	if err := s.interactiveExecutor.ExecuteInteractive(timeoutCtx, serverID, args...); err != nil {
		logging.Error("SteamCMD execution failed: %v", err)
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SteamCMD operation timed out after 15 minutes - this usually means Steam Guard confirmation is required")
		}
		return fmt.Errorf("failed to run SteamCMD: %v", err)
	}

	logging.Info("SteamCMD execution completed successfully, proceeding with verification...")

	// Add a delay to allow Steam to properly cleanup
	logging.Info("Waiting for Steam operations to complete...")
	time.Sleep(5 * time.Second)

	// Verify installation
	exePath := filepath.Join(absPath, "server", "accServer.exe")
	logging.Info("Checking for ACC server executable at: %s", exePath)

	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		// Log directory contents to help debug
		logging.Info("accServer.exe not found, checking directory contents...")
		if entries, dirErr := os.ReadDir(absPath); dirErr == nil {
			logging.Info("Contents of %s:", absPath)
			for _, entry := range entries {
				logging.Info("  - %s (dir: %v)", entry.Name(), entry.IsDir())
			}
		}

		// Check if there's a server subdirectory
		serverDir := filepath.Join(absPath, "server")
		if entries, dirErr := os.ReadDir(serverDir); dirErr == nil {
			logging.Info("Contents of %s:", serverDir)
			for _, entry := range entries {
				logging.Info("  - %s (dir: %v)", entry.Name(), entry.IsDir())
			}
		} else {
			logging.Info("Server directory %s does not exist or cannot be read: %v", serverDir, dirErr)
		}

		return fmt.Errorf("server installation failed: accServer.exe not found in %s", exePath)
	}

	logging.Info("Server installation completed successfully - accServer.exe found at %s", exePath)
	return nil
}

// InstallServerWithWebSocket installs a server with WebSocket output streaming
func (s *SteamService) InstallServerWithWebSocket(ctx context.Context, installPath string, serverID *uuid.UUID, wsService *WebSocketService) error {
	if err := s.ensureSteamCMD(ctx); err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Error ensuring SteamCMD: %v", err), true)
		return err
	}

	// Validate installation path for security
	if err := s.pathValidator.ValidateInstallPath(installPath); err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Invalid installation path: %v", err), true)
		return fmt.Errorf("invalid installation path: %v", err)
	}

	// Convert to absolute path and ensure proper Windows path format
	absPath, err := filepath.Abs(installPath)
	if err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Failed to get absolute path: %v", err), true)
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	absPath = filepath.Clean(absPath)

	// Ensure install path exists
	if err := os.MkdirAll(absPath, 0755); err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Failed to create install directory: %v", err), true)
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Installation directory prepared: %s", absPath), false)

	// Get Steam credentials
	creds, err := s.GetCredentials(ctx)
	if err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Failed to get Steam credentials: %v", err), true)
		return fmt.Errorf("failed to get Steam credentials: %v", err)
	}

	// Get SteamCMD path from environment variable
	steamCMDPath := env.GetSteamCMDPath()

	// Build SteamCMD command arguments
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

	// Execute SteamCMD with WebSocket output streaming
	args := steamCMDArgs

	wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("Starting SteamCMD: %s %s", steamCMDPath, strings.Join(args, " ")), false)

	// Create a context with timeout to prevent hanging indefinitely
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	// Update the executor to use SteamCMD directly
	originalExePath := s.interactiveExecutor.ExePath
	s.interactiveExecutor.ExePath = steamCMDPath
	defer func() {
		s.interactiveExecutor.ExePath = originalExePath
	}()

	// Create a modified interactive executor that streams output to WebSocket
	wsInteractiveExecutor := command.NewInteractiveCommandExecutorWithWebSocket(s.executor, s.tfaManager, wsService, *serverID)
	wsInteractiveExecutor.ExePath = steamCMDPath

	if err := wsInteractiveExecutor.ExecuteInteractive(timeoutCtx, serverID, args...); err != nil {
		wsService.BroadcastSteamOutput(*serverID, fmt.Sprintf("SteamCMD execution failed: %v", err), true)
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SteamCMD operation timed out after 15 minutes - this usually means Steam Guard confirmation is required")
		}
		return fmt.Errorf("failed to run SteamCMD: %v", err)
	}

	wsService.BroadcastSteamOutput(*serverID, "SteamCMD execution completed successfully, proceeding with verification...", false)

	// Add a delay to allow Steam to properly cleanup
	wsService.BroadcastSteamOutput(*serverID, "Waiting for Steam operations to complete...", false)
	time.Sleep(5 * time.Second)

	// Verify installation
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

		// Check if there's a server subdirectory
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

func (s *SteamService) UpdateServer(ctx context.Context, installPath string, serverID *uuid.UUID) error {
	return s.InstallServer(ctx, installPath, serverID) // Same process as install
}

func (s *SteamService) UninstallServer(installPath string) error {
	return os.RemoveAll(installPath)
}
