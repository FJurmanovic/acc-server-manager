package controller

import (
	"acc-server-manager/local/middleware"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/cache"
	"acc-server-manager/tests"

	"github.com/gofiber/fiber/v2"
)

// MockMiddleware simulates authentication for testing purposes
type MockMiddleware struct{}

// GetTestAuthMiddleware returns a mock auth middleware that can be used in place of the real one
// This works because we're adding real authentication tokens to requests
func GetTestAuthMiddleware(ms *service.MembershipService, cache *cache.InMemoryCache) *middleware.AuthMiddleware {
	// Cast our mock to the real type for testing
	// This is a type-unsafe cast but works for testing because we're using real JWT tokens
	return middleware.NewAuthMiddleware(ms, cache)
}

// AddAuthToRequest adds a valid authentication token to a test request
func AddAuthToRequest(req *fiber.Ctx) {
	token := tests.MustGenerateTestToken()
	req.Request().Header.Set("Authorization", "Bearer "+token)
}
