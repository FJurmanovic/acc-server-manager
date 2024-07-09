package api

import (
	"github.com/gofiber/fiber/v2"
)

/*
Init

Initializes Web API Routes.

	Args:
		*fiber.App: Fiber Application.
*/
func Init(app *fiber.App) {
	Routes(app)
}

type API struct {
	Api string `json:"api"`
}
