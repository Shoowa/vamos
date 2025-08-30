// Package ROUTER provides a sanely configured router.
package router

import (
	"net/http"
)

// NewRouter accepts the Backbone struct containing all the dependencies needed
// by the Middleware and the Routes.
func NewRouter(b HttpErrorHandler) *Bundle {
	mux := http.NewServeMux()

	// An Operations team can amend routes in routes_operations.go
	addOperationalRoutes(mux, b)

	routerWithLoggingMiddleware := NewBundle(b.GetLogger(), mux)

	return routerWithLoggingMiddleware
}
