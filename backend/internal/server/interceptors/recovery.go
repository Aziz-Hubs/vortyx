package interceptors

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
)

// RecovererMiddleware returns a middleware that recovers from panics, logs the stack trace,
// and returns a JSON 500 Internal Server Error response.
func RecovererMiddleware() func(next http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					// Log the panic with stack trace
					logger.Error("panic recovered",
						"error", fmt.Sprintf("%v", rvr),
						"stack", string(debug.Stack()),
						"path", r.URL.Path,
						"method", r.Method,
					)

					// Return JSON 500 response
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Internal Server Error",
						"message": "An unexpected error occurred",
					})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
