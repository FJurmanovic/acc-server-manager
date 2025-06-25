package api

import (
	"acc-server-manager/local/controller"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/configs"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

// Routes
// Initializes web api controllers and its corresponding routes.
//
//	Args:
//		*fiber.App: Fiber Application
func Init(di *dig.Container, app *fiber.App) {
	// Setup initial data for membership
	di.Invoke(func(membershipService *service.MembershipService) {
		if err := membershipService.SetupInitialData(context.Background()); err != nil {
			logging.Panic(fmt.Sprintf("failed to setup initial data: %v", err))
		}
	})

	// Protected routes
	groups := app.Group(configs.Prefix)



	serverIdGroup := groups.Group("/server/:id")
	routeGroups := &common.RouteGroups{
		Api:          groups.Group("/api"),
		Auth:         app.Group("/auth"),
		Server:       groups.Group("/server"),
		Config:       serverIdGroup.Group("/config"),
		Lookup:       groups.Group("/lookup"),
		StateHistory: serverIdGroup.Group("/state-history"),
	}

	err := di.Provide(func() *common.RouteGroups {
		return routeGroups
	})
	if err != nil {
		logging.Panic("unable to bind routes")
	}
	err = di.Provide(func() *dig.Container {
		return di
	})
	if err != nil {
		logging.Panic("unable to bind dig")
	}

	controller.InitializeControllers(di)
}

