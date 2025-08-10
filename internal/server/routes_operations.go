package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// addOperationalRoutes adds health checks and metrics to the router.
func addOperationalRoutes(router *http.ServeMux, b *Backbone) {

	healthCheck := http.HandlerFunc(b.Healthcheck)
	router.HandleFunc("GET /health", healthCheck)

	router.Handle("GET /metrics", promhttp.Handler())
}

func (b *Backbone) Healthcheck(w http.ResponseWriter, r *http.Request) {
	status := b.Health.PassFail()

	if status {
		w.WriteHeader(http.StatusNoContent)
	} else {
		b.Logger.Error("Failed health check")
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
