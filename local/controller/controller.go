package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/logging"

	"go.uber.org/dig"
)

// InitializeControllers
// Initializes Dependency Injection modules and registers controllers
//
//	Args:
//		*dig.Container: Dig Container
func InitializeControllers(c *dig.Container) {
	service.InitializeServices(c)

	if err := c.Provide(middleware.NewAuthMiddleware); err != nil {
		logging.Panic("unable to initialize auth middleware")
	}

	err := c.Invoke(NewSystemController)
	if err != nil {
		logging.Panic("unable to initialize system controller")
	}

	err = c.Invoke(NewServiceControlController)
	if err != nil {
		logging.Panic("unable to initialize service control controller")
	}

	err = c.Invoke(NewConfigController)
	if err != nil {
		logging.Panic("unable to initialize config controller")
	}

	err = c.Invoke(NewServerController)
	if err != nil {
		logging.Panic("unable to initialize server controller")
	}

	err = c.Invoke(NewLookupController)
	if err != nil {
		logging.Panic("unable to initialize lookup controller")
	}

	err = c.Invoke(NewStateHistoryController)
	if err != nil {
		logging.Panic("unable to initialize stateHistory controller")
	}

	err = c.Invoke(NewMembershipController)
	if err != nil {
		logging.Panic("unable to initialize membership controller")
	}

	err = c.Invoke(NewWebSocketController)
	if err != nil {
		logging.Panic("unable to initialize websocket controller")
	}
}
