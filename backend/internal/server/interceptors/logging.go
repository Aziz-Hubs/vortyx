package interceptors

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// LoggerMiddleware returns a middleware that logs HTTP requests using structured logging (slog).
// It logs the request method, path, status, duration, and other details.
func LoggerMiddleware() func(next http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Wrap the response writer to capture status code and bytes written
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				// Determine log level based on status code
				status := ww.Status()
				level := slog.LevelInfo
				if status >= 500 {
					level = slog.LevelError
				} else if status >= 400 {
					level = slog.LevelWarn
				}

				logger.Log(r.Context(), level, "http request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", status,
					"duration", time.Since(start).String(),
					"bytes", ww.BytesWritten(),
					"remote_addr", r.RemoteAddr,
					"user_agent", r.UserAgent(),
					"request_id", middleware.GetReqID(r.Context()),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
