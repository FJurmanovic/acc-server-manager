package controller

import (
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"

	"github.com/gofiber/fiber/v2"
)

type ApiController struct {
	service *service.ApiService
}

// NewApiController
// Initializes ApiController.
//
//	Args:
//		*services.ApiService: API service
//		*Fiber.RouterGroup: Fiber Router Group
//	Returns:
//		*ApiController: Controller for "api" interactions
func NewApiController(as *service.ApiService, routeGroups *common.RouteGroups) *ApiController {
	ac := &ApiController{
		service: as,
	}

	routeGroups.Api.Get("/", ac.getFirst)
	routeGroups.Api.Post("/", ac.startServer)

	return ac
}

// getFirst returns API
//
//	@Summary		Return API
//	@Description	Return API
//	@Tags			api
//	@Success		200	{array}		string
//	@Router			/v1/api [get]
func (ac *ApiController) getFirst(c *fiber.Ctx) error {
	apiModel := ac.service.GetFirst(c)
	return c.SendString(apiModel.Api)
}

// startServer returns API
//
//	@Summary		Return API
//	@Description	Return API
//	@Tags			api
//	@Success		200	{array}		string
//	@Router			/v1/api [post]
func (ac *ApiController) startServer(c *fiber.Ctx) error {
	c.Locals("service", "ACC-Server")
	apiModel, err := ac.service.ApiStartServer(c)
	if err != nil {
		return c.SendStatus(400)
	}
	return c.SendString(apiModel)
}
