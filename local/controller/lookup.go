package controller

import (
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"

	"github.com/gofiber/fiber/v2"
)

type LookupController struct {
	service      *service.LookupService
	errorHandler *error_handler.ControllerErrorHandler
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
		service:      as,
		errorHandler: error_handler.NewControllerErrorHandler(),
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
//	@Summary		Get available tracks
//	@Description	Get a list of all available ACC tracks with their identifiers
//	@Tags			Lookups
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	object{id=string,name=string} "List of tracks"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/v1/lookup/tracks [get]
func (ac *LookupController) GetTracks(c *fiber.Ctx) error {
	result, err := ac.service.GetTracks(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(result)
}

// getCarModels returns CarModels
//
//	@Summary		Get available car models
//	@Description	Get a list of all available ACC car models with their identifiers
//	@Tags			Lookups
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	object{id=string,name=string,class=string} "List of car models"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/v1/lookup/car-models [get]
func (ac *LookupController) GetCarModels(c *fiber.Ctx) error {
	result, err := ac.service.GetCarModels(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(result)
}

// getDriverCategories returns DriverCategories
//
//	@Summary		Get driver categories
//	@Description	Get a list of all driver categories (Bronze, Silver, Gold, Platinum)
//	@Tags			Lookups
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	object{id=number,name=string,description=string} "List of driver categories"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/v1/lookup/driver-categories [get]
func (ac *LookupController) GetDriverCategories(c *fiber.Ctx) error {
	result, err := ac.service.GetDriverCategories(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(result)
}

// getCupCategories returns CupCategories
//
//	@Summary		Get cup categories
//	@Description	Get a list of all available racing cup categories
//	@Tags			Lookups
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	object{id=number,name=string} "List of cup categories"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/v1/lookup/cup-categories [get]
func (ac *LookupController) GetCupCategories(c *fiber.Ctx) error {
	result, err := ac.service.GetCupCategories(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(result)
}

// getSessionTypes returns SessionTypes
//
//	@Summary		Get session types
//	@Description	Get a list of all available session types (Practice, Qualifying, Race)
//	@Tags			Lookups
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	object{id=string,name=string,code=string} "List of session types"
//	@Failure		401	{object} error_handler.ErrorResponse "Unauthorized"
//	@Failure		500	{object} error_handler.ErrorResponse "Internal server error"
//	@Security		BearerAuth
//	@Router			/v1/lookup/session-types [get]
func (ac *LookupController) GetSessionTypes(c *fiber.Ctx) error {
	result, err := ac.service.GetSessionTypes(c)
	if err != nil {
		return ac.errorHandler.HandleServiceError(c, err)
	}
	return c.JSON(result)
}
