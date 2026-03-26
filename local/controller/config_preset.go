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

type ConfigPresetController struct {
	service      *service.ConfigPresetService
	apiService   *service.ServiceControlService
	errorHandler *error_handler.ControllerErrorHandler
}

// NewConfigPresetController initializes ConfigPresetController.
func NewConfigPresetController(s *service.ConfigPresetService, apiService *service.ServiceControlService, routeGroups *common.RouteGroups, auth *middleware.AuthMiddleware) *ConfigPresetController {
	c := &ConfigPresetController{
		service:      s,
		apiService:   apiService,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	presetRoutes := routeGroups.Presets
	presetRoutes.Use(auth.Authenticate)
	presetRoutes.Get("/", auth.HasPermission(model.PresetView), c.ListPresets)
	presetRoutes.Post("/", auth.HasPermission(model.PresetCreate), c.CreatePreset)
	presetRoutes.Get("/:presetId", auth.HasPermission(model.PresetView), c.GetPreset)
	presetRoutes.Put("/:presetId", auth.HasPermission(model.PresetUpdate), c.UpdatePreset)
	presetRoutes.Delete("/:presetId", auth.HasPermission(model.PresetDelete), c.DeletePreset)

	configRoutes := routeGroups.Config
	configRoutes.Post("/presets/:presetId/apply", auth.HasPermission(model.ConfigUpdate), c.ApplyPreset)

	return c
}

// ListPresets returns all config presets.
//
//	@Summary		List config presets
//	@Description	Get a list of all saved configuration presets
//	@Tags			Config Presets
//	@Accept			json
//	@Produce		json
//	@Param			name query string false "Filter by preset name (partial match)"
//	@Success		200	{array} model.ConfigPreset "List of presets"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/presets [get]
func (c *ConfigPresetController) ListPresets(ctx *fiber.Ctx) error {
	var filter model.ConfigPresetFilter
	if err := common.ParseQueryFilter(ctx, &filter); err != nil {
		return c.errorHandler.HandleParsingError(ctx, err)
	}

	presets, err := c.service.ListPresets(ctx.UserContext(), &filter)
	if err != nil {
		return c.errorHandler.HandleServiceError(ctx, err)
	}
	return ctx.JSON(presets)
}

// GetPreset returns a single config preset by ID.
//
//	@Summary		Get config preset
//	@Description	Retrieve a single configuration preset by ID
//	@Tags			Config Presets
//	@Accept			json
//	@Produce		json
//	@Param			presetId path string true "Preset ID (UUID format)"
//	@Success		200	{object} model.ConfigPreset "Preset"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid preset ID"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Preset not found"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/presets/{presetId} [get]
func (c *ConfigPresetController) GetPreset(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("presetId"))
	if err != nil {
		return c.errorHandler.HandleUUIDError(ctx, "preset ID")
	}

	preset, err := c.service.GetPreset(ctx.UserContext(), id)
	if err != nil {
		return c.errorHandler.HandleServiceError(ctx, err)
	}
	return ctx.JSON(preset)
}

// CreatePreset creates a new config preset.
//
//	@Summary		Create config preset
//	@Description	Save a new configuration preset with one or more ACC config sections
//	@Tags			Config Presets
//	@Accept			json
//	@Produce		json
//	@Param			content body model.ConfigPresetCreateRequest true "Preset data"
//	@Success		201	{object} model.ConfigPreset "Created preset"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid request"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/presets [post]
func (c *ConfigPresetController) CreatePreset(ctx *fiber.Ctx) error {
	var req model.ConfigPresetCreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return c.errorHandler.HandleParsingError(ctx, err)
	}

	preset, err := c.service.CreatePreset(ctx.UserContext(), &req)
	if err != nil {
		return c.errorHandler.HandleServiceError(ctx, err)
	}
	return ctx.Status(fiber.StatusCreated).JSON(preset)
}

// UpdatePreset updates an existing config preset.
//
//	@Summary		Update config preset
//	@Description	Update name, description, or any config sections of an existing preset
//	@Tags			Config Presets
//	@Accept			json
//	@Produce		json
//	@Param			presetId path string true "Preset ID (UUID format)"
//	@Param			content body model.ConfigPresetUpdateRequest true "Fields to update (omit to keep existing)"
//	@Success		200	{object} model.ConfigPreset "Updated preset"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid request or preset ID"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Preset not found"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/presets/{presetId} [put]
func (c *ConfigPresetController) UpdatePreset(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("presetId"))
	if err != nil {
		return c.errorHandler.HandleUUIDError(ctx, "preset ID")
	}

	var req model.ConfigPresetUpdateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return c.errorHandler.HandleParsingError(ctx, err)
	}

	preset, err := c.service.UpdatePreset(ctx.UserContext(), id, &req)
	if err != nil {
		return c.errorHandler.HandleServiceError(ctx, err)
	}
	return ctx.JSON(preset)
}

// DeletePreset deletes a config preset.
//
//	@Summary		Delete config preset
//	@Description	Permanently delete a configuration preset
//	@Tags			Config Presets
//	@Accept			json
//	@Produce		json
//	@Param			presetId path string true "Preset ID (UUID format)"
//	@Success		204	"No Content"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid preset ID"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/presets/{presetId} [delete]
func (c *ConfigPresetController) DeletePreset(ctx *fiber.Ctx) error {
	id, err := uuid.Parse(ctx.Params("presetId"))
	if err != nil {
		return c.errorHandler.HandleUUIDError(ctx, "preset ID")
	}

	if err := c.service.DeletePreset(ctx.UserContext(), id); err != nil {
		return c.errorHandler.HandleServiceError(ctx, err)
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// ApplyPreset applies a saved preset to a server's config files.
//
//	@Summary		Apply config preset to server
//	@Description	Apply a saved configuration preset to an ACC server, writing only the sections stored in the preset
//	@Tags			Server Configuration
//	@Accept			json
//	@Produce		json
//	@Param			id path string true "Server ID (UUID format)"
//	@Param			presetId path string true "Preset ID (UUID format)"
//	@Param			restart query bool false "Restart the server after applying the preset"
//	@Success		200	{array} model.Config "List of updated config audit records"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid server or preset ID"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Server or preset not found"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/server/{id}/config/presets/{presetId}/apply [post]
func (c *ConfigPresetController) ApplyPreset(ctx *fiber.Ctx) error {
	restart := ctx.QueryBool("restart")
	serverID := ctx.Params("id")

	if _, err := uuid.Parse(serverID); err != nil {
		return c.errorHandler.HandleUUIDError(ctx, "server ID")
	}

	ctx.Locals("serverId", serverID)

	presetID, err := uuid.Parse(ctx.Params("presetId"))
	if err != nil {
		return c.errorHandler.HandleUUIDError(ctx, "preset ID")
	}

	actorID, _ := ctx.Locals("userID").(string)
	actorUsername, _ := ctx.Locals("actorUsername").(string)
	if actorUsername == "" {
		actorUsername = "system"
	}

	configs, err := c.service.ApplyPreset(ctx.UserContext(), serverID, presetID, actorID, actorUsername)
	if err != nil {
		return c.errorHandler.HandleServiceError(ctx, err)
	}

	if restart {
		if _, err := c.apiService.ServiceControlRestartServer(ctx); err != nil {
			logging.ErrorWithContext("PRESET_APPLY_RESTART", "Failed to restart server after preset apply: %v", err)
		}
	}

	return ctx.JSON(configs)
}
