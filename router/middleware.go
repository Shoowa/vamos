package router

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Shoowa/vamos/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader conforms to an interface so that a library struct can be inserted
// into the standatd Router, copy the served HTTP status code, then read the
// copied status code after a response is served.
func (recorder *statusRecorder) WriteHeader(code int) {
	recorder.statusCode = code
	recorder.ResponseWriter.WriteHeader(code)
}

func recordResponses(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		if recorder.statusCode == 404 {
			metrics.HttpRequestCounter.WithLabelValues("404", "invalid path", r.Method).Inc()
		} else {
			status := strconv.Itoa(recorder.statusCode)
			metrics.HttpRequestCounter.WithLabelValues(status, r.URL.Path, r.Method).Inc()
		}
	})
}

func logRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info(
			"Inbound",
			"method", r.Method,
			"path", r.URL.Path,
			"uagent", r.Header.Get("User-Agent"),
		)

		next.ServeHTTP(w, r)
	})
}

func gaugeRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.HttpRequestsGauge.Inc()
		defer metrics.HttpRequestsGauge.Dec()

		next.ServeHTTP(w, r)
	})
}

// Endpoint is a custom struct that can be used to create a menu of routes.
// Ideally, viewing a populated Endpoint in a file is easy on the eyes during
// code review.
type Endpoint struct {
	VerbAndPath string
	Handler     http.HandlerFunc
}
