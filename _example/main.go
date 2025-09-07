package main

import (
	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/data/rdbms"
	"github.com/Shoowa/vamos/logging"
	"github.com/Shoowa/vamos/router"
	"github.com/Shoowa/vamos/server"

	"_example/routes"
)

const DB_FIRST = 0

func main() {
	cfg := config.Read()

	logger := logging.CreateLogger(cfg)

	db1, db1Err := rdbms.ConnectDB(cfg, DB_FIRST)
	if db1Err != nil {
		logger.Error(db1Err.Error())
		panic(db1Err)
	}
	defer db1.Close()

	// child logger for webserver
	srvLogger := logger.WithGroup("server")

	backbone := router.NewBackbone(
		router.WithLogger(srvLogger),
		router.WithDbHandle(db1),
	)

	// Launch background health checks.
	backbone.SetupHealthChecks(cfg)

	// Wrap the backbone with a native struct that has its own HTTP Handlers.
	backboneWrapper := routes.WrapBackbone(backbone)

	// Create router with dependencies and errHandlers.
	rtr := router.NewRouter(backboneWrapper)
	rtr.AddRoutes(backboneWrapper)

	// Also add a HTTP Handler directly to router.
	rtr.Router.HandleFunc("GET /test1", backboneWrapper.Hndlr1)

	// Create a webserver with accessible dependencies.
	webserver := server.NewServer(cfg, rtr)

	// Activate webserver.
	server.Start(logger, webserver)
}
