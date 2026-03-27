package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"

	"github.com/gofiber/fiber/v2"
)

type ActivityLogController struct {
	service      *service.ActivityLogService
	errorHandler *error_handler.ControllerErrorHandler
}

func NewActivityLogController(as *service.ActivityLogService, routeGroups *common.RouteGroups, auth *middleware.AuthMiddleware) *ActivityLogController {
	ac := &ActivityLogController{
		service:      as,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	// Global activity log: GET /activity-log (admin use)
	routeGroups.Groups.Group("/activity-log").Use(auth.Authenticate).Get("/", ac.GetAllLogs)

	return ac
}

// GetAllLogs returns activity logs across all servers (admin).
//
//	@Summary		Return all activity logs
//	@Tags			ActivityLog
//	@Success		200	{array}	model.ActivityLog
//	@Router			/activity-log [get]
func (ac *ActivityLogController) GetAllLogs(c *fiber.Ctx) error {
	var filter model.ActivityLogFilter
	if err := common.ParseQueryFilter(c, &filter); err != nil {
		return ac.errorHandler.HandleValidationError(c, err, "query_filter")
	}

	result, err := ac.service.GetAll(c.UserContext(), &filter)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(result)
}
