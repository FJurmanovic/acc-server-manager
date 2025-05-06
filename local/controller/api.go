package controller

import (
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"log"
	"strings"

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
	routeGroups.Api.Get("/:service", ac.getStatus)
	routeGroups.Api.Post("/start", ac.startServer)
	routeGroups.Api.Post("/stop", ac.stopServer)
	routeGroups.Api.Post("/restart", ac.restartServer)

	return ac
}

// getFirst returns API
//
//	@Summary		Return API
//	@Description	Return API
//	@Tags			api
//	@Success		200	{array}	string
//	@Router			/v1/api [get]
func (ac *ApiController) getFirst(c *fiber.Ctx) error {
	apiModel := ac.service.GetFirst(c)
	return c.SendString(apiModel.Api)
}

// getStatus returns service status
//
//	@Summary		Return service status
//	@Description	Returns service status
//	@Param			service path string true "required"
//	@Tags			api
//	@Success		200	{array}	string
//	@Router			/v1/api/{service} [get]
func (ac *ApiController) getStatus(c *fiber.Ctx) error {
	service := c.Params("service")
	if service == "" {
		serverId, err := c.ParamsInt("service")
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		c.Locals("serverId", serverId)
	} else {
		c.Locals("service", service)
	}
	apiModel, err := ac.service.GetStatus(c)
	if err != nil {
		return c.Status(400).SendString(strings.ReplaceAll(err.Error(), "\x00", ""))
	}
	return c.SendString(string(apiModel))
}

// startServer starts service
//
//	@Summary		Start service
//	@Description	Starts service
//	@Param			name body string true "required"
//	@Tags			api
//	@Success		200	{array}	string
//	@Router			/v1/api/start [post]
func (ac *ApiController) startServer(c *fiber.Ctx) error {
	model := new(Service)
	if err := c.BodyParser(model); err != nil {
		c.SendStatus(400)
	}
	c.Locals("service", model.Name)
	c.Locals("serverId", model.ServerId)
	apiModel, err := ac.service.ApiStartServer(c)
	if err != nil {
		return c.Status(400).SendString(strings.ReplaceAll(err.Error(), "\x00", ""))
	}
	return c.SendString(apiModel)
}

// stopServer stops service
//
//	@Summary		Stop service
//	@Description	Stops service
//	@Param			name body string true "required"
//	@Tags			api
//	@Success		200	{array}	string
//	@Router			/v1/api/stop [post]
func (ac *ApiController) stopServer(c *fiber.Ctx) error {
	model := new(Service)
	if err := c.BodyParser(model); err != nil {
		c.SendStatus(400)
	}
	c.Locals("service", model.Name)
	c.Locals("serverId", model.ServerId)
	apiModel, err := ac.service.ApiStopServer(c)
	if err != nil {
		log.Print(strings.ReplaceAll(err.Error(), "\x00", ""))
		return c.Status(400).SendString(strings.ReplaceAll(err.Error(), "\x00", ""))
	}
	return c.SendString(apiModel)
}

// restartServer returns API
//
//	@Summary		Restart service
//	@Description	Restarts service
//	@Param			name body string true "required"
//	@Tags			api
//	@Success		200	{array}	string
//	@Router			/v1/api/restart [post]
func (ac *ApiController) restartServer(c *fiber.Ctx) error {
	model := new(Service)
	if err := c.BodyParser(model); err != nil {
		c.SendStatus(400)
	}
	c.Locals("service", model.Name)
	c.Locals("serverId", model.ServerId)
	apiModel, err := ac.service.ApiRestartServer(c)
	if err != nil {
		log.Print(strings.ReplaceAll(err.Error(), "\x00", ""))
		return c.Status(400).SendString(strings.ReplaceAll(err.Error(), "\x00", ""))
	}
	return c.SendString(apiModel)
}

type Service struct {
	Name     string `json:"name" xml:"name" form:"name"`
	ServerId int    `json:"serverId" xml:"serverId" form:"serverId"`
}
