package server

import (
	"net/http"
)

// NewRouter accepts the Backbone struct containing all the dependencies needed
// by the Middleware and the Routes.
func NewRouter(b *Backbone) *Bundle {
	mux := http.NewServeMux()

	routerWithLoggingMiddleware := NewBundle(b.Logger, mux)

	return routerWithLoggingMiddleware
}
