package server

import (
	"acc-server-manager/local/api"
	"acc-server-manager/local/middleware/security"
	"acc-server-manager/local/utl/logging"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/swagger"
	"go.uber.org/dig"
)

func Start(di *dig.Container) *fiber.App {
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		BodyLimit:         10 * 1024 * 1024, // 10MB
	})

	// Initialize security middleware
	securityMW := security.NewSecurityMiddleware()

	// Add security middleware stack
	app.Use(securityMW.SecurityHeaders())
	app.Use(securityMW.LogSecurityEvents())
	app.Use(securityMW.TimeoutMiddleware(30 * time.Second))
	app.Use(securityMW.RequestContextTimeout(60 * time.Second))
	app.Use(securityMW.RequestSizeLimit(10 * 1024 * 1024)) // 10MB
	app.Use(securityMW.ValidateUserAgent())
	app.Use(securityMW.ValidateContentType("application/json", "application/x-www-form-urlencoded", "multipart/form-data"))
	app.Use(securityMW.InputSanitization())
	app.Use(securityMW.RateLimit(100, 1*time.Minute)) // 100 requests per minute global

	app.Use(helmet.New())

	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigin,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))

	app.Get("/swagger/*", swagger.HandlerDefault)

	api.Init(di, app)

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default port
	}

	logging.Info("Starting server on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		logging.Error("Failed to start server: %v", err)
		os.Exit(1)
	}

	return app
}
