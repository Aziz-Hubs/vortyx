// Package middleware provides modular middleware components for the Vortyx server.
// It organizes middleware by concern: authentication, CORS, and standard utilities.
//
// The middleware is designed to be composable and configurable. Each middleware
// component can be used independently or combined using the Builder pattern.
package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// RecovererMiddleware returns a middleware that recovers from panics and prevents
// the server from crashing. When a panic occurs, it returns a 500 Internal
// Server Error response instead of crashing the process.
//
// This middleware should be applied early in the stack to catch panics from
// any downstream handler or middleware. It logs the panic details for
// debugging purposes.
func RecovererMiddleware() func(http.Handler) http.Handler {
	return middleware.Recoverer
}
