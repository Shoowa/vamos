// Package SERVER provides a sanely configured web server.
package server

import (
	"net/http"
	"time"

	"vamos/internal/config"
)

// NewServer creates a custom http.Server struct. And passes the dependencies in
// Backbone to the routing layer, so that HTTP Handlers can access the databases
// and cache.
func NewServer(cfg *config.Config, b *Backbone) *http.Server {
	return &http.Server{
		Addr:         cfg.HttpServer.Port,
		Handler:      NewRouter(b),
		ReadTimeout:  time.Second * time.Duration(cfg.HttpServer.TimeoutRead),
		WriteTimeout: time.Second * time.Duration(cfg.HttpServer.TimeoutWrite),
		IdleTimeout:  time.Second * time.Duration(cfg.HttpServer.TimeoutIdle),
	}
}
