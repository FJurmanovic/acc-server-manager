package platform

import (
	"acc-server-manager/local/model"
	"context"

	"github.com/google/uuid"
)

type ServerRuntime interface {
	Create(ctx context.Context, server *model.Server) error
	Delete(ctx context.Context, serverID uuid.UUID) error
	Start(ctx context.Context, serverID uuid.UUID) (string, error)
	Stop(ctx context.Context, serverID uuid.UUID) (string, error)
	Restart(ctx context.Context, serverID uuid.UUID) (string, error)
	Status(ctx context.Context, serverID uuid.UUID) (string, error)
}

type FirewallManager interface {
	CreateServerRules(serverName string, tcpPorts, udpPorts []int) error
	DeleteServerRules(serverName string, tcpPorts, udpPorts []int) error
	UpdateServerRules(serverName string, tcpPorts, udpPorts []int) error
}

type LogStreamer interface {
	Start(ctx context.Context, server *model.Server, handleLine func(string)) error
	Stop(serverID uuid.UUID)
	GetLastLines(ctx context.Context, server *model.Server, n int) ([]string, error)
}
