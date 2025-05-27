package service

import (
	"acc-server-manager/local/repository"
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

	err := c.Invoke(func(server *ServerService, api *ApiService, config *ConfigService) {
		api.SetServerService(server)
		config.SetServerService(server)
	})
	if err != nil {
		log.Panic("unable to initialize server service in api service")
	}
}
