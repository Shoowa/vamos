// Package ROUTER provides a sanely configured router.
package router

import (
	"net/http"

	"github.com/Shoowa/vamos/config"
)

// NewRouter accepts the any struct that conforms to the Gatherer interface. So
// it accepts the library Backbone struct, and any downstream executable struct
// that wraps around a Backbone. Ideally, dependencies are smuggled into the
// router, and this interface allows for the easy creation of a server in both
// production and testing.
func NewRouter(cfg *config.Config, b Gatherer) http.Handler {
	health := setupHealthChecks(cfg, b.GetBackbone())
	mux := http.NewServeMux()

	// Create a fileserver to offer static files.
	if cfg.HttpServer.StaticDir != "" {
		staticDir := http.Dir(cfg.HttpServer.StaticDir)
		fileServer := http.FileServer(staticDir)
		mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	}

	// Add health check.
	addOperationalRoutes(mux, health, b.GetLogger())

	// Conveniently add routes.
	endpoints := b.GetEndpoints()
	for _, endpoint := range endpoints {
		mux.HandleFunc(endpoint.VerbAndPath, endpoint.Handler)
	}

	// Add mandatory middleware.
	responseRecordingMW := recordResponses(mux)
	loggingMW := logRequests(b.GetLogger(), responseRecordingMW)
	gaugingMW := gaugeRequests(loggingMW)

	// Add optional middleware or stop at gaugeMW.
	corfMW := preventCORF(cfg.HttpServer.CheckCORF, gaugingMW)
	finalMW := optionalGlobalRateLimiter(cfg.HttpServer.GlobalRateLimiter, corfMW)
	return finalMW
}
