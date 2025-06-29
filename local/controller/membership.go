package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/model"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MembershipController handles API requests for membership.
type MembershipController struct {
	service *service.MembershipService
	auth    *middleware.AuthMiddleware
}

// NewMembershipController creates a new MembershipController.
func NewMembershipController(service *service.MembershipService, auth *middleware.AuthMiddleware, routeGroups *common.RouteGroups) *MembershipController {
	mc := &MembershipController{
		service: service,
		auth:    auth,
	}
	// Setup initial data for membership
	if err := service.SetupInitialData(context.Background()); err != nil {
		logging.Panic(fmt.Sprintf("failed to setup initial data: %v", err))
	}

	routeGroups.Auth.Post("/login", mc.Login)

	usersGroup := routeGroups.Api.Group("/users", mc.auth.Authenticate)
	usersGroup.Post("/", mc.auth.HasPermission(model.MembershipCreate), mc.CreateUser)
	usersGroup.Get("/", mc.auth.HasPermission(model.MembershipView), mc.ListUsers)
	usersGroup.Get("/:id", mc.auth.HasPermission(model.MembershipView), mc.GetUser)
	usersGroup.Put("/:id", mc.auth.HasPermission(model.MembershipEdit), mc.UpdateUser)

	routeGroups.Auth.Get("/me", mc.auth.Authenticate, mc.GetMe)

	return mc
}

// Login handles user login.
func (c *MembershipController) Login(ctx *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req request
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	logging.Debug("Login request received")
	token, err := c.service.Login(ctx.UserContext(), req.Username, req.Password)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{"token": token})
}

// CreateUser creates a new user.
func (mc *MembershipController) CreateUser(c *fiber.Ctx) error {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	user, err := mc.service.CreateUser(c.UserContext(), req.Username, req.Password, req.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(user)
}

// ListUsers lists all users.
func (mc *MembershipController) ListUsers(c *fiber.Ctx) error {
	users, err := mc.service.ListUsers(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(users)
}

// GetUser gets a single user by ID.
func (mc *MembershipController) GetUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	user, err := mc.service.GetUser(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

// GetMe returns the currently authenticated user's details.
func (mc *MembershipController) GetMe(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	user, err := mc.service.GetUserWithPermissions(c.UserContext(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Sanitize the user object to not expose password
	user.Password = ""

	return c.JSON(user)
}

// UpdateUser updates a user.
func (mc *MembershipController) UpdateUser(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	var req service.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	user, err := mc.service.UpdateUser(c.UserContext(), id, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(user)
}
