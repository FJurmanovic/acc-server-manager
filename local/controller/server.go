package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ServerController struct {
	service      *service.ServerService
	errorHandler *error_handler.ControllerErrorHandler
}

// NewServerController initializes ServerController.
func NewServerController(ss *service.ServerService, routeGroups *common.RouteGroups, auth *middleware.AuthMiddleware) *ServerController {
	ac := &ServerController{
		service:      ss,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	serverRoutes := routeGroups.Server
	serverRoutes.Use(auth.Authenticate)

	serverRoutes.Get("/", auth.HasPermission(model.ServerView), ac.GetAll)
	serverRoutes.Get("/:id", auth.HasPermission(model.ServerView), ac.GetById)
	serverRoutes.Post("/", auth.HasPermission(model.ServerCreate), ac.CreateServer)
	serverRoutes.Put("/:id", auth.HasPermission(model.ServerUpdate), ac.UpdateServer)
	serverRoutes.Delete("/:id", auth.HasPermission(model.ServerDelete), ac.DeleteServer)

	apiServerRoutes := routeGroups.Api.Group("/server")
	apiServerRoutes.Get("/", auth.HasPermission(model.ServerView), ac.GetAllApi)

	return ac
}

// GetAllApi returns all servers in API format
// @Summary List all servers (API format)
// @Description Get a list of all ACC servers with filtering options
// @Tags Server
// @Accept json
// @Produce json
// @Param filter query model.ServerFilter false "Filter options"
// @Success 200 {array} model.ServerAPI "List of servers"
// @Failure 400 {object} error_handler.ErrorResponse "Invalid filter parameters"
// @Failure 401 {object} error_handler.ErrorResponse "Unauthorized"
// @Failure 500 {object} error_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /server [get]
func (ac *ServerController) GetAllApi(c *fiber.Ctx) error {
	var filter model.ServerFilter
	if err := common.ParseQueryFilter(c, &filter); err != nil {
		return ac.errorHandler.HandleValidationError(c, err, "query_filter")
	}
	ServerModel, err := ac.service.GetAll(c, &filter)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	var apiServers []model.ServerAPI
	for _, server := range *ServerModel {
		apiServers = append(apiServers, *server.ToServerAPI())
	}
	return c.JSON(apiServers)
}

// GetAll returns all servers
// @Summary List all servers
// @Description Get a list of all ACC servers with detailed information
// @Tags Server
// @Accept json
// @Produce json
// @Param filter query model.ServerFilter false "Filter options"
// @Success 200 {array} model.Server "List of servers with full details"
// @Failure 400 {object} error_handler.ErrorResponse "Invalid filter parameters"
// @Failure 401 {object} error_handler.ErrorResponse "Unauthorized"
// @Failure 500 {object} error_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /v1/server [get]
func (ac *ServerController) GetAll(c *fiber.Ctx) error {
	var filter model.ServerFilter
	if err := common.ParseQueryFilter(c, &filter); err != nil {
		return ac.errorHandler.HandleValidationError(c, err, "query_filter")
	}
	ServerModel, err := ac.service.GetAll(c, &filter)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(ServerModel)
}

// GetById returns a single server by its ID
// @Summary Get server by ID
// @Description Get detailed information about a specific ACC server
// @Tags Server
// @Accept json
// @Produce json
// @Param id path string true "Server ID (UUID format)"
// @Success 200 {object} model.Server "Server details"
// @Failure 400 {object} error_handler.ErrorResponse "Invalid server ID format"
// @Failure 401 {object} error_handler.ErrorResponse "Unauthorized"
// @Failure 404 {object} error_handler.ErrorResponse "Server not found"
// @Failure 500 {object} error_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /v1/server/{id} [get]
func (ac *ServerController) GetById(c *fiber.Ctx) error {
	serverIDStr := c.Params("id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		return ac.errorHandler.HandleUUIDError(c, "server ID")
	}

	ServerModel, err := ac.service.GetById(c, serverID)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(ServerModel)
}

// CreateServer creates a new server
// @Summary Create a new ACC server
// @Description Create a new ACC server instance with the provided configuration
// @Tags Server
// @Accept json
// @Produce json
// @Param server body model.Server true "Server configuration"
// @Success 200 {object} object "Created server details"
// @Failure 400 {object} error_handler.ErrorResponse "Invalid server data"
// @Failure 401 {object} error_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} error_handler.ErrorResponse "Insufficient permissions"
// @Failure 500 {object} error_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /v1/server [post]
func (ac *ServerController) CreateServer(c *fiber.Ctx) error {
	server := new(model.Server)
	if err := c.BodyParser(server); err != nil {
		return ac.errorHandler.HandleParsingError(c, err)
	}
	ac.service.GenerateServerPath(server)
	if err := ac.service.CreateServer(c, server); err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(server)
}

// UpdateServer updates an existing server
// @Summary Update an ACC server
// @Description Update configuration for an existing ACC server
// @Tags Server
// @Accept json
// @Produce json
// @Param id path string true "Server ID (UUID format)"
// @Param server body model.Server true "Updated server configuration"
// @Success 200 {object} object "Updated server details"
// @Failure 400 {object} error_handler.ErrorResponse "Invalid server data or ID"
// @Failure 401 {object} error_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} error_handler.ErrorResponse "Insufficient permissions"
// @Failure 404 {object} error_handler.ErrorResponse "Server not found"
// @Failure 500 {object} error_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /v1/server/{id} [put]
func (ac *ServerController) UpdateServer(c *fiber.Ctx) error {
	serverIDStr := c.Params("id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		return ac.errorHandler.HandleUUIDError(c, "server ID")
	}

	server := new(model.Server)
	if err := c.BodyParser(server); err != nil {
		return ac.errorHandler.HandleParsingError(c, err)
	}
	server.ID = serverID

	if err := ac.service.UpdateServer(c, server); err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(server)
}

// DeleteServer deletes a server
func (ac *ServerController) DeleteServer(c *fiber.Ctx) error {
	serverIDStr := c.Params("id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		return ac.errorHandler.HandleUUIDError(c, "server ID")
	}

	if err := ac.service.DeleteServer(c, serverID); err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}

	return c.SendStatus(204)
}
