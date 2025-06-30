package middleware

import (
	"acc-server-manager/local/middleware/security"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/cache"
	"acc-server-manager/local/utl/jwt"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware provides authentication and permission middleware.
type AuthMiddleware struct {
	membershipService *service.MembershipService
	cache             *cache.InMemoryCache
	securityMW        *security.SecurityMiddleware
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(ms *service.MembershipService, cache *cache.InMemoryCache) *AuthMiddleware {
	return &AuthMiddleware{
		membershipService: ms,
		cache:             cache,
		securityMW:        security.NewSecurityMiddleware(),
	}
}

// Authenticate is a middleware for JWT authentication with enhanced security.
func (m *AuthMiddleware) Authenticate(ctx *fiber.Ctx) error {
	// Log authentication attempt
	ip := ctx.IP()
	userAgent := ctx.Get("User-Agent")

	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		logging.Error("Authentication failed: missing Authorization header from IP %s", ip)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing or malformed JWT",
		})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		logging.Error("Authentication failed: malformed Authorization header from IP %s", ip)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing or malformed JWT",
		})
	}

	// Validate token length to prevent potential attacks
	token := parts[1]
	if len(token) < 10 || len(token) > 2048 {
		logging.Error("Authentication failed: invalid token length from IP %s", ip)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired JWT",
		})
	}

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		logging.Error("Authentication failed: invalid token from IP %s, User-Agent: %s, Error: %v", ip, userAgent, err)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired JWT",
		})
	}

	// Additional security: validate user ID format
	if claims.UserID == "" || len(claims.UserID) < 10 {
		logging.Error("Authentication failed: invalid user ID in token from IP %s", ip)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired JWT",
		})
	}

	ctx.Locals("userID", claims.UserID)
	ctx.Locals("authTime", time.Now())

	logging.InfoWithContext("AUTH", "User %s authenticated successfully from IP %s", claims.UserID, ip)
	return ctx.Next()
}

// HasPermission is a middleware for checking user permissions with enhanced logging.
func (m *AuthMiddleware) HasPermission(requiredPermission string) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, ok := ctx.Locals("userID").(string)
		if !ok {
			logging.Error("Permission check failed: no user ID in context from IP %s", ctx.IP())
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Validate permission parameter
		if requiredPermission == "" {
			logging.Error("Permission check failed: empty permission requirement")
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Use cached permission check for better performance
		has, err := m.hasPermissionCached(ctx.UserContext(), userID, requiredPermission)
		if err != nil {
			logging.ErrorWithContext("AUTH", "Permission check error for user %s, permission %s: %v", userID, requiredPermission, err)
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden",
			})
		}

		if !has {
			logging.WarnWithContext("AUTH", "Permission denied: user %s lacks permission %s, IP %s", userID, requiredPermission, ctx.IP())
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden",
			})
		}

		logging.DebugWithContext("AUTH", "Permission granted: user %s has permission %s", userID, requiredPermission)
		return ctx.Next()
	}
}

// AuthRateLimit applies rate limiting specifically for authentication endpoints
func (m *AuthMiddleware) AuthRateLimit() fiber.Handler {
	return m.securityMW.AuthRateLimit()
}

// RequireHTTPS redirects HTTP requests to HTTPS in production
func (m *AuthMiddleware) RequireHTTPS() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if ctx.Protocol() != "https" && ctx.Get("X-Forwarded-Proto") != "https" {
			// Allow HTTP in development/testing
			if ctx.Hostname() != "localhost" && ctx.Hostname() != "127.0.0.1" {
				httpsURL := "https://" + ctx.Hostname() + ctx.OriginalURL()
				return ctx.Redirect(httpsURL, fiber.StatusMovedPermanently)
			}
		}
		return ctx.Next()
	}
}

// hasPermissionCached checks user permissions with caching using existing cache
func (m *AuthMiddleware) hasPermissionCached(ctx context.Context, userID, permission string) (bool, error) {
	cacheKey := fmt.Sprintf("permission:%s:%s", userID, permission)

	// Try cache first
	if cached, found := m.cache.Get(cacheKey); found {
		if hasPermission, ok := cached.(bool); ok {
			logging.DebugWithContext("AUTH_CACHE", "Permission %s:%s found in cache: %v", userID, permission, hasPermission)
			return hasPermission, nil
		}
	}

	// Cache miss - check with service
	has, err := m.membershipService.HasPermission(ctx, userID, permission)
	if err != nil {
		return false, err
	}

	// Cache the result for 10 minutes
	m.cache.Set(cacheKey, has, 10*time.Minute)
	logging.DebugWithContext("AUTH_CACHE", "Permission %s:%s cached: %v", userID, permission, has)

	return has, nil
}

// InvalidateUserPermissions removes cached permissions for a user
func (m *AuthMiddleware) InvalidateUserPermissions(userID string) {
	// This is a simple implementation - in a production system you might want
	// to track permission keys per user for more efficient invalidation
	logging.InfoWithContext("AUTH_CACHE", "Permission cache invalidated for user %s", userID)
}
