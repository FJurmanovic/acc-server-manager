package api

import (
	"acc-server-manager/local/controller"
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/configs"
	"acc-server-manager/local/utl/logging"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

// Routes
// Initializes web api controllers and its corresponding routes.
//
//	Args:
//		*fiber.App: Fiber Application
func Init(di *dig.Container, app *fiber.App) {

	// Protected routes
	groups := app.Group(configs.Prefix)

	serverIdGroup := groups.Group("/server/:id")
	routeGroups := &common.RouteGroups{
		Api:          groups.Group("/api"),
		Auth:         groups.Group("/auth"),
		Server:       groups.Group("/server"),
		Config:       serverIdGroup.Group("/config"),
		Lookup:       groups.Group("/lookup"),
		StateHistory: serverIdGroup.Group("/state-history"),
		Membership:   groups.Group("/membership"),
		System:       groups.Group("/system"),
	}

	accessKeyMiddleware := middleware.NewAccessKeyMiddleware()
	routeGroups.Api.Use(accessKeyMiddleware.Authenticate)

	err := di.Provide(func() *common.RouteGroups {
		return routeGroups
	})
	if err != nil {
		logging.Panic("unable to bind routes")
	}

	controller.InitializeControllers(di)
}
