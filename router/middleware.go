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

func (recorder *statusRecorder) WriteHeader(code int) {
	recorder.statusCode = code
	recorder.ResponseWriter.WriteHeader(code)
}

// Bundle holds the logger & router, and satisfies the ServeHTTP interface. This
// enables the creation of a middleware for standardized logging.
type Bundle struct {
	Logger *slog.Logger
	Router *http.ServeMux
}

func NewBundle(logger *slog.Logger, router *http.ServeMux) *Bundle {
	return &Bundle{
		Logger: logger,
		Router: router,
	}
}

func (b *Bundle) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	metrics.HttpRequestsGauge.Inc()
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
	metrics.HttpRequestsGauge.Dec()

	status := strconv.Itoa(recorder.statusCode)
	metrics.HttpRequestCounter.WithLabelValues(status, req.URL.Path, req.Method).Inc()
}

// Endpoint is a struct that can be used to create a menu of routes.
type Endpoint struct {
	VerbAndPath string
	Handler     errHandler
}

// AddRoutes is a convenient method designed to read from a list of Endpoints
// and add them to the http.ServeMux residing inside the Bundle struct. This
// method also enforces that HTTP Handlers written by a downstream user must
// return an error to conform to the errHandler type.
func (b *Bundle) AddRoutes(routeMenu []Endpoint, deps *Backbone) {
	for _, endpoint := range routeMenu {
		b.Router.HandleFunc(endpoint.VerbAndPath, deps.eHand(endpoint.Handler))
	}
}
