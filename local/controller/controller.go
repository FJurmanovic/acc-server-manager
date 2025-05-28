package controller

import (
	"acc-server-manager/local/service"
	"log"

	"go.uber.org/dig"
)

// InitializeControllers
// Initializes Dependency Injection modules and registers controllers
//
//	Args:
//		*dig.Container: Dig Container
func InitializeControllers(c *dig.Container) {
	service.InitializeServices(c)

	err := c.Invoke(NewApiController)
	if err != nil {
		log.Panic("unable to initialize api controller")
	}

	err = c.Invoke(NewConfigController)
	if err != nil {
		log.Panic("unable to initialize config controller")
	}

	err = c.Invoke(NewServerController)
	if err != nil {
		log.Panic("unable to initialize server controller")
	}

	err = c.Invoke(NewLookupController)
	if err != nil {
		log.Panic("unable to initialize lookup controller")
	}

	err = c.Invoke(NewStateHistoryController)
	if err != nil {
		log.Panic("unable to initialize stateHistory controller")
	}
}
