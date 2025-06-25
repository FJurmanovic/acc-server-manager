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
	routeGroups.Lookup.Get("/tracks", ac.GetTracks)
	routeGroups.Lookup.Get("/car-models", ac.GetCarModels)
	routeGroups.Lookup.Get("/driver-categories", ac.GetDriverCategories)
	routeGroups.Lookup.Get("/cup-categories", ac.GetCupCategories)
	routeGroups.Lookup.Get("/session-types", ac.GetSessionTypes)

	return ac
}

// getTracks returns Tracks
//
//	@Summary		Return Tracks Lookup
//	@Description	Return Tracks Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/tracks [get]
func (ac *LookupController) GetTracks(c *fiber.Ctx) error {
	result, err := ac.service.GetTracks(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching tracks",
		})
	}
	return c.JSON(result)
}

// getCarModels returns CarModels
//
//	@Summary		Return CarModels Lookup
//	@Description	Return CarModels Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/car-models [get]
func (ac *LookupController) GetCarModels(c *fiber.Ctx) error {
	result, err := ac.service.GetCarModels(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching car models",
		})
	}
	return c.JSON(result)
}

// getDriverCategories returns DriverCategories
//
//	@Summary		Return DriverCategories Lookup
//	@Description	Return DriverCategories Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/driver-categories [get]
func (ac *LookupController) GetDriverCategories(c *fiber.Ctx) error {
	result, err := ac.service.GetDriverCategories(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching driver categories",
		})
	}
	return c.JSON(result)
}

// getCupCategories returns CupCategories
//
//	@Summary		Return CupCategories Lookup
//	@Description	Return CupCategories Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/cup-categories [get]
func (ac *LookupController) GetCupCategories(c *fiber.Ctx) error {
	result, err := ac.service.GetCupCategories(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching cup categories",
		})
	}
	return c.JSON(result)
}

// getSessionTypes returns SessionTypes
//
//	@Summary		Return SessionTypes Lookup
//	@Description	Return SessionTypes Lookup
//	@Tags			Lookup
//	@Success		200	{array}	string
//	@Router			/v1/lookup/session-types [get]
func (ac *LookupController) GetSessionTypes(c *fiber.Ctx) error {
	result, err := ac.service.GetSessionTypes(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error fetching session types",
		})
	}
	return c.JSON(result)
}
