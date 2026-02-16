package interceptors

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware returns a middleware that adds OpenTelemetry tracing to requests.
func TracingMiddleware(serviceName string) func(next http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)
	propagator := otel.GetTextMapPropagator()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
			ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			// Inject trace context into response headers (optional but useful for debugging)
			// propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
