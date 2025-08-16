package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"
	"acc-server-manager/local/utl/jwt"

	"github.com/gofiber/fiber/v2"
)

type Steam2FAController struct {
	tfaManager   *model.Steam2FAManager
	errorHandler *error_handler.ControllerErrorHandler
	jwtHandler   *jwt.OpenJWTHandler
}

func NewSteam2FAController(tfaManager *model.Steam2FAManager, routeGroups *common.RouteGroups, auth *middleware.AuthMiddleware, jwtHandler *jwt.OpenJWTHandler) *Steam2FAController {
	controller := &Steam2FAController{
		tfaManager:   tfaManager,
		errorHandler: error_handler.NewControllerErrorHandler(),
		jwtHandler:   jwtHandler,
	}

	steam2faRoutes := routeGroups.Steam2FA
	steam2faRoutes.Use(auth.AuthenticateOpen)

	// Define routes
	steam2faRoutes.Get("/pending", auth.HasPermission(model.ServerView), controller.GetPendingRequests)
	steam2faRoutes.Get("/:id", auth.HasPermission(model.ServerView), controller.GetRequest)
	steam2faRoutes.Post("/:id/complete", auth.HasPermission(model.ServerUpdate), controller.CompleteRequest)
	steam2faRoutes.Post("/:id/cancel", auth.HasPermission(model.ServerUpdate), controller.CancelRequest)

	return controller
}

// GetPendingRequests gets all pending 2FA requests
//
//	@Summary		Get pending 2FA requests
//	@Description	Get all pending Steam 2FA authentication requests
//	@Tags			Steam 2FA
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.Steam2FARequest
//	@Failure		500	{object}	error_handler.ErrorResponse
//	@Router			/steam2fa/pending [get]
func (c *Steam2FAController) GetPendingRequests(ctx *fiber.Ctx) error {
	requests := c.tfaManager.GetPendingRequests()
	return ctx.JSON(requests)
}

// GetRequest gets a specific 2FA request by ID
//
//	@Summary		Get 2FA request
//	@Description	Get a specific Steam 2FA authentication request by ID
//	@Tags			Steam 2FA
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"2FA Request ID"
//	@Success		200	{object}	model.Steam2FARequest
//	@Failure		404	{object}	error_handler.ErrorResponse
//	@Failure		500	{object}	error_handler.ErrorResponse
//	@Router			/steam2fa/{id} [get]
func (c *Steam2FAController) GetRequest(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return c.errorHandler.HandleError(ctx, fiber.ErrBadRequest, fiber.StatusBadRequest)
	}

	request, exists := c.tfaManager.GetRequest(id)
	if !exists {
		return c.errorHandler.HandleNotFoundError(ctx, "2FA request")
	}

	return ctx.JSON(request)
}

// CompleteRequest marks a 2FA request as completed
//
//	@Summary		Complete 2FA request
//	@Description	Mark a Steam 2FA authentication request as completed
//	@Tags			Steam 2FA
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"2FA Request ID"
//	@Success		200	{object}	model.Steam2FARequest
//	@Failure		400	{object}	error_handler.ErrorResponse
//	@Failure		404	{object}	error_handler.ErrorResponse
//	@Failure		500	{object}	error_handler.ErrorResponse
//	@Router			/steam2fa/{id}/complete [post]
func (c *Steam2FAController) CompleteRequest(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return c.errorHandler.HandleError(ctx, fiber.ErrBadRequest, fiber.StatusBadRequest)
	}

	if err := c.tfaManager.CompleteRequest(id); err != nil {
		return c.errorHandler.HandleError(ctx, err, fiber.StatusBadRequest)
	}

	request, exists := c.tfaManager.GetRequest(id)
	if !exists {
		return c.errorHandler.HandleNotFoundError(ctx, "2FA request")
	}

	return ctx.JSON(request)
}

// CancelRequest cancels a 2FA request
//
//	@Summary		Cancel 2FA request
//	@Description	Cancel a Steam 2FA authentication request
//	@Tags			Steam 2FA
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"2FA Request ID"
//	@Success		200	{object}	model.Steam2FARequest
//	@Failure		400	{object}	error_handler.ErrorResponse
//	@Failure		404	{object}	error_handler.ErrorResponse
//	@Failure		500	{object}	error_handler.ErrorResponse
//	@Router			/steam2fa/{id}/cancel [post]
func (c *Steam2FAController) CancelRequest(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	if id == "" {
		return c.errorHandler.HandleError(ctx, fiber.ErrBadRequest, fiber.StatusBadRequest)
	}

	if err := c.tfaManager.ErrorRequest(id, "cancelled by user"); err != nil {
		return c.errorHandler.HandleError(ctx, err, fiber.StatusBadRequest)
	}

	request, exists := c.tfaManager.GetRequest(id)
	if !exists {
		return c.errorHandler.HandleNotFoundError(ctx, "2FA request")
	}

	return ctx.JSON(request)
}
