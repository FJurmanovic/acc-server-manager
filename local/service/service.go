package service

import (
	"acc-server-manager/local/platform"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/docker"
	"acc-server-manager/local/utl/env"
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/network"
	"fmt"

	dockerclient "github.com/docker/docker/client"
	"go.uber.org/dig"
	"gorm.io/gorm"
)

// InitializeServices registers all service providers into the DI container.
// Platform-specific components (ServerRuntime, FirewallManager, LogStreamer)
// are conditionally provided based on the PLATFORM environment variable.
func InitializeServices(c *dig.Container) {
	logging.Debug("Initializing repositories")
	repository.InitializeRepositories(c)

	logging.Debug("Registering services")
	c.Provide(NewSteamService)
	c.Provide(NewServerService)
	c.Provide(NewStateHistoryService)
	c.Provide(NewServiceControlService)
	c.Provide(NewConfigService)
	c.Provide(NewLookupService)
	c.Provide(NewMembershipService)
	c.Provide(NewWebSocketService)
	c.Provide(NewLeaderboardService)

	// Always provide the PortPoolManager; in Windows mode it seeds from an
	// empty result set and sits idle.
	c.Provide(func(db *gorm.DB) (*network.PortPoolManager, error) {
		return network.NewPortPoolManager(db)
	})

	if env.IsDockerPlatform() {
		registerDockerProviders(c)
	} else {
		registerWindowsProviders(c)
	}

	logging.Debug("Initializing service dependencies")
	err := c.Invoke(func(server *ServerService, api *ServiceControlService, config *ConfigService) {
		logging.Debug("Setting up service cross-references")
		api.SetServerService(server)
		config.SetServerService(server)
	})
	if err != nil {
		logging.Panic("unable to initialize services: " + err.Error())
	}
	logging.Debug("Completed service initialization")
}

func registerWindowsProviders(c *dig.Container) {
	logging.Debug("Registering Windows platform providers")

	c.Provide(NewWindowsService)

	c.Provide(func(ws *WindowsService, repo *repository.ServerRepository) platform.ServerRuntime {
		return NewWindowsRuntime(ws, repo)
	})

	c.Provide(func() platform.FirewallManager {
		return NewFirewallService()
	})

	c.Provide(func() platform.LogStreamer {
		return NewFileLogStreamer()
	})
}

func registerDockerProviders(c *dig.Container) {
	logging.Debug("Registering Docker platform providers")

	if err := c.Provide(func() (*dockerclient.Client, error) {
		client, err := docker.NewDockerClient()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to Docker daemon: %v", err)
		}
		return client, nil
	}); err != nil {
		logging.Panic("failed to provide Docker client: " + err.Error())
	}

	c.Provide(func(client *dockerclient.Client, repo *repository.ServerRepository) platform.ServerRuntime {
		return NewDockerRuntime(client, repo)
	})

	c.Provide(func() platform.FirewallManager {
		return &NoOpFirewall{}
	})

	c.Provide(func(client *dockerclient.Client) platform.LogStreamer {
		return NewDockerLogStreamer(client)
	})
}
