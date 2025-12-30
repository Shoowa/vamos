package main

import (
	"github.com/Shoowa/vamos/config"
	"github.com/Shoowa/vamos/data/cache"
	"github.com/Shoowa/vamos/data/rdbms"
	"github.com/Shoowa/vamos/logging"
	"github.com/Shoowa/vamos/router"
	"github.com/Shoowa/vamos/secrets"
	"github.com/Shoowa/vamos/server"

	"_example/routes"
)

const DB_FIRST = 0

func main() {
	cfg := config.Read()
	cfg.Secrets.Openbao.ReadToken()

	logger := logging.CreateLogger(cfg)

	// Create secretsClient to read multiple credentials.
	secretsReader := new(secrets.SkeletonKey)
	secretsReader.Create(cfg)

	db1, db1Err := rdbms.ConnectDB(cfg, DB_FIRST)
	if db1Err != nil {
		logger.Error(db1Err.Error())
		panic(db1Err)
	}
	defer db1.Close()

	cache, cacheErr := cache.CreateClient(cfg, secretsReader)
	if cacheErr != nil {
		panic(cacheErr.Error())
	}
	defer cache.Close()

	// child logger for webserver
	srvLogger := logger.WithGroup("server")

	backbone := router.NewBackbone(
		router.WithLogger(srvLogger),
		router.WithDbHandle(db1),
		router.WithCache(cache),
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

	// Read TLS certificate & key.
	X509, X509Err := secretsReader.ReadTlsCertAndKey(cfg.HttpServer.TlsServer)
	if X509Err != nil {
		logger.Error(X509Err.Error())
		panic(X509Err)
	}

	// Create a webserver with accessible dependencies.
	webserver := server.NewServer(cfg, rtr, X509, srvLogger)

	// Activate webserver.
	server.Start(logger, webserver)
}
