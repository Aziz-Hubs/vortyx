package interceptors

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestIDMiddleware injects a request ID into the context of each request.
// It retrieves the ID from the X-Request-Id header or generates a new one.
// The request ID is then available in the context and included in the response headers.
func RequestIDMiddleware() func(http.Handler) http.Handler {
	return middleware.RequestID
}
