package service

import (
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"

	"go.uber.org/dig"
)

// *dig.Container: Dig Container
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
	c.Provide(NewWindowsService)
	c.Provide(NewFirewallService)
	c.Provide(NewMembershipService)
	c.Provide(NewWebSocketService)

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
