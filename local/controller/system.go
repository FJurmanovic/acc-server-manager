package controller

import (
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/configs"

	"github.com/gofiber/fiber/v2"
)

type SystemController struct {
}

// NewSystemController
// Initializes SystemController.
//
//	Args:
//		*services.SystemService: Service control service
//		*Fiber.RouterGroup: Fiber Router Group
//	Returns:
//		*SystemController: Controller for service control interactions
func NewSystemController(routeGroups *common.RouteGroups) *SystemController {
	ac := &SystemController{}

	apiGroup := routeGroups.System
	apiGroup.Get("/health", ac.getFirst)

	return ac
}

// getFirst returns service control status
//
//	@Summary		Return service control status
//	@Description	Return service control status
//	@Tags			service-control
//	@Success		200	{array}	string
//	@Router			/v1/service-control [get]
func (ac *SystemController) getFirst(c *fiber.Ctx) error {
	return c.SendString(configs.Version)
}
