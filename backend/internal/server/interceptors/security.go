package interceptors

import (
	"net/http"
)

// SecurityConfig defines configuration for security headers.
type SecurityConfig struct {
	HSTSMaxAge            int
	ContentSecurityPolicy string
	FrameOptions          string
	ContentTypeOptions    string
	ReferrerPolicy        string
}

// DefaultSecurityConfig returns a safe default configuration.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		HSTSMaxAge:            31536000, // 1 year
		ContentSecurityPolicy: "default-src 'self'",
		FrameOptions:          "DENY",
		ContentTypeOptions:    "nosniff",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}
}

// SecurityMiddleware returns a middleware that adds security headers to responses.
func SecurityMiddleware(config SecurityConfig) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// HSTS
			if r.TLS != nil {
				w.Header().Set("Strict-Transport-Security", "max-age="+string(rune(config.HSTSMaxAge))+"; includeSubDomains")
			}

			// X-Frame-Options
			if config.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.FrameOptions)
			}

			// X-Content-Type-Options
			if config.ContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.ContentTypeOptions)
			}

			// Content-Security-Policy
			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// Referrer-Policy
			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			next.ServeHTTP(w, r)
		})
	}
}
