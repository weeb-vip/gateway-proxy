package metrics

import (
	"github.com/weeb-vip/gateway-proxy/config"
	metricsLib "github.com/weeb-vip/go-metrics-lib"
)

// AppMetrics provides a centralized metrics interface with default tags
type AppMetrics struct {
	metricsImpl metricsLib.MetricsImpl
	defaultTags map[string]string
}

// Global instance
var appMetrics *AppMetrics

// GetAppMetrics returns the singleton metrics instance
func GetAppMetrics() *AppMetrics {
	if appMetrics == nil {
		cfg := config.LoadConfigOrPanic()

		// Initialize with Prometheus metrics
		impl := NewMetricsInstance()

		appMetrics = &AppMetrics{
			metricsImpl: impl,
			defaultTags: map[string]string{
				"service": cfg.APPConfig.APPName,
				"env":     cfg.APPConfig.Env,
				"version": cfg.APPConfig.Version,
			},
		}
	}
	return appMetrics
}

// ProxyMetric records proxy request performance metrics
func (m *AppMetrics) ProxyMetric(duration float64, method string, statusCode string, result string) {
	labels := metricsLib.DatabaseMetricLabels{
		Service: m.defaultTags["service"],
		Table:   "proxy_request",
		Method:  method,
		Result:  result,
		Env:     m.defaultTags["env"],
	}
	m.metricsImpl.DatabaseMetric(duration, labels)
}

// JWTValidationMetric records JWT validation performance metrics
func (m *AppMetrics) JWTValidationMetric(duration float64, result string) {
	labels := metricsLib.DatabaseMetricLabels{
		Service: m.defaultTags["service"],
		Table:   "jwt_validation",
		Method:  "validate",
		Result:  result,
		Env:     m.defaultTags["env"],
	}
	m.metricsImpl.DatabaseMetric(duration, labels)
}

// KeyFetchMetric records key fetch performance metrics
func (m *AppMetrics) KeyFetchMetric(duration float64, result string) {
	labels := metricsLib.DatabaseMetricLabels{
		Service: m.defaultTags["service"],
		Table:   "key_fetch",
		Method:  "fetch",
		Result:  result,
		Env:     m.defaultTags["env"],
	}
	m.metricsImpl.DatabaseMetric(duration, labels)
}

// GetDefaultTags returns the default tags for this metrics instance
func (m *AppMetrics) GetDefaultTags() map[string]string {
	// Return a copy to prevent modification
	tags := make(map[string]string)
	for k, v := range m.defaultTags {
		tags[k] = v
	}
	return tags
}

// RequestMetric records general HTTP request metrics with status code
func (m *AppMetrics) RequestMetric(duration float64, method string, statusCode string, result string) {
	labels := metricsLib.DatabaseMetricLabels{
		Service: m.defaultTags["service"],
		Table:   "http_request",
		Method:  method,
		Result:  result,
		Env:     m.defaultTags["env"],
	}
	// Add status code as part of the method for more granular tracking
	labels.Method = method + "_" + statusCode
	m.metricsImpl.DatabaseMetric(duration, labels)
}

// ErrorMetric records error metrics by type
func (m *AppMetrics) ErrorMetric(errorType string) {
	labels := metricsLib.DatabaseMetricLabels{
		Service: m.defaultTags["service"],
		Table:   "errors",
		Method:  errorType,
		Result:  "error",
		Env:     m.defaultTags["env"],
	}
	m.metricsImpl.DatabaseMetric(1, labels) // Count errors as 1ms duration
}

// CacheMetric records cache hit/miss performance metrics
func (m *AppMetrics) CacheMetric(duration float64, result string) {
	prometheusClient := NewPrometheusInstance()
	labels := map[string]string{
		"service": m.defaultTags["service"],
		"method":  "lookup",
		"result":  result,
		"env":     m.defaultTags["env"],
	}
	prometheusClient.Histogram("cache_operation_duration_histogram_milliseconds", duration, labels, 1.0)
}

// CacheCounterMetric records cache hit/miss counters
func (m *AppMetrics) CacheCounterMetric(result string) {
	prometheusClient := NewPrometheusInstance()
	labels := map[string]string{
		"service": m.defaultTags["service"],
		"method":  "count",
		"result":  result,
		"env":     m.defaultTags["env"],
	}
	prometheusClient.Histogram("cache_counter_histogram", 1, labels, 1.0)
}

// WithTags returns a new metrics instance with additional tags
func (m *AppMetrics) WithTags(additionalTags map[string]string) *AppMetrics {
	newTags := make(map[string]string)

	// Copy default tags
	for k, v := range m.defaultTags {
		newTags[k] = v
	}

	// Add additional tags (will override defaults if same key)
	for k, v := range additionalTags {
		newTags[k] = v
	}

	return &AppMetrics{
		metricsImpl: m.metricsImpl,
		defaultTags: newTags,
	}
}