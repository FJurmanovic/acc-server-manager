package mocks

import (
	"acc-server-manager/local/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MockAuthMiddleware provides a test implementation of AuthMiddleware
// that can be used as a drop-in replacement for the real AuthMiddleware
type MockAuthMiddleware struct{}

// NewMockAuthMiddleware creates a new MockAuthMiddleware
func NewMockAuthMiddleware() *MockAuthMiddleware {
	return &MockAuthMiddleware{}
}

// Authenticate is a middleware that allows all requests without authentication for testing
func (m *MockAuthMiddleware) Authenticate(ctx *fiber.Ctx) error {
	// Set a mock user ID in context
	mockUserID := uuid.New().String()
	ctx.Locals("userID", mockUserID)

	// Set mock user info
	mockUserInfo := &middleware.CachedUserInfo{
		UserID:      mockUserID,
		Username:    "test_user",
		RoleName:    "Admin", // Admin role to bypass permission checks
		Permissions: map[string]bool{"*": true},
		CachedAt:    time.Now(),
	}

	ctx.Locals("userInfo", mockUserInfo)
	ctx.Locals("authTime", time.Now())

	return ctx.Next()
}

// HasPermission is a middleware that allows all permission checks to pass for testing
func (m *MockAuthMiddleware) HasPermission(requiredPermission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Next()
	}
}

// AuthRateLimit is a test implementation that allows all requests
func (m *MockAuthMiddleware) AuthRateLimit() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Next()
	}
}

// RequireHTTPS is a test implementation that allows all HTTP requests
func (m *MockAuthMiddleware) RequireHTTPS() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Next()
	}
}
