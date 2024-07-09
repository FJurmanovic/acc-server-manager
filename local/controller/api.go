package controller

import (
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"

	"github.com/gofiber/fiber/v2"
)

type ApiController struct {
	service *service.ApiService
}

/*
NewApiController

Initializes ApiController.

	Args:
		*services.ApiService: API service
		*Fiber.RouterGroup: Fiber Router Group
	Returns:
		*ApiController: Controller for "api" interactions
*/
func NewApiController(as *service.ApiService, routeGroups *common.RouteGroups) *ApiController {
	ac := &ApiController{
		service: as,
	}

	routeGroups.Api.Get("", ac.getFirst)

	return ac
}

/*
getFirst
	Args:
		*fiber.Ctx: Fiber Application Context
*/
// ROUTE (GET /api).
func (ac *ApiController) getFirst(c *fiber.Ctx) error {
	apiModel := ac.service.GetFirst(c)
	return c.SendString(apiModel)
}
