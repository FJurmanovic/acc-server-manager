package server

import (
	"acc-server-manager/local/api"
	"acc-server-manager/local/utl/logging"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"go.uber.org/dig"
)

func Start(di *dig.Container) *fiber.App {
	// Initialize logger
	logger, err := logging.Initialize()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Set up panic recovery
	defer logging.RecoverAndLog()

	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
	})

	app.Use(cors.New())

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
