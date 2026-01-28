package router

import (
	"log/slog"
	"net/http"

	"github.com/Shoowa/vamos/metrics"
)

// addOperationalRoutes adds health checks and metrics to the router.
func addOperationalRoutes(router *http.ServeMux, health *Health, logger *slog.Logger) {
	healthCheck := http.HandlerFunc(assessHealth(health, logger))
	router.HandleFunc("GET /health", healthCheck)

	metricsHandler := metrics.CreateHandler()
	router.Handle("GET /metrics", metricsHandler)
}

func assessHealth(health *Health, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := health.PassFail()

		if status {
			w.WriteHeader(http.StatusNoContent)
		} else {
			logger.Error("Failed health check")
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
}
