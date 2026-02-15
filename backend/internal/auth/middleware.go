package auth

import (
	"context"
	"net/http"
	"strings"

	"connectrpc.com/connect"
)

// AuthMiddleware is a simple middleware that checks for a Bearer token.
// In a real implementation, this would validate the token against Zitadel.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// For now, we allow unauthenticated requests to pass through
			// because the frontend might not be sending tokens yet.
			// In production, you might want to return 401 here for protected routes.
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		// token := parts[1]
		// TODO: Validate token with Zitadel public keys

		// Mock: If token is "invalid", reject.
		if parts[1] == "invalid" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Interceptor for ConnectRPC
func NewAuthInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			// You can also check context metadata here if needed
			if req.Header().Get("Authorization") == "" {
				// return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing token"))
			}
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
