package api

import (
	"acc-server-manager/local/controller"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/configs"
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

	routeGroups := &common.RouteGroups{
		Api: groups.Group("/api"),
	}

	routeGroups.Api.Use(basicAuthConfig)

	err := di.Provide(func() *common.RouteGroups {
		return routeGroups
	})
	if err != nil {
		panic("unable to bind routes")
	}

	controller.InitializeControllers(di)
}
