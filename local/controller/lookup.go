package controller

import (
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"

	"github.com/gofiber/fiber/v2"
)

type LookupController struct {
	service *service.LookupService
}

// NewLookupController
// Initializes LookupController.
//
//	Args:
//		*services.LookupService: Lookup service
//		*Fiber.RouterGroup: Fiber Router Group
//	Returns:
//		*LookupController: Controller for "Lookup" interactions
func NewLookupController(as *service.LookupService, routeGroups *common.RouteGroups) *LookupController {
	ac := &LookupController{
		service: as,
	}
	routeGroups.Lookup.Get("/tracks", ac.getTracks)
	routeGroups.Lookup.Get("/car-models", ac.getCarModels)
	routeGroups.Lookup.Get("/driver-categories", ac.getDriverCategories)
	routeGroups.Lookup.Get("/cup-categories", ac.getCupCategories)
	routeGroups.Lookup.Get("/session-types", ac.getSessionTypes)

	return ac
}

// getTracks returns Tracks
//
//	@Summary		Return Tracks Lookup
//	@Description	Return Tracks Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/tracks [get]
func (ac *LookupController) getTracks(c *fiber.Ctx) error {
	LookupModel := ac.service.GetTracks(c)
	return c.JSON(LookupModel)
}

// getCarModels returns CarModels
//
//	@Summary		Return CarModels Lookup
//	@Description	Return CarModels Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/car-models [get]
func (ac *LookupController) getCarModels(c *fiber.Ctx) error {
	LookupModel := ac.service.GetCarModels(c)
	return c.JSON(LookupModel)
}

// getDriverCategories returns DriverCategories
//
//	@Summary		Return DriverCategories Lookup
//	@Description	Return DriverCategories Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/driver-categories [get]
func (ac *LookupController) getDriverCategories(c *fiber.Ctx) error {
	LookupModel := ac.service.GetDriverCategories(c)
	return c.JSON(LookupModel)
}

// getCupCategories returns CupCategories
//
//	@Summary		Return CupCategories Lookup
//	@Description	Return CupCategories Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/cup-categories [get]
func (ac *LookupController) getCupCategories(c *fiber.Ctx) error {
	LookupModel := ac.service.GetCupCategories(c)
	return c.JSON(LookupModel)
}

// getSessionTypes returns SessionTypes
//
//	@Summary		Return SessionTypes Lookup
//	@Description	Return SessionTypes Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/session-types [get]
func (ac *LookupController) getSessionTypes(c *fiber.Ctx) error {
	LookupModel := ac.service.GetSessionTypes(c)
	return c.JSON(LookupModel)
}
