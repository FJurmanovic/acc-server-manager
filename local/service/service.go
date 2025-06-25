package service

import (
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"

	"go.uber.org/dig"
)

// InitializeServices
// Initializes Dependency Injection modules for services
//
//	Args:
//		*dig.Container: Dig Container
func InitializeServices(c *dig.Container) {
	logging.Debug("Initializing repositories")
	repository.InitializeRepositories(c)

	logging.Debug("Registering services")
	// Provide services
	c.Provide(NewServerService)
	c.Provide(NewStateHistoryService)
	c.Provide(NewApiService)
	c.Provide(NewConfigService)
	c.Provide(NewLookupService)
	c.Provide(NewSystemConfigService)
	c.Provide(NewSteamService)
	c.Provide(NewWindowsService)
	c.Provide(NewFirewallService)

	logging.Debug("Initializing service dependencies")
	err := c.Invoke(func(server *ServerService, api *ApiService, config *ConfigService, systemConfig *SystemConfigService) {
		logging.Debug("Setting up service cross-references")
		api.SetServerService(server)
		config.SetServerService(server)


	})
	if err != nil {
		logging.Panic("unable to initialize services: " + err.Error())
	}
	logging.Debug("Completed service initialization")
}
