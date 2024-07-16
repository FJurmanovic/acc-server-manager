package service

import (
	"acc-server-manager/local/repository"

	"go.uber.org/dig"
)

// InitializeServices
// Initializes Dependency Injection modules for services
//
//	Args:
//		*dig.Container: Dig Container
func InitializeServices(c *dig.Container) {
	repository.InitializeRepositories(c)

	c.Provide(NewApiService)
}
