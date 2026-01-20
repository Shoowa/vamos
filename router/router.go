// Package ROUTER provides a sanely configured router.
package router

import (
	"net/http"
)

// NewRouter accepts the any struct that conforms to the Gatherer interface. So
// it accepts the library Backbone struct, and any downstream executable struct
// that wraps around a Backbone. Ideally, dependencies are smuggled into the
// router, and this interface allows for the easy creation of a server in both
// production and testing.
func NewRouter(b Gatherer) *Bundle {
	mux := http.NewServeMux()

	addOperationalRoutes(mux, b)

	routerWithLoggingMiddleware := NewBundle(b.GetLogger(), mux)

	return routerWithLoggingMiddleware
}
