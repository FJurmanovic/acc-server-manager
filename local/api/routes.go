package api

import (
	"acc-server-manager/local/controller"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/configs"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

/*
Routes

Initializes web api controllers and its corresponding routes.

	Args:
		*fiber.App: Fiber Application
*/
func Routes(app *fiber.App) {
	c := dig.New()
	groups := app.Group(configs.Prefix)

	apiGroup := groups.Group("api")
	routeGroups := &common.RouteGroups{
		Api: apiGroup,
	}

	c.Provide(func() *common.RouteGroups {
		return routeGroups
	})

	controller.InitializeControllers(c)
}
