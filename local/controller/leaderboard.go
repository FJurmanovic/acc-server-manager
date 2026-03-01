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

type LeaderboardController struct {
	service      *service.LeaderboardService
	errorHandler *error_handler.ControllerErrorHandler
}

func NewLeaderboardController(ls *service.LeaderboardService, routeGroups *common.RouteGroups, auth *middleware.AuthMiddleware) *LeaderboardController {
	lc := &LeaderboardController{
		service:      ls,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	apiServerRoutes := routeGroups.Api.Group("/server/:id")
	apiServerRoutes.Get("/leaderboard", lc.Get)

	routeGroups.Leaderboard.Get("/", lc.Get)
	routeGroups.Leaderboard.Put("/", auth.Authenticate, lc.Update)

	return lc
}

// Get returns the leaderboard for a server (public, no auth required).
func (lc *LeaderboardController) Get(c *fiber.Ctx) error {
	serverIDStr := c.Params("id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		return lc.errorHandler.HandleUUIDError(c, "server ID")
	}

	data, err := lc.service.Get(c.UserContext(), serverID)
	if err != nil {
		return lc.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(data)
}

// Update replaces the leaderboard for a server (requires JWT auth).
func (lc *LeaderboardController) Update(c *fiber.Ctx) error {
	serverIDStr := c.Params("id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		return lc.errorHandler.HandleUUIDError(c, "server ID")
	}

	var input model.Leaderboard
	if err := c.BodyParser(&input); err != nil {
		return lc.errorHandler.HandleParsingError(c, err)
	}

	data, err := lc.service.Update(c.UserContext(), serverID, &input)
	if err != nil {
		return lc.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(data)
}
