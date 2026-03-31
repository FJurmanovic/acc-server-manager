package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/configs"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

type SystemController struct {
	systemService *service.SystemService
}

type SystemControllerParams struct {
	dig.In
	RouteGroups    *common.RouteGroups
	AuthMiddleware *middleware.AuthMiddleware
	SystemService  *service.SystemService `optional:"true"`
}

// NewSystemController initializes SystemController.
func NewSystemController(p SystemControllerParams) *SystemController {
	ac := &SystemController{systemService: p.SystemService}

	apiGroup := p.RouteGroups.System
	apiGroup.Get("/health", ac.getHealth)

	if p.SystemService != nil {
		apiGroup.Post("/migrate-image", p.AuthMiddleware.Authenticate, ac.migrateImage)
	}

	return ac
}

// getHealth returns the current version string.
//
//	@Summary		Return service health / version
//	@Description	Return service health / version
//	@Tags			system
//	@Success		200	{string}	string
//	@Router			/system/health [get]
func (ac *SystemController) getHealth(c *fiber.Ctx) error {
	return c.SendString(configs.Version)
}

// migrateImage godoc
//
//	@Summary		Migrate all Docker server containers to current ACC image
//	@Description	Pulls the current ACC_IMAGE, removes old containers, optionally stops running servers
//	@Tags			system
//	@Accept			json
//	@Produce		json
//	@Param			body	body		model.MigrateImageRequest	true	"Migration options"
//	@Success		200		{object}	model.MigrationResult
//	@Router			/system/migrate-image [post]
func (ac *SystemController) migrateImage(c *fiber.Ctx) error {
	var req model.MigrateImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	result, err := ac.systemService.MigrateImage(c.Context(), req.StopRunning)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}
