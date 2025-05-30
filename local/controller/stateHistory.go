package controller

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"

	"github.com/gofiber/fiber/v2"
)

type StateHistoryController struct {
	service *service.StateHistoryService
}

// NewStateHistoryController
// Initializes StateHistoryController.
//
//	Args:
//		*services.StateHistoryService: StateHistory service
//		*Fiber.RouterGroup: Fiber Router Group
//	Returns:
//		*StateHistoryController: Controller for "StateHistory" interactions
func NewStateHistoryController(as *service.StateHistoryService, routeGroups *common.RouteGroups) *StateHistoryController {
	ac := &StateHistoryController{
		service: as,
	}

	routeGroups.StateHistory.Get("/", ac.getAll)
	routeGroups.StateHistory.Get("/statistics", ac.getStatistics)

	return ac
}

// getAll returns StateHistorys
//
//	@Summary		Return StateHistorys
//	@Description	Return StateHistorys
//	@Tags			StateHistory
//	@Success		200	{array}	string
//	@Router			/v1/state-history [get]
func (ac *StateHistoryController) getAll(c *fiber.Ctx) error {
	var filter model.StateHistoryFilter
	if err := common.ParseQueryFilter(c, &filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := ac.service.GetAll(c, &filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error retrieving state history",
		})
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
func (ac *StateHistoryController) getStatistics(c *fiber.Ctx) error {
	var filter model.StateHistoryFilter
	if err := common.ParseQueryFilter(c, &filter); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	result, err := ac.service.GetStatistics(c, &filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error retrieving state history statistics",
		})
	}

	return c.JSON(result)
}