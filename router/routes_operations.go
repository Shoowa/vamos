package router

import (
	"net/http"

	"github.com/Shoowa/vamos/metrics"
)

// addOperationalRoutes adds health checks and metrics to the router.
func addOperationalRoutes(router *http.ServeMux, heh Gatherer) {
	b := heh.GetBackbone()
	healthCheck := http.HandlerFunc(b.Healthcheck)
	router.HandleFunc("GET /health", healthCheck)

	metricsHandler := metrics.CreateHandler()
	router.Handle("GET /metrics", metricsHandler)
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
