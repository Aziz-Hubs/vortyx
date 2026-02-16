package interceptors

import (
	"log/slog"
	"net/http"
	"os"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// AuditMiddleware returns a middleware that logs authenticated user actions.
// It should be placed AFTER the authentication middleware.
func AuditMiddleware() func(http.Handler) http.Handler {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user info from context
			userID := GetUserID(r.Context())
			orgID := GetOrganizationID(r.Context())

			// Only log if we have an authenticated user
			if userID != "" {
				logger.Info("audit log",
					"type", "access",
					"user_id", userID,
					"org_id", orgID,
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", chiMiddleware.GetReqID(r.Context()),
				)
			}

			next.ServeHTTP(w, r)
		})
	}
}
