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
		steamCMDArgs = append(steamCMDArgs, creds.Username)
		if creds.Password != "" {
			steamCMDArgs = append(steamCMDArgs, creds.Password)
		}
	} else {
		steamCMDArgs = append(steamCMDArgs, "anonymous")
	}

	steamCMDArgs = append(steamCMDArgs,
		"+app_update", ACCServerAppID,
		"validate",
		"+quit",
	)

	// Build PowerShell arguments to execute SteamCMD directly
	// This matches the format: powershell -nologo -noprofile c:\steamcmd\steamcmd.exe +args...
	args := []string{"-nologo", "-noprofile"}
	args = append(args, steamCMDPath)
	args = append(args, steamCMDArgs...)

	// Use interactive executor to handle potential 2FA prompts with timeout
	logging.Info("Installing ACC server to %s...", absPath)
	
	// Create a context with timeout to prevent hanging indefinitely
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	
	if err := s.interactiveExecutor.ExecuteInteractive(timeoutCtx, serverID, args...); err != nil {
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("SteamCMD operation timed out after 10 minutes")
		}
		return fmt.Errorf("failed to run SteamCMD: %v", err)
	}

	// Add a delay to allow Steam to properly cleanup
	logging.Info("Waiting for Steam operations to complete...")
	time.Sleep(5 * time.Second)

	// Verify installation
	exePath := filepath.Join(absPath, "server", "accServer.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		return fmt.Errorf("server installation failed: accServer.exe not found in %s", absPath)
	}

	logging.Info("Server installation completed successfully")
	return nil
}

func (s *SteamService) UpdateServer(ctx context.Context, installPath string, serverID *uuid.UUID) error {
	return s.InstallServer(ctx, installPath, serverID) // Same process as install
}

func (s *SteamService) UninstallServer(installPath string) error {
	return os.RemoveAll(installPath)
}
