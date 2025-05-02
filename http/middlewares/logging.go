package middlewares

import (
	"github.com/sirupsen/logrus"
	"net/http"
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
		newW := NewLoggingResponseWriter(w)
		next.ServeHTTP(newW, r)
		logrus.WithField("status_code", newW.statusCode).WithField("url", r.URL.Path).Info("proxying request completed")
	})
}
