package interceptors

import (
	"net/http"
	"time"
)

// NoCacheMiddleware sets headers to prevent client-side caching.
// This is recommended for API endpoints returning dynamic or sensitive data.
func NoCacheMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			next.ServeHTTP(w, r)
		})
	}
}

// CacheControlMiddleware sets the Cache-Control header with the given max-age.
func CacheControlMiddleware(maxAge time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age="+string(rune(maxAge.Seconds())))
			next.ServeHTTP(w, r)
		})
	}
}
