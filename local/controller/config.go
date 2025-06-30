package controller

import (
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"
	"acc-server-manager/local/utl/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ConfigController struct {
	service      *service.ConfigService
	apiService   *service.ApiService
	errorHandler *error_handler.ControllerErrorHandler
}

// NewConfigController
// Initializes ConfigController.
//
//	Args:
//		*services.ConfigService: Config service
//		*Fiber.RouterGroup: Fiber Router Group
//	Returns:
//		*ConfigController: Controller for "Config" interactions
func NewConfigController(as *service.ConfigService, routeGroups *common.RouteGroups, as2 *service.ApiService) *ConfigController {
	ac := &ConfigController{
		service:      as,
		apiService:   as2,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	routeGroups.Config.Put("/:file", ac.UpdateConfig)
	routeGroups.Config.Get("/:file", ac.GetConfig)
	routeGroups.Config.Get("/", ac.GetConfigs)

	return ac
}

// updateConfig returns Config
//
//	@Summary		Update config
//	@Description	Updates config
//	@Param			id path number true "required"
//	@Param			file path string true "required"
//	@Param			content body string true "required"
//	@Tags			Config
//	@Success		200	{array}	string
//	@Router			/v1/server/{id}/config/{file} [put]
func (ac *ConfigController) UpdateConfig(c *fiber.Ctx) error {
	restart := c.QueryBool("restart")
	serverID := c.Params("id")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		return ac.errorHandler.HandleUUIDError(c, "server ID")
	}

	c.Locals("serverId", serverID)

	var config map[string]interface{}
	if err := c.BodyParser(&config); err != nil {
		return ac.errorHandler.HandleParsingError(c, err)
	}

	ConfigModel, err := ac.service.UpdateConfig(c, &config)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	logging.Info("restart: %v", restart)
	if restart {
		_, err := ac.apiService.ApiRestartServer(c)
		if err != nil {
			logging.ErrorWithContext("CONFIG_RESTART", "Failed to restart server after config update: %v", err)
		}
	}

	return c.JSON(ConfigModel)
}

// getConfig returns Config
//
//	@Summary		Return Config file
//	@Description	Returns Config file
//	@Param			id path number true "required"
//	@Param			file path string true "required"
//	@Tags			Config
//	@Success		200	{array}	string
//	@Router			/v1/server/{id}/config/{file} [get]
func (ac *ConfigController) GetConfig(c *fiber.Ctx) error {
	Model, err := ac.service.GetConfig(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(Model)
}

// getConfigs returns Config
//
//	@Summary		Return Configs
//	@Description	Return Config files
//	@Param			id path number true "required"
//	@Tags			Config
//	@Success		200	{array}	string
//	@Router			/v1/server/{id}/config [get]
func (ac *ConfigController) GetConfigs(c *fiber.Ctx) error {
	Model, err := ac.service.GetConfigs(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(Model)
}
