package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"

	"github.com/gofiber/fiber/v2"
)

type SteamController struct {
	steamService *service.SteamService
	auth         *middleware.AuthMiddleware
	errorHandler *error_handler.ControllerErrorHandler
}

func NewSteamController(steamService *service.SteamService, auth *middleware.AuthMiddleware, routeGroups *common.RouteGroups) *SteamController {
	sc := &SteamController{
		steamService: steamService,
		auth:         auth,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	routeGroups.Steam.Use(auth.Authenticate, auth.RequireSuperAdmin())
	routeGroups.Steam.Post("/credentials", sc.SetCredentials)

	return sc
}

func (sc *SteamController) SetCredentials(c *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return sc.errorHandler.HandleParsingError(c, err)
	}

	creds := &model.SteamCredentials{
		Username: req.Username,
		Password: req.Password,
	}

	if err := creds.Validate(); err != nil {
		return sc.errorHandler.HandleServiceError(c, err)
	}

	if err := sc.steamService.SaveCredentials(c.UserContext(), creds); err != nil {
		return sc.errorHandler.HandleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": creds.ID})
}
