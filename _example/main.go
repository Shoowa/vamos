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
	// Read configuration file. Read OPENBAO_TOKEN.
	cfg := config.Read()
	cfg.Secrets.Openbao.ReadToken()

	// Create a structured JSON logger.
	logger := logging.CreateLogger(cfg)

	// Connect to Postgres server. The ConnectDB func builds its own copy of the
	// Openbao client and assigns it to a Postgres "BeforeConnect" func to
	// re-use whenever a password changes.
	db1, db1Err := rdbms.ConnectDB(cfg, DB_FIRST)
	if db1Err != nil {
		logger.Error(db1Err.Error())
		panic(db1Err)
	}
	defer db1.Close()

	// Create Openbao client inside the custom SkeletonKey. The former reads
	// secrets from an Openbao server, and the latter offers convenient methods
	// to work with those secrets.
	secretsReader := new(secrets.SkeletonKey)
	secretsReader.Create(cfg)

	// Create a Redis client. The Openbao client reads x509 data from the
	// Openbao server, and the SkeletonKey assembles it into a working TLS
	// configuration.
	cache, cacheErr := cache.CreateClient(cfg, secretsReader)
	if cacheErr != nil {
		panic(cacheErr.Error())
	}
	defer cache.Close()

	// Create a child logger intended for the http.Server.
	srvLogger := logger.WithGroup("server")

	// Dependency wrapping happens here. Backbone holds pointers to a logger, a
	// Postgres connection pool, and a Redis client.
	backbone := router.NewBackbone(
		router.WithLogger(srvLogger),
		router.WithDbHandle(db1),
		router.WithCache(cache),
	)

	// In your executable, wrap the library Backbone with a native struct that
	// has its own HTTP Handlers. Wrap the wrapping. This secondary wrapper will
	// include a sqlC generated Query handle that can be accessed from the body
	// of a http.Handler, and ease querying Postgres.
	backboneWrapper := routes.WrapBackbone(backbone)

	// Feed the dependencies into a router. NewRouter will use the Gatherer
	// interface method GetEndpoints to add paths & handlers to a router.
	appRouter := router.NewRouter(cfg, backboneWrapper)

	// Read x509 certificate & key for this server to secure connections with
	// clients.
	X509, X509Err := secretsReader.ReadTlsCertAndKey(cfg.HttpServer.TlsServer)
	if X509Err != nil {
		logger.Error(X509Err.Error())
		panic(X509Err)
	}

	// Create a webserver with a router, an Error logger, and TLS configuration.
	webserver := server.NewServer(cfg, appRouter, X509, srvLogger)

	// Activate webserver gracefully, and await any termination signals. The
	// original logger will record any shutdown errors.
	server.Start(logger, webserver)
}
