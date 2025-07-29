package service

import (
	"acc-server-manager/local/utl/command"
	"acc-server-manager/local/utl/logging"
	"fmt"
)

type FirewallService struct {
	executor *command.CommandExecutor
}

func NewFirewallService() *FirewallService {
	return &FirewallService{
		executor: &command.CommandExecutor{
			ExePath:   "netsh",
			LogOutput: true,
		},
	}
}

func (s *FirewallService) CreateServerRules(serverName string, tcpPorts, udpPorts []int) error {
	for _, port := range tcpPorts {
		ruleName := fmt.Sprintf("\"%s-TCP-%d\"", serverName, port)
		builder := command.NewCommandBuilder().
			Add("advfirewall").
			Add("firewall").
			Add("add").
			Add("rule").
			AddFlag("name", ruleName).
			AddFlag("dir", "in").
			AddFlag("action", "allow").
			AddFlag("protocol", "TCP").
			AddFlag("localport", port)

		if err := s.executor.ExecuteWithBuilder(builder); err != nil {
			return fmt.Errorf("failed to create TCP firewall rule for port %d: %v", port, err)
		}
		logging.Info("Created TCP firewall rule: %s", ruleName)
	}

	for _, port := range udpPorts {
		ruleName := fmt.Sprintf("%s-UDP-%d", serverName, port)
		builder := command.NewCommandBuilder().
			Add("advfirewall").
			Add("firewall").
			Add("add").
			Add("rule").
			AddFlag("name", ruleName).
			AddFlag("dir", "in").
			AddFlag("action", "allow").
			AddFlag("protocol", "UDP").
			AddFlag("localport", port)

		if err := s.executor.ExecuteWithBuilder(builder); err != nil {
			return fmt.Errorf("failed to create UDP firewall rule for port %d: %v", port, err)
		}
		logging.Info("Created UDP firewall rule: %s", ruleName)
	}

	return nil
}

func (s *FirewallService) DeleteServerRules(serverName string, tcpPorts, udpPorts []int) error {
	for _, port := range tcpPorts {
		ruleName := fmt.Sprintf("\"%s-TCP-%d\"", serverName, port)
		builder := command.NewCommandBuilder().
			Add("advfirewall").
			Add("firewall").
			Add("delete").
			Add("rule").
			AddFlag("name", ruleName)

		if err := s.executor.ExecuteWithBuilder(builder); err != nil {
			return fmt.Errorf("failed to delete TCP firewall rule for port %d: %v", port, err)
		}
		logging.Info("Deleted TCP firewall rule: %s", ruleName)
	}

	for _, port := range udpPorts {
		ruleName := fmt.Sprintf("\"%s-UDP-%d\"", serverName, port)
		builder := command.NewCommandBuilder().
			Add("advfirewall").
			Add("firewall").
			Add("delete").
			Add("rule").
			AddFlag("name", ruleName)

		if err := s.executor.ExecuteWithBuilder(builder); err != nil {
			return fmt.Errorf("failed to delete UDP firewall rule for port %d: %v", port, err)
		}
		logging.Info("Deleted UDP firewall rule: %s", ruleName)
	}

	return nil
}

func (s *FirewallService) UpdateServerRules(serverName string, tcpPorts, udpPorts []int) error {
	// First delete existing rules
	if err := s.DeleteServerRules(serverName, tcpPorts, udpPorts); err != nil {
		return err
	}

	// Then create new rules
	return s.CreateServerRules(serverName, tcpPorts, udpPorts)
} 