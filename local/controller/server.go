package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"

	"github.com/gofiber/fiber/v2"
)

type ServerController struct {
	service *service.ServerService
}

// NewServerController initializes ServerController.
func NewServerController(ss *service.ServerService, routeGroups *common.RouteGroups, auth *middleware.AuthMiddleware) *ServerController {
	ac := &ServerController{
		service: ss,
	}

	serverRoutes := routeGroups.Server
	serverRoutes.Use(auth.Authenticate)

	serverRoutes.Get("/", auth.HasPermission(model.ServerView), ac.GetAll)
	serverRoutes.Get("/:id", auth.HasPermission(model.ServerView), ac.GetById)
	serverRoutes.Post("/", auth.HasPermission(model.ServerCreate), ac.CreateServer)
	serverRoutes.Put("/:id", auth.HasPermission(model.ServerUpdate), ac.UpdateServer)
	serverRoutes.Delete("/:id", auth.HasPermission(model.ServerDelete), ac.DeleteServer)
	return ac
}

// GetAll returns Servers
func (ac *ServerController) GetAll(c *fiber.Ctx) error {
	var filter model.ServerFilter
	if err := common.ParseQueryFilter(c, &filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	ServerModel, err := ac.service.GetAll(c, &filter)
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}
	return c.JSON(ServerModel)
}

// GetById returns a single server by its ID
func (ac *ServerController) GetById(c *fiber.Ctx) error {
	serverID, _ := c.ParamsInt("id")
	ServerModel, err := ac.service.GetById(c, serverID)
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}
	return c.JSON(ServerModel)
}

// CreateServer creates a new server
func (ac *ServerController) CreateServer(c *fiber.Ctx) error {
	server := new(model.Server)
	if err := c.BodyParser(server); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	ac.service.GenerateServerPath(server)
	if err := ac.service.CreateServer(c, server); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(server)
}

// UpdateServer updates an existing server
func (ac *ServerController) UpdateServer(c *fiber.Ctx) error {
	serverID, _ := c.ParamsInt("id")
	server := new(model.Server)
	if err := c.BodyParser(server); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	server.ID = uint(serverID)

	if err := ac.service.UpdateServer(c, server); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(server)
}

// DeleteServer deletes a server
func (ac *ServerController) DeleteServer(c *fiber.Ctx) error {
	serverID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid server ID"})
	}

	if err := ac.service.DeleteServer(c, serverID); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.SendStatus(204)
}