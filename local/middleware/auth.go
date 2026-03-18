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

type CachedUserInfo struct {
	UserID      string
	Username    string
	RoleName    string
	Permissions map[string]bool
	CachedAt    time.Time
}

type AuthMiddleware struct {
	membershipService *service.MembershipService
	cache             *cache.InMemoryCache
	securityMW        *security.SecurityMiddleware
	jwtHandler        *jwt.JWTHandler
	openJWTHandler    *jwt.OpenJWTHandler
}

func NewAuthMiddleware(ms *service.MembershipService, cache *cache.InMemoryCache, jwtHandler *jwt.JWTHandler, openJWTHandler *jwt.OpenJWTHandler) *AuthMiddleware {
	auth := &AuthMiddleware{
		membershipService: ms,
		cache:             cache,
		securityMW:        security.NewSecurityMiddleware(),
		jwtHandler:        jwtHandler,
		openJWTHandler:    openJWTHandler,
	}

	ms.SetCacheInvalidator(auth)

	return auth
}

func (m *AuthMiddleware) AuthenticateOpen(ctx *fiber.Ctx) error {
	return m.AuthenticateWithHandler(m.openJWTHandler.JWTHandler, true, ctx)
}

func (m *AuthMiddleware) Authenticate(ctx *fiber.Ctx) error {
	return m.AuthenticateWithHandler(m.jwtHandler, false, ctx)
}

func (m *AuthMiddleware) AuthenticateWithHandler(jwtHandler *jwt.JWTHandler, isOpenToken bool, ctx *fiber.Ctx) error {
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

		if requiredPermission == "" {
			logging.Error("Permission check failed: empty permission requirement")
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}

		userInfo, ok := ctx.Locals("userInfo").(*CachedUserInfo)
		if !ok {
			logging.Error("Permission check failed: no cached user info in context from IP %s", ctx.IP())
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}

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

func (m *AuthMiddleware) AuthRateLimit() fiber.Handler {
	return m.securityMW.AuthRateLimit()
}

func (m *AuthMiddleware) RequireHTTPS() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if ctx.Protocol() != "https" && ctx.Get("X-Forwarded-Proto") != "https" {
			if ctx.Hostname() != "localhost" && ctx.Hostname() != "127.0.0.1" {
				httpsURL := "https://" + ctx.Hostname() + ctx.OriginalURL()
				return ctx.Redirect(httpsURL, fiber.StatusMovedPermanently)
			}
		}
		return ctx.Next()
	}
}

func (m *AuthMiddleware) getCachedUserInfo(ctx context.Context, userID string) (*CachedUserInfo, error) {
	cacheKey := fmt.Sprintf("userinfo:%s", userID)

	if cached, found := m.cache.Get(cacheKey); found {
		if userInfo, ok := cached.(*CachedUserInfo); ok {
			logging.DebugWithContext("AUTH_CACHE", "User info for %s found in cache", userID)
			return userInfo, nil
		}
	}

	user, err := m.membershipService.GetUserWithPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

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

	m.cache.Set(cacheKey, userInfo, 15*time.Minute)
	logging.DebugWithContext("AUTH_CACHE", "User info for %s cached with %d permissions", userID, len(permissions))

	return userInfo, nil
}

func (m *AuthMiddleware) hasPermissionFromCache(userInfo *CachedUserInfo, permission string) bool {
	if userInfo.RoleName == "Super Admin" || userInfo.RoleName == "Admin" {
		return true
	}

	return userInfo.Permissions[permission]
}

func (m *AuthMiddleware) InvalidateUserPermissions(userID string) {
	cacheKey := fmt.Sprintf("userinfo:%s", userID)
	m.cache.Delete(cacheKey)
	logging.InfoWithContext("AUTH_CACHE", "User info cache invalidated for user %s", userID)
}

func (m *AuthMiddleware) InvalidateAllUserPermissions() {
	logging.InfoWithContext("AUTH_CACHE", "All user info caches invalidation requested")
}
