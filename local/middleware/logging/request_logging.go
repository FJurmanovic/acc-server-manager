package logging

import (
	"acc-server-manager/local/utl/logging"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RequestLoggingMiddleware struct {
	infoLogger *logging.InfoLogger
}

func NewRequestLoggingMiddleware() *RequestLoggingMiddleware {
	return &RequestLoggingMiddleware{
		infoLogger: logging.GetInfoLogger(),
	}
}

func (rlm *RequestLoggingMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		userAgent := c.Get("User-Agent")
		if userAgent == "" {
			userAgent = "Unknown"
		}

		rlm.infoLogger.LogRequest(c.Method(), c.OriginalURL(), userAgent)

		err := c.Next()

		duration := time.Since(start)

		statusCode := c.Response().StatusCode()
		rlm.infoLogger.LogResponse(c.Method(), c.OriginalURL(), statusCode, duration.String())

		if err != nil {
			logging.ErrorWithContext("REQUEST_MIDDLEWARE", "Request failed: %v", err)
		}

		return err
	}
}

var globalRequestLoggingMiddleware *RequestLoggingMiddleware

func GetRequestLoggingMiddleware() *RequestLoggingMiddleware {
	if globalRequestLoggingMiddleware == nil {
		globalRequestLoggingMiddleware = NewRequestLoggingMiddleware()
	}
	return globalRequestLoggingMiddleware
}

func Handler() fiber.Handler {
	return GetRequestLoggingMiddleware().Handler()
}
