package middleware

import (
	"acc-server-manager/local/utl/configs"
	"acc-server-manager/local/utl/logging"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AccessKeyMiddleware provides authentication and permission middleware.
type AccessKeyMiddleware struct {
	userInfo CachedUserInfo
}

// NewAccessKeyMiddleware creates a new AccessKeyMiddleware.
func NewAccessKeyMiddleware() *AccessKeyMiddleware {
	auth := &AccessKeyMiddleware{
		userInfo: CachedUserInfo{UserID: uuid.New().String(), Username: "access_key", RoleName: "Admin", Permissions: make(map[string]bool), CachedAt: time.Now()},
	}
	return auth
}

// Authenticate is a middleware for JWT authentication with enhanced security.
func (m *AccessKeyMiddleware) Authenticate(ctx *fiber.Ctx) error {
	// Log authentication attempt
	ip := ctx.IP()
	userAgent := ctx.Get("User-Agent")

	authHeader := ctx.Get("Access-Key")
	if authHeader == "" {
		logging.Error("Authentication failed: missing Access-Key header from IP %s", ip)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing or malformed JWT",
		})
	}

	if len(authHeader) < 10 || len(authHeader) > 2048 {
		logging.Error("Authentication failed: invalid token length from IP %s", ip)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired JWT",
		})
	}

	if authHeader != configs.AccessKey {
		logging.Error("Authentication failed: invalid token from IP %s, User-Agent: %s", ip, userAgent)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired JWT",
		})
	}

	ctx.Locals("userID", m.userInfo.UserID)
	ctx.Locals("userInfo", m.userInfo)
	ctx.Locals("authTime", time.Now())

	logging.InfoWithContext("AUTH", "User %s authenticated successfully from IP %s", m.userInfo.UserID, ip)
	return ctx.Next()
}
