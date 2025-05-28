package controller

import (
	"acc-server-manager/local/model"
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
	routeGroups.Server.Get("/:id", ac.getById)

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
	var filter model.ServerFilter	
	if err := common.ParseQueryFilter(c, &filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	ServerModel, err := ac.service.GetAll(c, &filter)
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}
	return c.JSON(ServerModel)
}

// getById returns Servers
//
//	@Summary		Return Servers
//	@Description	Return Servers
//	@Tags			Server
//	@Success		200	{array}	string
//	@Router			/v1/server [get]
func (ac *ServerController) getById(c *fiber.Ctx) error {
	serverID, _ := c.ParamsInt("id")
	ServerModel, err := ac.service.GetById(c, serverID)
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}
	return c.JSON(ServerModel)
}
