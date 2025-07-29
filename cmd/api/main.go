package main

import (
	"acc-server-manager/local/utl/cache"
	"acc-server-manager/local/utl/configs"
	"acc-server-manager/local/utl/db"
	"acc-server-manager/local/utl/jwt"
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/server"
	"fmt"
	"os"

	"go.uber.org/dig"

	_ "acc-server-manager/swagger"
)

func main() {
	configs.Init()
	jwt.Init()
	// Initialize new logging system
	if err := logging.InitializeLogging(); err != nil {
		fmt.Printf("Failed to initialize logging system: %v\n", err)
		os.Exit(1)
	}

	// Get legacy logger for backward compatibility
	logger := logging.GetLegacyLogger()
	if logger != nil {
		defer logger.Close()
	}

	// Set up panic recovery
	defer logging.RecoverAndLog()

	// Log application startup
	logging.InfoStartup("APPLICATION", "ACC Server Manager starting up")

	di := dig.New()
	cache.Start(di)
	db.Start(di)
	server.Start(di)

	// Log successful startup
	logging.InfoStartup("APPLICATION", "ACC Server Manager started successfully")
}
