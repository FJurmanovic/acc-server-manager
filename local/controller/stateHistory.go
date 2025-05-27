package controller

import (
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

	return ac
}

// getAll returns StateHistorys
//
//	@Summary		Return StateHistorys
//	@Description	Return StateHistorys
//	@Tags			StateHistory
//	@Success		200	{array}	string
//	@Router			/v1/StateHistory [get]
func (ac *StateHistoryController) getAll(c *fiber.Ctx) error {
	StateHistoryID, _ := c.ParamsInt("id")
	StateHistoryModel := ac.service.GetAll(c, StateHistoryID)
	return c.JSON(StateHistoryModel)
}