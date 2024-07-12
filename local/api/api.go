package api

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

/*
Init

Initializes Web API Routes.

	Args:
		*fiber.App: Fiber Application.
*/
func Init(di *dig.Container, app *fiber.App) {
	Routes(di, app)
}

type API struct {
	Api string `json:"api"`
}
