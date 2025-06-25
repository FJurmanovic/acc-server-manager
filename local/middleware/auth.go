package middleware

import (
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/jwt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AuthMiddleware provides authentication and permission middleware.
type AuthMiddleware struct {
	membershipService *service.MembershipService
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(ms *service.MembershipService) *AuthMiddleware {
	return &AuthMiddleware{
		membershipService: ms,
	}
}

// Authenticate is a middleware for JWT authentication.
func (m *AuthMiddleware) Authenticate(ctx *fiber.Ctx) error {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or malformed JWT"})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or malformed JWT"})
	}

	claims, err := jwt.ValidateToken(parts[1])
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired JWT"})
	}

	ctx.Locals("userID", claims.UserID)
	return ctx.Next()
}

// HasPermission is a middleware for checking user permissions.
func (m *AuthMiddleware) HasPermission(requiredPermission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, ok := ctx.Locals("userID").(uuid.UUID)
		if !ok {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
		}

		has, err := m.membershipService.HasPermission(ctx.UserContext(), userID, requiredPermission)
		if err != nil || !has {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}

		return ctx.Next()
	}
}
