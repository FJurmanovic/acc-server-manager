package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/error_handler"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MembershipController handles API requests for membership.
type MembershipController struct {
	service      *service.MembershipService
	auth         *middleware.AuthMiddleware
	errorHandler *error_handler.ControllerErrorHandler
}

// NewMembershipController creates a new MembershipController.
func NewMembershipController(service *service.MembershipService, auth *middleware.AuthMiddleware, routeGroups *common.RouteGroups) *MembershipController {
	mc := &MembershipController{
		service:      service,
		auth:         auth,
		errorHandler: error_handler.NewControllerErrorHandler(),
	}
	// Setup initial data for membership
	if err := service.SetupInitialData(context.Background()); err != nil {
		logging.Panic(fmt.Sprintf("failed to setup initial data: %v", err))
	}

	routeGroups.Auth.Post("/login", mc.Login)

	usersGroup := routeGroups.Membership
	usersGroup.Use(mc.auth.Authenticate)
	usersGroup.Post("/", mc.auth.HasPermission(model.MembershipCreate), mc.CreateUser)
	usersGroup.Get("/", mc.auth.HasPermission(model.MembershipView), mc.ListUsers)

	usersGroup.Get("/roles", mc.auth.HasPermission(model.RoleView), mc.GetRoles)
	usersGroup.Get("/:id", mc.auth.HasPermission(model.MembershipView), mc.GetUser)
	usersGroup.Put("/:id", mc.auth.HasPermission(model.MembershipEdit), mc.UpdateUser)
	usersGroup.Delete("/:id", mc.auth.HasPermission(model.MembershipEdit), mc.DeleteUser)

	routeGroups.Auth.Get("/me", mc.auth.Authenticate, mc.GetMe)

	return mc
}

// Login handles user login.
// @Summary User login
// @Description Authenticate a user and receive a JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body object{username=string,password=string} true "Login credentials"
// @Success 200 {object} object{token=string} "JWT token"
// @Failure 400 {object} error_handler.ErrorResponse "Invalid request body"
// @Failure 401 {object} error_handler.ErrorResponse "Invalid credentials"
// @Failure 500 {object} error_handler.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (c *MembershipController) Login(ctx *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req request
	if err := ctx.BodyParser(&req); err != nil {
		return c.errorHandler.HandleParsingError(ctx, err)
	}

	logging.Debug("Login request received")
	token, err := c.service.Login(ctx.UserContext(), req.Username, req.Password)
	if err != nil {
		return c.errorHandler.HandleAuthError(ctx, err)
	}

	return ctx.JSON(fiber.Map{"token": token})
}

// CreateUser creates a new user.
// @Summary Create a new user
// @Description Create a new user account with specified role
// @Tags User Management
// @Accept json
// @Produce json
// @Param user body object{username=string,password=string,role=string} true "User details"
// @Success 200 {object} model.User "Created user details"
// @Failure 400 {object} error_handler.ErrorResponse "Invalid request body"
// @Failure 401 {object} error_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} error_handler.ErrorResponse "Insufficient permissions"
// @Failure 409 {object} error_handler.ErrorResponse "User already exists"
// @Failure 500 {object} error_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /membership [post]
func (mc *MembershipController) CreateUser(c *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return mc.errorHandler.HandleParsingError(c, err)
	}

	user, err := mc.service.CreateUser(c.UserContext(), req.Username, req.Password, req.Role)
	if err != nil {
		return mc.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(user)
}

// ListUsers lists all users.
// @Summary List all users
// @Description Get a list of all registered users
// @Tags User Management
// @Accept json
// @Produce json
// @Success 200 {array} model.User "List of users"
// @Failure 401 {object} error_handler.ErrorResponse "Unauthorized"
// @Failure 403 {object} error_handler.ErrorResponse "Insufficient permissions"
// @Failure 500 {object} error_handler.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /membership [get]
func (mc *MembershipController) ListUsers(c *fiber.Ctx) error {
	users, err := mc.service.ListUsers(c.UserContext())
	if err != nil {
		return mc.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(users)
}

// GetUser gets a single user by ID.
func (mc *MembershipController) GetUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return mc.errorHandler.HandleUUIDError(c, "user ID")
	}

	user, err := mc.service.GetUser(c.UserContext(), id)
	if err != nil {
		return mc.errorHandler.HandleNotFoundError(c, "User")
	}

	return c.JSON(user)
}

// GetMe returns the currently authenticated user's details.
func (mc *MembershipController) GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return mc.errorHandler.HandleAuthError(c, fmt.Errorf("unauthorized: user ID not found in context"))
	}

	user, err := mc.service.GetUserWithPermissions(c.UserContext(), userID)
	if err != nil {
		return mc.errorHandler.HandleNotFoundError(c, "User")
	}

	// Sanitize the user object to not expose password
	user.Password = ""

	return c.JSON(user)
}

// DeleteUser deletes a user.
func (mc *MembershipController) DeleteUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return mc.errorHandler.HandleUUIDError(c, "user ID")
	}

	err = mc.service.DeleteUser(c.UserContext(), id)
	if err != nil {
		return mc.errorHandler.HandleServiceError(c, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// UpdateUser updates a user.
func (mc *MembershipController) UpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return mc.errorHandler.HandleUUIDError(c, "user ID")
	}

	var req service.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return mc.errorHandler.HandleParsingError(c, err)
	}

	user, err := mc.service.UpdateUser(c.UserContext(), id, req)
	if err != nil {
		return mc.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(user)
}

// GetRoles returns all available roles.
func (mc *MembershipController) GetRoles(c *fiber.Ctx) error {
	roles, err := mc.service.GetAllRoles(c.UserContext())
	if err != nil {
		return mc.errorHandler.HandleServiceError(c, err)
	}

	return c.JSON(roles)
}
