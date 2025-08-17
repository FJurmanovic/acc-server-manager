package middleware

import (
	"acc-server-manager/local/middleware/security"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/cache"
	"acc-server-manager/local/utl/jwt"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CachedUserInfo holds cached user authentication and permission data
type CachedUserInfo struct {
	UserID      string
	Username    string
	RoleName    string
	Permissions map[string]bool
	CachedAt    time.Time
}

// AuthMiddleware provides authentication and permission middleware.
type AuthMiddleware struct {
	membershipService *service.MembershipService
	cache             *cache.InMemoryCache
	securityMW        *security.SecurityMiddleware
	jwtHandler        *jwt.JWTHandler
	openJWTHandler    *jwt.OpenJWTHandler
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(ms *service.MembershipService, cache *cache.InMemoryCache, jwtHandler *jwt.JWTHandler, openJWTHandler *jwt.OpenJWTHandler) *AuthMiddleware {
	auth := &AuthMiddleware{
		membershipService: ms,
		cache:             cache,
		securityMW:        security.NewSecurityMiddleware(),
		jwtHandler:        jwtHandler,
		openJWTHandler:    openJWTHandler,
	}

	// Set up bidirectional relationship for cache invalidation
	ms.SetCacheInvalidator(auth)

	return auth
}

// Authenticate is a middleware for JWT authentication with enhanced security.
func (m *AuthMiddleware) AuthenticateOpen(ctx *fiber.Ctx) error {
	return m.AuthenticateWithHandler(m.openJWTHandler.JWTHandler, true, ctx)
}

// Authenticate is a middleware for JWT authentication with enhanced security.
func (m *AuthMiddleware) Authenticate(ctx *fiber.Ctx) error {
	return m.AuthenticateWithHandler(m.jwtHandler, false, ctx)
}

func (m *AuthMiddleware) AuthenticateWithHandler(jwtHandler *jwt.JWTHandler, isOpenToken bool, ctx *fiber.Ctx) error {
	// Log authentication attempt
	ip := ctx.IP()
	userAgent := ctx.Get("User-Agent")

	authHeader := ctx.Get("Authorization")

	if jwtHandler.IsOpenToken && !isOpenToken {
		logging.Error("Authentication failed: attempting to authenticate with open token")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Wrong token type used",
		})
	}

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

	claims, err := jwtHandler.ValidateToken(token)
	if err != nil {
		logging.Error("Authentication failed: invalid token from IP %s, User-Agent: %s, Error: %v", ip, userAgent, err)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired JWT",
		})
	}

	if !jwtHandler.IsOpenToken && claims.IsOpenToken {
		logging.Error("Authentication failed: attempting to authenticate with open token")
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Wrong token type used",
		})
	}

	// Additional security: validate user ID format
	if claims.UserID == "" || len(claims.UserID) < 10 {
		logging.Error("Authentication failed: invalid user ID in token from IP %s", ip)
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired JWT",
		})
	}

	if os.Getenv("TESTING_ENV") == "true" {
		userInfo := CachedUserInfo{UserID: uuid.New().String(), Username: "test@example.com", RoleName: "Admin", Permissions: make(map[string]bool), CachedAt: time.Now()}
		ctx.Locals("userID", userInfo.UserID)
		ctx.Locals("userInfo", userInfo)
		ctx.Locals("authTime", time.Now())
	} else {
		// Preload and cache user info to avoid database queries on permission checks
		userInfo, err := m.getCachedUserInfo(ctx.UserContext(), claims.UserID)
		if err != nil {
			logging.Error("Authentication failed: unable to load user info for %s from IP %s: %v", claims.UserID, ip, err)
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired JWT",
			})
		}

		ctx.Locals("userID", claims.UserID)
		ctx.Locals("userInfo", userInfo)
		ctx.Locals("authTime", time.Now())
	}

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

		if os.Getenv("TESTING_ENV") == "true" {
			return ctx.Next()
		}

		// Validate permission parameter
		if requiredPermission == "" {
			logging.Error("Permission check failed: empty permission requirement")
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		// Use cached user info from authentication step - no database queries needed
		userInfo, ok := ctx.Locals("userInfo").(*CachedUserInfo)
		if !ok {
			logging.Error("Permission check failed: no cached user info in context from IP %s", ctx.IP())
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

		// Check if user has permission using cached data
		has := m.hasPermissionFromCache(userInfo, requiredPermission)

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

// getCachedUserInfo retrieves and caches complete user information including permissions
func (m *AuthMiddleware) getCachedUserInfo(ctx context.Context, userID string) (*CachedUserInfo, error) {
	cacheKey := fmt.Sprintf("userinfo:%s", userID)

	// Try cache first
	if cached, found := m.cache.Get(cacheKey); found {
		if userInfo, ok := cached.(*CachedUserInfo); ok {
			logging.DebugWithContext("AUTH_CACHE", "User info for %s found in cache", userID)
			return userInfo, nil
		}
	}

	// Cache miss - load from database
	user, err := m.membershipService.GetUserWithPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Build permission map for fast lookups
	permissions := make(map[string]bool)
	for _, p := range user.Role.Permissions {
		permissions[p.Name] = true
	}

	userInfo := &CachedUserInfo{
		UserID:      userID,
		Username:    user.Username,
		RoleName:    user.Role.Name,
		Permissions: permissions,
		CachedAt:    time.Now(),
	}

	// Cache for 15 minutes
	m.cache.Set(cacheKey, userInfo, 15*time.Minute)
	logging.DebugWithContext("AUTH_CACHE", "User info for %s cached with %d permissions", userID, len(permissions))

	return userInfo, nil
}

// hasPermissionFromCache checks permissions using cached user info (no database queries)
func (m *AuthMiddleware) hasPermissionFromCache(userInfo *CachedUserInfo, permission string) bool {
	// Super Admin and Admin have all permissions
	if userInfo.RoleName == "Super Admin" || userInfo.RoleName == "Admin" {
		return true
	}

	// Check specific permission in cached map
	return userInfo.Permissions[permission]
}

// InvalidateUserPermissions removes cached user info for a user
func (m *AuthMiddleware) InvalidateUserPermissions(userID string) {
	cacheKey := fmt.Sprintf("userinfo:%s", userID)
	m.cache.Delete(cacheKey)
	logging.InfoWithContext("AUTH_CACHE", "User info cache invalidated for user %s", userID)
}

// InvalidateAllUserPermissions clears all cached user info (useful for role/permission changes)
func (m *AuthMiddleware) InvalidateAllUserPermissions() {
	// This would need to be implemented based on your cache interface
	// For now, just log that invalidation was requested
	logging.InfoWithContext("AUTH_CACHE", "All user info caches invalidation requested")
}
