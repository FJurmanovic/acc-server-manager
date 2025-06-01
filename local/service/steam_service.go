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
	SteamCMDPath    = "steamcmd"
	ACCServerAppID  = "1430110"
)

type SteamService struct {
	executor   *command.CommandExecutor
	repository *repository.SteamCredentialsRepository
}

func NewSteamService(repository *repository.SteamCredentialsRepository) *SteamService {
	return &SteamService{
		executor: &command.CommandExecutor{
			ExePath:   "powershell",
			LogOutput: true,
		},
		repository: repository,
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

func (s *SteamService) ensureSteamCMD() error {
	// Check if SteamCMD exists
	if _, err := os.Stat(SteamCMDPath); !os.IsNotExist(err) {
		return nil
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
		"Expand-Archive -Path 'steamcmd.zip' -DestinationPath 'steamcmd'"); err != nil {
		return fmt.Errorf("failed to extract SteamCMD: %v", err)
	}

	// Clean up zip file
	os.Remove("steamcmd.zip")
	return nil
}

func (s *SteamService) InstallServer(ctx context.Context, installPath string) error {
	if err := s.ensureSteamCMD(); err != nil {
		return err
	}

	// Ensure install path exists
	if err := os.MkdirAll(installPath, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	// Get Steam credentials
	creds, err := s.GetCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Steam credentials: %v", err)
	}

	// Build SteamCMD command
	args := []string{
		"-nologo",
		"-noprofile",
		filepath.Join(SteamCMDPath, "steamcmd.exe"),
		"+force_install_dir", installPath,
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
	logging.Info("Installing ACC server to %s...", installPath)
	if err := s.executor.Execute(args...); err != nil {
		return fmt.Errorf("failed to install server: %v", err)
	}

	return nil
}

func (s *SteamService) UpdateServer(ctx context.Context, installPath string) error {
	return s.InstallServer(ctx, installPath) // Same process as install
}

func (s *SteamService) UninstallServer(installPath string) error {
	return os.RemoveAll(installPath)
} 