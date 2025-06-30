package logging

import (
	"acc-server-manager/local/utl/logging"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RequestLoggingMiddleware logs HTTP requests and responses
type RequestLoggingMiddleware struct {
	infoLogger *logging.InfoLogger
}

// NewRequestLoggingMiddleware creates a new request logging middleware
func NewRequestLoggingMiddleware() *RequestLoggingMiddleware {
	return &RequestLoggingMiddleware{
		infoLogger: logging.GetInfoLogger(),
	}
}

// Handler returns the middleware handler function
func (rlm *RequestLoggingMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Record start time
		start := time.Now()

		// Log incoming request
		userAgent := c.Get("User-Agent")
		if userAgent == "" {
			userAgent = "Unknown"
		}

		rlm.infoLogger.LogRequest(c.Method(), c.OriginalURL(), userAgent)

		// Continue to next handler
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log response
		statusCode := c.Response().StatusCode()
		rlm.infoLogger.LogResponse(c.Method(), c.OriginalURL(), statusCode, duration.String())

		// Log error if present
		if err != nil {
			logging.ErrorWithContext("REQUEST_MIDDLEWARE", "Request failed: %v", err)
		}

		return err
	}
}

// Global request logging middleware instance
var globalRequestLoggingMiddleware *RequestLoggingMiddleware

// GetRequestLoggingMiddleware returns the global request logging middleware
func GetRequestLoggingMiddleware() *RequestLoggingMiddleware {
	if globalRequestLoggingMiddleware == nil {
		globalRequestLoggingMiddleware = NewRequestLoggingMiddleware()
	}
	return globalRequestLoggingMiddleware
}

// Handler returns the global request logging middleware handler
func Handler() fiber.Handler {
	return GetRequestLoggingMiddleware().Handler()
}
