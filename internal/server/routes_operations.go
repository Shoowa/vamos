package server

import (
	"net/http"
)

// addOperationalRoutes adds health checks and metrics to the router.
func addOperationalRoutes(router *http.ServeMux, b *Backbone) {

	healthCheck := http.HandlerFunc(b.health)
	router.HandleFunc("GET /health", healthCheck)
}

func (b *Backbone) health(w http.ResponseWriter, r *http.Request) {
	status := b.Health.PassFail()

	if status {
		w.WriteHeader(http.StatusNoContent)
	} else {
		b.Logger.Error("Failed health check")
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
