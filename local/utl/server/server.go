package server

import (
	"acc-server-manager/local/api"
	"acc-server-manager/local/utl/common"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"go.uber.org/dig"
)

func Start(di *dig.Container) *fiber.App {
	app := fiber.New(fiber.Config{
		EnablePrintRoutes: true,
	})

	app.Use(cors.New())

	app.Get("/swagger/*", swagger.HandlerDefault)

	file, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Print("Cannot open file logs.log")
	}
	log.SetOutput(file)

	api.Init(di, app)

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})
	port := os.Getenv("PORT")
	err = app.Listen(":" + port)
	if err != nil {
		msg := fmt.Sprintf("Running on %s:%s", common.GetIP(), port)
		println(msg)
	}
	return app
}
