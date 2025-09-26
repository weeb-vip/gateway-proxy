package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/weeb-vip/gateway-proxy/metrics"
)

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newMetricsResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{w, http.StatusOK}
}

func (mrw *metricsResponseWriter) WriteHeader(code int) {
	mrw.statusCode = code
	mrw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware tracks HTTP request metrics
func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return metricsHandler(h)
	}
}

func metricsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		mrw := newMetricsResponseWriter(w)

		next.ServeHTTP(mrw, r)

		duration := float64(time.Since(start).Nanoseconds()) / float64(time.Millisecond)
		statusCode := strconv.Itoa(mrw.statusCode)
		result := "success"
		if mrw.statusCode >= 400 {
			result = "error"
		}

		appMetrics := metrics.GetAppMetrics()
		// Record both general request metrics and proxy-specific metrics
		appMetrics.RequestMetric(duration, r.Method, statusCode, result)
		appMetrics.ProxyMetric(duration, r.Method, statusCode, result)
	})
}
