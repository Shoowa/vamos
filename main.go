package main

import (
	"vamos/config"
	"vamos/data/rdbms"
	"vamos/logging"
	"vamos/metrics"
	"vamos/router"
	"vamos/server"
)

const (
	DB_FIRST = 0
)

func main() {
	cfg := config.Read()

	logger := logging.CreateLogger(cfg)

	db1, db1Err := rdbms.ConnectDB(cfg, DB_FIRST)
	if db1Err != nil {
		logger.Error(db1Err.Error())
		panic(db1Err)
	}
	defer db1.Close()

	// Setup metrics.
	metrics.Register()

	// child logger for webserver
	srvLogger := logger.WithGroup("server")

	backbone := router.NewBackbone(
		router.WithLogger(srvLogger),
		router.WithQueryHandleForFirstDB(db1),
		router.WithDbHandle(db1),
	)

	// Launch background health checks.
	backbone.SetupHealthChecks(cfg)

	// Create router with dependencies and handlers.
	router := router.NewRouter(backbone)

	// Create a webserver with accessible dependencies.
	webserver := server.NewServer(cfg, router)

	// Activate webserver.
	go server.GracefulIgnition(webserver)
	logger.Info("HTTP Server activated")

	// Listen for termination signal. This is a blocking call.
	server.CatchSigTerm()
	logger.Info("Begin decommissioning application")

	// After receiving signal, begin deactivating server.
	shutErr := server.GracefulShutdown(webserver)

	// Record errors & force shutdown.
	if shutErr != nil {
		logger.Error("HTTP Server shutdown error", "ERR:", shutErr.Error())
		killErr := webserver.Close()
		if killErr != nil {
			logger.Error("HTTP Server kill error", "ERR:", killErr.Error())
		}
	}
	logger.Info("HTTP Server halted")
}
