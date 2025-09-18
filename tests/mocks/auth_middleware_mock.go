package mocks

import (
	"acc-server-manager/local/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MockAuthMiddleware struct{}

func NewMockAuthMiddleware() *MockAuthMiddleware {
	return &MockAuthMiddleware{}
}

func (m *MockAuthMiddleware) Authenticate(ctx *fiber.Ctx) error {
	mockUserID := uuid.New().String()
	ctx.Locals("userID", mockUserID)

	mockUserInfo := &middleware.CachedUserInfo{
		UserID:      mockUserID,
		Username:    "test_user",
		RoleName:    "Admin",
		Permissions: map[string]bool{"*": true},
		CachedAt:    time.Now(),
	}

	ctx.Locals("userInfo", mockUserInfo)
	ctx.Locals("authTime", time.Now())

	return ctx.Next()
}

func (m *MockAuthMiddleware) HasPermission(requiredPermission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Next()
	}
}

func (m *MockAuthMiddleware) AuthRateLimit() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Next()
	}
}

func (m *MockAuthMiddleware) RequireHTTPS() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Next()
	}
}
