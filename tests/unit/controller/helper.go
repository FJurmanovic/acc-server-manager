package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/cache"
	"acc-server-manager/local/utl/jwt"
	"acc-server-manager/tests"
	"os"

	"github.com/gofiber/fiber/v2"
)

type MockMiddleware struct{}

func GetTestAuthMiddleware(ms *service.MembershipService, cache *cache.InMemoryCache) *middleware.AuthMiddleware {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}

	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)

	return middleware.NewAuthMiddleware(ms, cache, jwtHandler, openJWTHandler)
}

func AddAuthToRequest(req *fiber.Ctx) {
	token := tests.MustGenerateTestToken()
	req.Request().Header.Set("Authorization", "Bearer "+token)
}
