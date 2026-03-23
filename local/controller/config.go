package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"
	"acc-server-manager/local/utl/logging"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ConfigController struct {
	service      *service.ConfigService
	apiService   *service.ServiceControlService
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
func NewConfigController(as *service.ConfigService, routeGroups *common.RouteGroups, as2 *service.ServiceControlService, auth *middleware.AuthMiddleware) *ConfigController {
	ac := &ConfigController{
		service:      as,
		apiService:   as2,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	configGroup := routeGroups.Config
	configGroup.Use(auth.Authenticate)
	configGroup.Patch("/", ac.UpdateAllConfigs)
	configGroup.Put("/:file", ac.UpdateConfig)
	configGroup.Get("/:file", ac.GetConfig)
	configGroup.Get("/", ac.GetConfigs)

	return ac
}

// updateAllConfigs updates all configuration files
//
//	@Summary		Bulk update all server configuration files
//	@Description	Update multiple ACC server configuration files in one request. Null or omitted sections are skipped.
//	@Tags			Server Configuration
//	@Accept			json
//	@Produce		json
//	@Param			id path string true "Server ID (UUID format)"
//	@Param			restart query bool false "Restart the server after update"
//	@Param			override query bool false "Override instead of merging with existing config"
//	@Param			content body model.ConfigurationsUpdateRequest true "Configuration sections to update (null sections are skipped)"
//	@Success		200	{array} model.Config "List of updated config audit records"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid request or JSON format"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Server not found"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/server/{id}/config [patch]
func (ac *ConfigController) UpdateAllConfigs(c *fiber.Ctx) error {
	restart := c.QueryBool("restart")
	serverID := c.Params("id")

	if _, err := uuid.Parse(serverID); err != nil {
		return ac.errorHandler.HandleUUIDError(c, "server ID")
	}

	c.Locals("serverId", serverID)

	var req model.ConfigurationsUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return ac.errorHandler.HandleParsingError(c, err)
	}

	configs, err := ac.service.UpdateAllConfigs(c, &req)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}

	if restart {
		if _, err := ac.apiService.ServiceControlRestartServer(c); err != nil {
			logging.ErrorWithContext("CONFIG_RESTART", "Failed to restart server after bulk config update: %v", err)
		}
	}

	return c.JSON(configs)
}

// updateConfig returns Config
//
//	@Summary		Update server configuration file
//	@Description	Update a specific configuration file for an ACC server
//	@Tags			Server Configuration
//	@Accept			json
//	@Produce		json
//	@Param			id path string true "Server ID (UUID format)"
//	@Param			file path string true "Config file name (e.g., configuration.json, settings.json, event.json)"
//	@Param			content body object true "Configuration file content as JSON"
//	@Success		200	{object} object{message=string} "Update successful"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid request or JSON format"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Server or config file not found"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/server/{id}/config/{file} [put]
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
		_, err := ac.apiService.ServiceControlRestartServer(c)
		if err != nil {
			logging.ErrorWithContext("CONFIG_RESTART", "Failed to restart server after config update: %v", err)
		}
	}

	return c.JSON(ConfigModel)
}

// getConfig returns Config
//
//	@Summary		Get server configuration file
//	@Description	Retrieve a specific configuration file for an ACC server
//	@Tags			Server Configuration
//	@Accept			json
//	@Produce		json
//	@Param			id path string true "Server ID (UUID format)"
//	@Param			file path string true "Config file name (e.g., configuration.json, settings.json, event.json)"
//	@Success		200	{object} object "Configuration file content as JSON"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid server ID"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Server or config file not found"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/server/{id}/config/{file} [get]
func (ac *ConfigController) GetConfig(c *fiber.Ctx) error {
	Model, err := ac.service.GetConfig(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(Model)
}

// getConfigs returns Config
//
//	@Summary		List available configuration files
//	@Description	Get a list of all available configuration files for an ACC server
//	@Tags			Server Configuration
//	@Accept			json
//	@Produce		json
//	@Param			id path string true "Server ID (UUID format)"
//	@Success		200	{array} string "List of available configuration files"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid server ID"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Server not found"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/server/{id}/config [get]
func (ac *ConfigController) GetConfigs(c *fiber.Ctx) error {
	Model, err := ac.service.GetConfigs(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(Model)
}
