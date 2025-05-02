package middlewares

import (
	"github.com/weeb-vip/gateway-proxy/internal/measurements"
	"net/http"
)

// Measurement tracks metrics about query
func Measurement(metrics measurements.Measurer) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return measurementHandler(metrics, h)
	}
}

func measurementHandler(metrics measurements.Measurer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer metrics.MeasureExecutionTime("request.time", []string{})()
		next.ServeHTTP(w, r)
	})
}
