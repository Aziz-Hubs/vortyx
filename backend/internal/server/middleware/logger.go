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

// LoggerMiddleware returns a logger middleware that logs incoming HTTP requests.
// The logger uses the default chi logger which writes structured log entries
// including request ID, method, path, status, and latency.
//
// This middleware should be applied early in the middleware stack to capture
// all request information including downstream middleware execution time.
func LoggerMiddleware() func(http.Handler) http.Handler {
	return middleware.Logger
}
