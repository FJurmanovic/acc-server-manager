package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/command"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ACCServerAppID  = "1430110"
)

type SteamService struct {
	executor         *command.CommandExecutor
	repository      *repository.SteamCredentialsRepository
	configService   *SystemConfigService
}

func NewSteamService(repository *repository.SteamCredentialsRepository, configService *SystemConfigService) *SteamService {
	return &SteamService{
		executor: &command.CommandExecutor{
			ExePath:   "powershell",
			LogOutput: true,
		},
		repository:    repository,
		configService: configService,
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

func (s *SteamService) ensureSteamCMD(ctx context.Context) error {
	// Get SteamCMD path from config
	steamCMDPath, err := s.configService.GetSteamCMDDirPath(ctx)
	if err != nil {
		return fmt.Errorf("failed to get SteamCMD path from config: %v", err)
	}

	steamCMDDir := filepath.Dir(steamCMDPath)

	// Check if SteamCMD exists
	if _, err := os.Stat(steamCMDPath); !os.IsNotExist(err) {
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(steamCMDDir, 0755); err != nil {
		return fmt.Errorf("failed to create SteamCMD directory: %v", err)
	}

	// Download and install SteamCMD
	logging.Info("Downloading SteamCMD...")
	if err := s.executor.Execute("-Command",
		"Invoke-WebRequest -Uri 'https://steamcdn-a.akamaihd.net/client/installer/steamcmd.zip' -OutFile 'steamcmd.zip'"); err != nil {
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

func (s *SteamService) InstallServer(ctx context.Context, installPath string) error {
	if err := s.ensureSteamCMD(ctx); err != nil {
		return err
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

	// Get SteamCMD path from config
	steamCMDPath, err := s.configService.GetSteamCMDPath(ctx)
	if err != nil {
		return fmt.Errorf("failed to get SteamCMD path from config: %v", err)
	}

	// Build SteamCMD command
	args := []string{
		"-nologo",
		"-noprofile",
		steamCMDPath,
		"+force_install_dir", absPath,
		"+login",
	}

	if creds != nil && creds.Username != "" {
		args = append(args, creds.Username)
		if creds.Password != "" {
			args = append(args, creds.Password)
		}
	} else {
		args = append(args, "anonymous")
	}

	args = append(args,
		"+app_update", ACCServerAppID,
		"validate",
		"+quit",
	)

	// Run SteamCMD
	logging.Info("Installing ACC server to %s...", absPath)
	if err := s.executor.Execute(args...); err != nil {
		return fmt.Errorf("failed to run SteamCMD: %v", err)
	}

	// Add a delay to allow Steam to properly cleanup
	logging.Info("Waiting for Steam operations to complete...")
	if err := s.executor.Execute("-Command", "Start-Sleep -Seconds 5"); err != nil {
		logging.Warn("Failed to wait after Steam operations: %v", err)
	}

	// Verify installation
	exePath := filepath.Join(absPath, "server", "accServer.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		return fmt.Errorf("server installation failed: accServer.exe not found in %s", absPath)
	}

	logging.Info("Server installation completed successfully")
	return nil
}

func (s *SteamService) UpdateServer(ctx context.Context, installPath string) error {
	return s.InstallServer(ctx, installPath) // Same process as install
}

func (s *SteamService) UninstallServer(installPath string) error {
	return os.RemoveAll(installPath)
} 