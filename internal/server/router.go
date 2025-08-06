package server

import (
	"net/http"
)

// NewRouter accepts the Backbone struct containing all the dependencies needed
// by the Middleware and the Routes.
func NewRouter(b *Backbone) *Bundle {
	mux := http.NewServeMux()

	// An Operations team can amend routes in routes_operations.go
	addOperationalRoutes(mux, b)

	// A Development team can amend routes_features_v1.go
	addFeaturesV1(mux, b)

	routerWithLoggingMiddleware := NewBundle(b.Logger, mux)

	return routerWithLoggingMiddleware
}
