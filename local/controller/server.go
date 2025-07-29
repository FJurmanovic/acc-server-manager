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

// GetAll returns Servers
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
