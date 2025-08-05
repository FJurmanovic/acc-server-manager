package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"

	"github.com/gofiber/fiber/v2"
)

type ServiceControlController struct {
	service      *service.ServiceControlService
	errorHandler *error_handler.ControllerErrorHandler
}

// NewServiceControlController
// Initializes ServiceControlController.
//
//	Args:
//		*services.ServiceControlService: Service control service
//		*Fiber.RouterGroup: Fiber Router Group
//	Returns:
//		*ServiceControlController: Controller for service control interactions
func NewServiceControlController(as *service.ServiceControlService, routeGroups *common.RouteGroups, auth *middleware.AuthMiddleware) *ServiceControlController {
	ac := &ServiceControlController{
		service:      as,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	serviceRoutes := routeGroups.Server.Group("/:id/service")
	serviceRoutes.Get("/:service", ac.getStatus)
	serviceRoutes.Post("/start", ac.startServer)
	serviceRoutes.Post("/stop", ac.stopServer)
	serviceRoutes.Post("/restart", ac.restartServer)

	return ac
}

// getStatus returns service status
//
//	@Summary		Get service status
//	@Description	Get the current status of a Windows service
//	@Tags			Service Control
//	@Accept			json
//	@Produce		json
//	@Param			service path string true "Service name"
//	@Success		200	{object} object{status=string,state=string} "Service status information"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid service name"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		404	{object} error_handler.ErrorResponse "Service not found"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/v1/server/{id}/service/{service} [get]
func (ac *ServiceControlController) getStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	c.Locals("serverId", id)
	apiModel, err := ac.service.GetStatus(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.SendString(string(apiModel))
}

// startServer starts service
//
//	@Summary		Start a Windows service
//	@Description	Start a stopped Windows service for an ACC server
//	@Tags			Service Control
//	@Accept			json
//	@Produce		json
//	@Param			service body object{name=string} true "Service name to start"
//	@Success		200	{object} object{message=string} "Service started successfully"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid request body"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Service not found"
//	@Failure		409	{object} error_handler.ErrorResponse "Service already running"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/v1/server/{id}/service/start [post]
func (ac *ServiceControlController) startServer(c *fiber.Ctx) error {
	id := c.Params("id")
	c.Locals("serverId", id)
	apiModel, err := ac.service.ServiceControlStartServer(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.SendString(apiModel)
}

// stopServer stops service
//
//	@Summary		Stop a Windows service
//	@Description	Stop a running Windows service for an ACC server
//	@Tags			Service Control
//	@Accept			json
//	@Produce		json
//	@Param			service body object{name=string} true "Service name to stop"
//	@Success		200	{object} object{message=string} "Service stopped successfully"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid request body"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Service not found"
//	@Failure		409	{object} error_handler.ErrorResponse "Service already stopped"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/v1/server/{id}/service/stop [post]
func (ac *ServiceControlController) stopServer(c *fiber.Ctx) error {
	id := c.Params("id")
	c.Locals("serverId", id)
	apiModel, err := ac.service.ServiceControlStopServer(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.SendString(apiModel)
}

// restartServer restarts service
//
//	@Summary		Restart a Windows service
//	@Description	Stop and start a Windows service for an ACC server
//	@Tags			Service Control
//	@Accept			json
//	@Produce		json
//	@Param			service body object{name=string} true "Service name to restart"
//	@Success		200	{object} object{message=string} "Service restarted successfully"
//	@Failure		400	{object} error_handler.ErrorResponse "Invalid request body"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		403	{object} error_handler.ErrorResponse "Insufficient permissions"
//	@Failure		404	{object} error_handler.ErrorResponse "Service not found"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/v1/server/{id}/service/restart [post]
func (ac *ServiceControlController) restartServer(c *fiber.Ctx) error {
	id := c.Params("id")
	c.Locals("serverId", id)
	apiModel, err := ac.service.ServiceControlRestartServer(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.SendString(apiModel)
}

type Service struct {
	Name     string `json:"name" xml:"name" form:"name"`
	ServerId string `json:"serverId" xml:"serverId" form:"serverId"`
}
