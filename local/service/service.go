package service

import (
	"acc-server-manager/local/repository"
	"context"
	"log"

	"go.uber.org/dig"
)

// InitializeServices
// Initializes Dependency Injection modules for services
//
//	Args:
//		*dig.Container: Dig Container
func InitializeServices(c *dig.Container) {
	repository.InitializeRepositories(c)

	c.Provide(NewServerService)
	c.Provide(NewStateHistoryService)
	c.Provide(NewApiService)
	c.Provide(NewConfigService)
	c.Provide(NewLookupService)

	err := c.Invoke(func(server *ServerService, api *ApiService, config *ConfigService, lookup *LookupService) {
		api.SetServerService(server)
		config.SetServerService(server)
		
		// Initialize lookup data using repository directly
		lookup.cache.Set("tracks", lookup.repository.GetTracks(context.Background()))
		lookup.cache.Set("cars", lookup.repository.GetCarModels(context.Background()))
		lookup.cache.Set("drivers", lookup.repository.GetDriverCategories(context.Background()))
		lookup.cache.Set("cups", lookup.repository.GetCupCategories(context.Background()))
		lookup.cache.Set("sessions", lookup.repository.GetSessionTypes(context.Background()))
	})
	if err != nil {
		log.Panic("unable to initialize services:", err)
	}
}
