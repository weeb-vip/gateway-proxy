package middlewares

import (
	"net/http"
	"time"

	"github.com/weeb-vip/gateway-proxy/internal/logger"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// Logger logs request URL and response code
func Logger() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return loggingHandler(h)
	}
}

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		newW := NewLoggingResponseWriter(w)

		next.ServeHTTP(newW, r)

		duration := time.Since(start)
		log := logger.FromCtx(r.Context())

		log.Info().
			Str("method", r.Method).
			Str("url", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Int("status_code", newW.statusCode).
			Dur("duration", duration).
			Msg("proxying request completed")
	})
}
