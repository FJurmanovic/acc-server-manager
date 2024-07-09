package main

import (
	"acc-server-manager/local/api"
	"acc-server-manager/local/utl/server"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	app := fiber.New(fiber.Config{
		Immutable: true,
	})

	file, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Print("Cannot open file logs.log")
	}
	log.SetOutput(file)

	api.Init(app)

	server.Start(app)
}
