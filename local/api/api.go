package api

import (
	"acc-server-manager/local/controller"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/configs"
	"acc-server-manager/local/utl/logging"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"go.uber.org/dig"
)

// Routes
// Initializes web api controllers and its corresponding routes.
//
//	Args:
//		*fiber.App: Fiber Application
func Init(di *dig.Container, app *fiber.App) {
	groups := app.Group(configs.Prefix)

	basicAuthConfig := basicauth.New(basicauth.Config{
		Users: map[string]string{
			"admin": os.Getenv("PASSWORD"),
		},
	})

	serverIdGroup := groups.Group("/server/:id")
	routeGroups := &common.RouteGroups{
		Api:    groups.Group("/api"),
		Server: groups.Group("/server"),
		Config: serverIdGroup.Group("/config"),
		Lookup: groups.Group("/lookup"),
		StateHistory: serverIdGroup.Group("/state-history"),
	}

	groups.Use(basicAuthConfig)

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

