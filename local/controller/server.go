package controller

import (
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"

	"github.com/gofiber/fiber/v2"
)

type ServerController struct {
	service *service.ServerService
}

// NewServerController
// Initializes ServerController.
//
//	Args:
//		*services.ServerService: Server service
//		*Fiber.RouterGroup: Fiber Router Group
//	Returns:
//		*ServerController: Controller for "Server" interactions
func NewServerController(as *service.ServerService, routeGroups *common.RouteGroups) *ServerController {
	ac := &ServerController{
		service: as,
	}

	routeGroups.Server.Get("/", ac.getAll)

	return ac
}

// getAll returns Servers
//
//	@Summary		Return Servers
//	@Description	Return Servers
//	@Tags			Server
//	@Success		200	{array}	string
//	@Router			/v1/server [get]
func (ac *ServerController) getAll(c *fiber.Ctx) error {
	ServerModel := ac.service.GetAll(c)
	return c.JSON(ServerModel)
}
