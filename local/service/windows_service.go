package service

import (
	"acc-server-manager/local/utl/command"
	"acc-server-manager/local/utl/env"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type WindowsService struct {
	executor *command.CommandExecutor
}

func NewWindowsService() *WindowsService {
	return &WindowsService{
		executor: &command.CommandExecutor{
			ExePath:   "powershell",
			LogOutput: true,
		},
	}
}

func (s *WindowsService) ExecuteNSSM(ctx context.Context, args ...string) (string, error) {
	nssmPath := env.GetNSSMPath()

	nssmArgs := append([]string{"-NoProfile", "-NonInteractive", "-Command", "& " + nssmPath}, args...)

	output, err := s.executor.ExecuteWithOutput(nssmArgs...)
	if err != nil {
		logging.Error("NSSM command failed: powershell %s", strings.Join(nssmArgs, " "))
		logging.Error("NSSM error output: %s", output)
		return "", err
	}

	cleaned := strings.TrimSpace(strings.ReplaceAll(output, "\x00", ""))
	cleaned = strings.TrimSuffix(cleaned, "\r\n")

	return cleaned, nil
}

func (s *WindowsService) CreateService(ctx context.Context, serviceName, execPath, workingDir string, args []string) error {
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

	logging.Info("Creating service '%s' with:", serviceName)
	logging.Info("  Executable: %s", absExecPath)
	logging.Info("  Working Directory: %s", absWorkingDir)

	s.ExecuteNSSM(ctx, "remove", serviceName, "confirm")

	if _, err := s.ExecuteNSSM(ctx, "install", serviceName, absExecPath); err != nil {
		return fmt.Errorf("failed to install service: %v", err)
	}

	if len(args) > 0 {
		cmdArgs := append([]string{"set", serviceName, "AppParameters"}, args...)
		if _, err := s.ExecuteNSSM(ctx, cmdArgs...); err != nil {
			s.ExecuteNSSM(ctx, "remove", serviceName, "confirm")
			return fmt.Errorf("failed to set arguments: %v", err)
		}
	}

	if _, err := s.ExecuteNSSM(ctx, "get", serviceName, "Application"); err != nil {
		return fmt.Errorf("service creation verification failed: %v", err)
	}

	logging.Info("Created Windows service: %s", serviceName)
	return nil
}

func (s *WindowsService) DeleteService(ctx context.Context, serviceName string) error {
	if _, err := s.ExecuteNSSM(ctx, "remove", serviceName, "confirm"); err != nil {
		return fmt.Errorf("failed to remove service: %v", err)
	}

	logging.Info("Removed Windows service: %s", serviceName)
	return nil
}

func (s *WindowsService) UpdateService(ctx context.Context, serviceName, execPath, workingDir string, args []string) error {
	if err := s.DeleteService(ctx, serviceName); err != nil {
		return err
	}

	return s.CreateService(ctx, serviceName, execPath, workingDir, args)
}

func (s *WindowsService) Status(ctx context.Context, serviceName string) (string, error) {
	return s.ExecuteNSSM(ctx, "status", serviceName)
}

func (s *WindowsService) Start(ctx context.Context, serviceName string) (string, error) {
	return s.ExecuteNSSM(ctx, "start", serviceName)
}

func (s *WindowsService) Stop(ctx context.Context, serviceName string) (string, error) {
	return s.ExecuteNSSM(ctx, "stop", serviceName)
}

func (s *WindowsService) Restart(ctx context.Context, serviceName string) (string, error) {
	if _, err := s.Stop(ctx, serviceName); err != nil {
		return "", err
	}

	return s.Start(ctx, serviceName)
}
