package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LogController struct {
	service      *service.ServerService
	errorHandler *error_handler.ControllerErrorHandler
}

func NewLogController(ss *service.ServerService, routeGroups *common.RouteGroups, auth *middleware.AuthMiddleware) *LogController {
	lc := &LogController{
		service:      ss,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	logRoutes := routeGroups.Log
	logRoutes.Use(auth.Authenticate)
	logRoutes.Get("/", auth.HasPermission(model.ServerView), lc.GetLogLines)

	return lc
}

// GetLogLines godoc
// @Summary Get last N lines of server log
// @Tags Server
// @Param id path string true "Server ID"
// @Param lines query int false "Number of lines to return (default 100)"
// @Success 200 {object} object{lines=[]string}
// @Security BearerAuth
// @Router /server/{id}/log [get]
func (lc *LogController) GetLogLines(c *fiber.Ctx) error {
	serverIDStr := c.Params("id")
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		return lc.errorHandler.HandleUUIDError(c, "server ID")
	}

	n, _ := strconv.Atoi(c.Query("lines", "100"))
	if n <= 0 {
		n = 100
	}
	if n > 10000 {
		n = 10000
	}

	lines, err := lc.service.GetLastLogLines(c, serverID, n)
	if err != nil {
		return lc.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(fiber.Map{"lines": lines})
}
