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
func NewRouter(cfg *config.HttpServer, b Gatherer) http.Handler {
	mux := http.NewServeMux()

	// Add health check.
	addOperationalRoutes(mux, b)

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
	corfMW := preventCORF(cfg.CheckCORF, gaugingMW)
	finalMW := optionalGlobalRateLimiter(cfg.GlobalRateLimiter, corfMW)
	return finalMW
}
