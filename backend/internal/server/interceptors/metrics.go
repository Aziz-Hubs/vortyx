package interceptors

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

// MetricsMiddleware returns a middleware that records Prometheus metrics for HTTP requests.
func MetricsMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				status := strconv.Itoa(ww.Status())
				duration := time.Since(start).Seconds()

				httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
				httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
