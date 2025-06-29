package main

import (
	"acc-server-manager/local/utl/cache"
	"acc-server-manager/local/utl/db"
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/server"
	"fmt"
	"os"

	"go.uber.org/dig"

	_ "acc-server-manager/docs"
)

func main() {
	// Initialize logger
	logger, err := logging.Initialize()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Set up panic recovery
	defer logging.RecoverAndLog()

	di := dig.New()
	cache.Start(di)
	db.Start(di)
	server.Start(di)
}
