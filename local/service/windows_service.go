package service

import (
	"acc-server-manager/local/utl/command"
	"acc-server-manager/local/utl/logging"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	NSSMPath = ".\\nssm"
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

// executeNSSM runs an NSSM command through PowerShell with elevation
func (s *WindowsService) executeNSSM(args ...string) (string, error) {
	// Prepend NSSM path to arguments
	nssmArgs := append([]string{"-nologo", "-noprofile", NSSMPath}, args...)
	
	output, err := s.executor.ExecuteWithOutput(nssmArgs...)
	if err != nil {
		return "", err
	}

	// Clean up output by removing null bytes and trimming whitespace
	cleaned := strings.TrimSpace(strings.ReplaceAll(output, "\x00", ""))
	// Remove \r\n from status strings
	cleaned = strings.TrimSuffix(cleaned, "\r\n")
	
	return cleaned, nil
}

// Service Installation/Configuration Methods

func (s *WindowsService) CreateService(serviceName, execPath, workingDir string, args []string) error {
	// Ensure paths are absolute
	absExecPath, err := filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for executable: %v", err)
	}

	absWorkingDir, err := filepath.Abs(workingDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for working directory: %v", err)
	}

	// Install service
	if _, err := s.executeNSSM("install", serviceName, absExecPath); err != nil {
		return fmt.Errorf("failed to install service: %v", err)
	}

	// Set working directory
	if _, err := s.executeNSSM("set", serviceName, "AppDirectory", absWorkingDir); err != nil {
		return fmt.Errorf("failed to set working directory: %v", err)
	}

	// Set arguments if provided
	if len(args) > 0 {
		cmdArgs := append([]string{"set", serviceName, "AppParameters"}, args...)
		if _, err := s.executeNSSM(cmdArgs...); err != nil {
			return fmt.Errorf("failed to set arguments: %v", err)
		}
	}

	logging.Info("Created Windows service: %s", serviceName)
	return nil
}

func (s *WindowsService) DeleteService(serviceName string) error {
	if _, err := s.executeNSSM("remove", serviceName, "confirm"); err != nil {
		return fmt.Errorf("failed to remove service: %v", err)
	}

	logging.Info("Removed Windows service: %s", serviceName)
	return nil
}

func (s *WindowsService) UpdateService(serviceName, execPath, workingDir string, args []string) error {
	// First remove the existing service
	if err := s.DeleteService(serviceName); err != nil {
		return err
	}

	// Then create it again with new parameters
	return s.CreateService(serviceName, execPath, workingDir, args)
}

// Service Control Methods

func (s *WindowsService) Status(serviceName string) (string, error) {
	return s.executeNSSM("status", serviceName)
}

func (s *WindowsService) Start(serviceName string) (string, error) {
	return s.executeNSSM("start", serviceName)
}

func (s *WindowsService) Stop(serviceName string) (string, error) {
	return s.executeNSSM("stop", serviceName)
}

func (s *WindowsService) Restart(serviceName string) (string, error) {
	// First stop the service
	if _, err := s.Stop(serviceName); err != nil {
		return "", err
	}

	// Then start it again
	return s.Start(serviceName)
} 