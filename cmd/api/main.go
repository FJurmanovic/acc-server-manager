package main

import (
	"acc-server-manager/local/utl/db"
	"acc-server-manager/local/utl/server"

	"github.com/joho/godotenv"
	"go.uber.org/dig"

	_ "acc-server-manager/docs"
)

func main() {
	godotenv.Load()
	di := dig.New()
	db.Start(di)
	server.Start(di)
}
