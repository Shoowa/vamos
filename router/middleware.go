package router

import (
	"log/slog"
	"net/http"
)

// Bundle holds the logger & router, and satisfies the ServeHTTP interface. This
// enables the creation of a middleware for standardized logging.
type Bundle struct {
	Logger *slog.Logger
	Router http.Handler
}

func NewBundle(logger *slog.Logger, router http.Handler) *Bundle {
	return &Bundle{
		Logger: logger,
		Router: router,
	}
}

func (b *Bundle) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	b.Logger.Info(
		"Inbound",
		"method", req.Method,
		"path", req.URL.Path,
		"uagent", req.Header.Get("User-Agent"),
	)
	b.Router.ServeHTTP(w, req)
}
