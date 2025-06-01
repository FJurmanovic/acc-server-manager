package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"
	"context"

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

	// Provide caches
	logging.Debug("Creating lookup cache instance")
	c.Provide(func() *model.LookupCache {
		return model.NewLookupCache()
	})

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
	err := c.Invoke(func(server *ServerService, api *ApiService, config *ConfigService, lookup *LookupService, systemConfig *SystemConfigService) {
		logging.Debug("Setting up service cross-references")
		api.SetServerService(server)
		config.SetServerService(server)
		
		logging.Debug("Initializing lookup data cache")
		// Initialize lookup data using repository directly
		lookup.cache.Set("tracks", lookup.repository.GetTracks(context.Background()))
		lookup.cache.Set("cars", lookup.repository.GetCarModels(context.Background()))
		lookup.cache.Set("drivers", lookup.repository.GetDriverCategories(context.Background()))
		lookup.cache.Set("cups", lookup.repository.GetCupCategories(context.Background()))
		lookup.cache.Set("sessions", lookup.repository.GetSessionTypes(context.Background()))
		logging.Debug("Completed initializing lookup data cache")

		logging.Debug("Initializing system config service")
		// Initialize system config service
		if err := systemConfig.Initialize(context.Background()); err != nil {
			logging.Panic("failed to initialize system config service: " + err.Error())
		}
		logging.Debug("Completed initializing system config service")
	})
	if err != nil {
		logging.Panic("unable to initialize services: " + err.Error())
	}
	logging.Debug("Completed service initialization")
}
