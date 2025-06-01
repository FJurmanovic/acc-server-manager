package service

import (
	"acc-server-manager/local/utl/command"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	NSSMPath = ".\\nssm.exe"
)

type WindowsService struct {
	executor      *command.CommandExecutor
	configService *SystemConfigService
}

func NewWindowsService(configService *SystemConfigService) *WindowsService {
	return &WindowsService{
		executor: &command.CommandExecutor{
			ExePath:   "powershell",
			LogOutput: true,
		},
		configService: configService,
	}
}

// executeNSSM runs an NSSM command through PowerShell with elevation
func (s *WindowsService) executeNSSM(ctx context.Context, args ...string) (string, error) {
	// Get NSSM path from config
	nssmPath, err := s.configService.GetNSSMPath(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get NSSM path from config: %v", err)
	}

	// Prepend NSSM path to arguments
	nssmArgs := append([]string{"-NoProfile", "-NonInteractive", "-Command", "& " + nssmPath}, args...)
	
	output, err := s.executor.ExecuteWithOutput(nssmArgs...)
	if err != nil {
		// Log the full command and error for debugging
		logging.Error("NSSM command failed: powershell %s", strings.Join(nssmArgs, " "))
		logging.Error("NSSM error output: %s", output)
		return "", err
	}

	// Clean up output by removing null bytes and trimming whitespace
	cleaned := strings.TrimSpace(strings.ReplaceAll(output, "\x00", ""))
	// Remove \r\n from status strings
	cleaned = strings.TrimSuffix(cleaned, "\r\n")
	
	return cleaned, nil
}

// Service Installation/Configuration Methods

func (s *WindowsService) CreateService(ctx context.Context, serviceName, execPath, workingDir string, args []string) error {
	// Ensure paths are absolute and properly formatted for Windows
	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for executable: %v", err)
	}
	absExecPath = filepath.Clean(absExecPath)

	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for working directory: %v", err)
	}
	absWorkingDir = filepath.Clean(absWorkingDir)

	// Log the paths being used
	logging.Info("Creating service '%s' with:", serviceName)
	logging.Info("  Executable: %s", absExecPath)
	logging.Info("  Working Directory: %s", absWorkingDir)

	// First remove any existing service with the same name
	s.executeNSSM(ctx, "remove", serviceName, "confirm")

	// Install service
	if _, err := s.executeNSSM(ctx, "install", serviceName, absExecPath); err != nil {
		return fmt.Errorf("failed to install service: %v", err)
	}

	// Set arguments if provided
	if len(args) > 0 {
		cmdArgs := append([]string{"set", serviceName, "AppParameters"}, args...)
		if _, err := s.executeNSSM(ctx, cmdArgs...); err != nil {
			// Try to clean up on failure
			s.executeNSSM(ctx, "remove", serviceName, "confirm")
			return fmt.Errorf("failed to set arguments: %v", err)
		}
	}

	// Verify service was created
	if _, err := s.executeNSSM(ctx, "get", serviceName, "Application"); err != nil {
		return fmt.Errorf("service creation verification failed: %v", err)
	}

	logging.Info("Created Windows service: %s", serviceName)
	return nil
}

func (s *WindowsService) DeleteService(ctx context.Context, serviceName string) error {
	if _, err := s.executeNSSM(ctx, "remove", serviceName, "confirm"); err != nil {
		return fmt.Errorf("failed to remove service: %v", err)
	}

	logging.Info("Removed Windows service: %s", serviceName)
	return nil
}

func (s *WindowsService) UpdateService(ctx context.Context, serviceName, execPath, workingDir string, args []string) error {
	// First remove the existing service
	if err := s.DeleteService(ctx, serviceName); err != nil {
		return err
	}

	// Then create it again with new parameters
	return s.CreateService(ctx, serviceName, execPath, workingDir, args)
}

// Service Control Methods

func (s *WindowsService) Status(ctx context.Context, serviceName string) (string, error) {
	return s.executeNSSM(ctx, "status", serviceName)
}

func (s *WindowsService) Start(ctx context.Context, serviceName string) (string, error) {
	return s.executeNSSM(ctx, "start", serviceName)
}

func (s *WindowsService) Stop(ctx context.Context, serviceName string) (string, error) {
	return s.executeNSSM(ctx, "stop", serviceName)
}

func (s *WindowsService) Restart(ctx context.Context, serviceName string) (string, error) {
	// First stop the service
	if _, err := s.Stop(ctx, serviceName); err != nil {
		return "", err
	}

	// Then start it again
	return s.Start(ctx, serviceName)
} 