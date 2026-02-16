package interceptors

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// CompressionMiddleware returns a middleware that compresses response bodies using gzip.
// It checks the Accept-Encoding header and applies compression if supported by the client.
func CompressionMiddleware() func(http.Handler) http.Handler {
	return middleware.Compress(5) // Default compression level
}
