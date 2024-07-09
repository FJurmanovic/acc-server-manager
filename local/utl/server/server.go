package server

import (
	"acc-server-manager/local/utl/common"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

func Start(r *fiber.App) *fiber.App {
	r.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("pong")
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}
	err := r.Listen(":" + port)
	if err != nil {
		msg := fmt.Sprintf("Running on %s:%s", common.GetIP(), port)
		println(msg)
	}
	return r
}
