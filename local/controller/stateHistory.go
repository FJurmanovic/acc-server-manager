package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"

	"github.com/gofiber/fiber/v2"
)

type StateHistoryController struct {
	service      *service.StateHistoryService
	errorHandler *error_handler.ControllerErrorHandler
}

// NewStateHistoryController
// Initializes StateHistoryController.
//
//	Args:
//		*services.StateHistoryService: StateHistory service
//		*Fiber.RouterGroup: Fiber Router Group
//	Returns:
//		*StateHistoryController: Controller for "StateHistory" interactions
func NewStateHistoryController(as *service.StateHistoryService, routeGroups *common.RouteGroups, auth *middleware.AuthMiddleware) *StateHistoryController {
	ac := &StateHistoryController{
		service:      as,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}

	routeGroups.StateHistory.Use(auth.Authenticate)
	routeGroups.StateHistory.Get("/", ac.GetAll)
	routeGroups.StateHistory.Get("/statistics", ac.GetStatistics)

	return ac
}

// GetAll returns StateHistorys
//
//	@Summary		Return StateHistorys
//	@Description	Return StateHistorys
//	@Tags			StateHistory
//	@Success		200	{array}	string
//	@Router			/v1/state-history [get]
func (ac *StateHistoryController) GetAll(c *fiber.Ctx) error {
	var filter model.StateHistoryFilter
	if err := common.ParseQueryFilter(c, &filter); err != nil {
		return ac.errorHandler.HandleValidationError(c, err, "query_filter")
	}

	result, err := ac.service.GetAll(c, &filter)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(result)
}

// getStatistics returns StateHistorys
//
//	@Summary		Return StateHistorys
//	@Description	Return StateHistorys
//	@Tags			StateHistory
//	@Success		200	{array}	string
//	@Router			/v1/state-history/statistics [get]
func (ac *StateHistoryController) GetStatistics(c *fiber.Ctx) error {
	var filter model.StateHistoryFilter
	if err := common.ParseQueryFilter(c, &filter); err != nil {
		return ac.errorHandler.HandleValidationError(c, err, "query_filter")
	}

	result, err := ac.service.GetStatistics(c, &filter)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(result)
}
