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

// Bundle holds the logger & router, and satisfies the ServeHTTP interface. This
// enables the creation of a middleware for standardized logging.
type Bundle struct {
	Logger *slog.Logger
	Router *http.ServeMux
}

// NewBundle creates a custom struct that holds the standard router and a custom
// logger. The logger will be used in a middleware to record all incoming
// requests.
func NewBundle(logger *slog.Logger, router *http.ServeMux) *Bundle {
	return &Bundle{
		Logger: logger,
		Router: router,
	}
}

// ServeHTTP method on the Bundle struct conforms to the Handler interface, and
// is added to the router as middleware that can record every incoming HTTP
// request, and record the number of active TCP connections in a gauge.
func (b *Bundle) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	metrics.HttpRequestsGauge.Inc()
	defer metrics.HttpRequestsGauge.Dec()
	b.Logger.Info(
		"Inbound",
		"method", req.Method,
		"path", req.URL.Path,
		"uagent", req.Header.Get("User-Agent"),
	)

	recorder := &statusRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	b.Router.ServeHTTP(recorder, req)

	if recorder.statusCode == 404 {
		metrics.HttpRequestCounter.WithLabelValues("404", "invalid path", req.Method).Inc()
	} else {
		status := strconv.Itoa(recorder.statusCode)
		metrics.HttpRequestCounter.WithLabelValues(status, req.URL.Path, req.Method).Inc()
	}
}

// Endpoint is a custom struct that can be used to create a menu of routes.
// Ideally, viewing a populated Endpoint in a file is easy on the eyes during
// code review. This type is essential to use the Bundle struct's AddRoutes
// method.
type Endpoint struct {
	VerbAndPath string
	Handler     errHandler
}

// AddRoutes is a convenient method designed to read from a list of Endpoints
// and add them to the http.ServeMux residing inside the Bundle struct. This
// method also enforces that HTTP Handlers written by a downstream user must
// return an error to conform to the errHandler type.
func (b *Bundle) AddRoutes(deps Gatherer) {
	routeMenu := deps.GetEndpoints()
	for _, endpoint := range routeMenu {
		b.Router.HandleFunc(endpoint.VerbAndPath, deps.eHand(endpoint.Handler))
	}
}
