package service

import (
	"acc-server-manager/local/utl/command"
	"strings"
)

type ServiceManager struct {
	executor *command.CommandExecutor
	psExecutor *command.CommandExecutor
}

func NewServiceManager() *ServiceManager {
	return &ServiceManager{
		executor: &command.CommandExecutor{
			ExePath:   "nssm",
			LogOutput: true,
		},
		psExecutor: &command.CommandExecutor{
			ExePath:   "powershell",
			LogOutput: true,
		},
	}
}

func (s *ServiceManager) ManageService(serviceName, action string) (string, error) {
	// Run NSSM command through PowerShell to ensure elevation
	output, err := s.psExecutor.ExecuteWithOutput("-nologo", "-noprofile", ".\\nssm", action, serviceName)
	if err != nil {
		return "", err
	}

	// Clean up output by removing null bytes and trimming whitespace
	cleaned := strings.TrimSpace(strings.ReplaceAll(output, "\x00", ""))
	// Remove \r\n from status strings
	cleaned = strings.TrimSuffix(cleaned, "\r\n")
	
	return cleaned, nil
}

func (s *ServiceManager) Status(serviceName string) (string, error) {
	return s.ManageService(serviceName, "status")
}

func (s *ServiceManager) Start(serviceName string) (string, error) {
	return s.ManageService(serviceName, "start")
}

func (s *ServiceManager) Stop(serviceName string) (string, error) {
	return s.ManageService(serviceName, "stop")
}

func (s *ServiceManager) Restart(serviceName string) (string, error) {
	// First stop the service
	if _, err := s.Stop(serviceName); err != nil {
		return "", err
	}

	// Then start it again
	return s.Start(serviceName)
} 