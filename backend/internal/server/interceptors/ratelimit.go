package interceptors

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitConfig defines configuration for rate limiting.
type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
}

// DefaultRateLimitConfig returns a default configuration.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 100,
		Burst:             200,
	}
}

// RateLimitMiddleware returns a middleware that limits the number of requests per client IP.
func RateLimitMiddleware(config RateLimitConfig) func(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Background routine to clean up old clients
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr // Ideally, use a more robust way to get IP (e.g., X-Forwarded-For)

			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst),
				}
			}
			clients[ip].lastSeen = time.Now()
			
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
