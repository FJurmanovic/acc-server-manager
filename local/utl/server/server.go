package server

import (
	"acc-server-manager/local/api"
	"acc-server-manager/local/utl/logging"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/swagger"
	"go.uber.org/dig"
)

func Start(di *dig.Container) *fiber.App {
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
	})

	app.Use(helmet.New())

	allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigin,
		AllowHeaders: "Origin, Content-Type, Accept",
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
