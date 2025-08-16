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

// MockMiddleware simulates authentication for testing purposes
type MockMiddleware struct{}

// GetTestAuthMiddleware returns a mock auth middleware that can be used in place of the real one
// This works because we're adding real authentication tokens to requests
func GetTestAuthMiddleware(ms *service.MembershipService, cache *cache.InMemoryCache) *middleware.AuthMiddleware {
	// Use environment JWT secrets for consistency with token generation
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret) // Use same secret for test consistency
	
	// Cast our mock to the real type for testing
	// This is a type-unsafe cast but works for testing because we're using real JWT tokens
	return middleware.NewAuthMiddleware(ms, cache, jwtHandler, openJWTHandler)
}

// AddAuthToRequest adds a valid authentication token to a test request
func AddAuthToRequest(req *fiber.Ctx) {
	token := tests.MustGenerateTestToken()
	req.Request().Header.Set("Authorization", "Bearer "+token)
}
