// Package middleware provides modular middleware components for the Vortyx server.
// It organizes middleware by concern: authentication, CORS, and standard utilities.
//
// The middleware is designed to be composable and configurable. Each middleware
// component can be used independently or combined using the Builder pattern.
package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORSConfig holds configuration for Cross-Origin Resource Sharing.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns the default CORS configuration.
// It allows requests from localhost:3000 (development frontend)
// and includes common HTTP methods and headers.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Connect-Protocol-Version"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}

// CORSMiddleware returns a CORS middleware handler with the provided configuration.
// CORS is required for browser-based clients to make cross-origin requests.
func CORSMiddleware(cfg CORSConfig) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   cfg.AllowedMethods,
		AllowedHeaders:   cfg.AllowedHeaders,
		ExposedHeaders:   cfg.ExposedHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
