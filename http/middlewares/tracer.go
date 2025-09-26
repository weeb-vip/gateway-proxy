package middlewares

import (
	"fmt"
	"net/http"

	"github.com/weeb-vip/gateway-proxy/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Tracer creates OpenTelemetry tracing spans for HTTP requests
func Tracer() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return tracingHandler(h)
	}
}

func tracingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace context from incoming request headers
		propagator := propagation.TraceContext{}
		ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		tracer := tracing.GetTracer(ctx)

		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		tracedCtx, span := tracer.Start(ctx, spanName,
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
				attribute.String("http.route", r.URL.Path),
			),
			tracing.GetEnvironmentAttribute(),
		)
		defer span.End()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Inject trace context into response headers for downstream services
		propagator.Inject(tracedCtx, propagation.HeaderCarrier(rw.Header()))

		// Call the next handler with the traced context
		next.ServeHTTP(rw, r.WithContext(tracedCtx))

		// Add response attributes to span
		span.SetAttributes(
			attribute.Int("http.status_code", rw.statusCode),
		)

		// Set span status based on HTTP status code
		if rw.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", rw.statusCode))
		} else {
			span.SetStatus(codes.Ok, "")
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
