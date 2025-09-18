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
		ReadTimeout:       20 * time.Minute,
		WriteTimeout:      20 * time.Minute,
		IdleTimeout:       25 * time.Minute,
		BodyLimit:         10 * 1024 * 1024,
	})

	securityMW := security.NewSecurityMiddleware()

	app.Use(securityMW.SecurityHeaders())
	app.Use(securityMW.LogSecurityEvents())
	app.Use(securityMW.TimeoutMiddleware(20 * time.Minute))
	app.Use(securityMW.RequestContextTimeout(20 * time.Minute))
	app.Use(securityMW.RequestSizeLimit(10 * 1024 * 1024))
	app.Use(securityMW.ValidateUserAgent())
	app.Use(securityMW.ValidateContentType("application/json", "application/x-www-form-urlencoded", "multipart/form-data"))
	app.Use(securityMW.InputSanitization())
	app.Use(securityMW.RateLimit(100, 1*time.Minute))

	app.Use(helmet.New())

	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:3000"
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
		port = "3000"
	}

	logging.Info("Starting server on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		logging.Error("Failed to start server: %v", err)
		os.Exit(1)
	}

	return app
}
