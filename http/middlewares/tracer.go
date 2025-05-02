package middlewares

import (
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
)

// Tracer creates an initial span
func Tracer() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return tracingHandler(h)
	}
}

func tracingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span, ctx := tracer.StartSpanFromContext(r.Context(), "proxy")
		defer span.Finish()
		w.Header().Add("x-trace-id", fmt.Sprintf("%d", span.Context().TraceID()))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
