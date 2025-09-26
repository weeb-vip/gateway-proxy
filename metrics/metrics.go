package metrics

import (
	metricsLib "github.com/weeb-vip/go-metrics-lib"
	"github.com/weeb-vip/go-metrics-lib/clients/prometheus"
	"github.com/weeb-vip/gateway-proxy/config"
)

var metricsInstance metricsLib.MetricsImpl

var prometheusInstance *prometheus.PrometheusClient

func NewMetricsInstance() metricsLib.MetricsImpl {
	if metricsInstance == nil {
		prometheusInstance = NewPrometheusInstance()
		initMetrics(prometheusInstance)
		metricsInstance = metricsLib.NewMetrics(prometheusInstance, 1)
	}
	return metricsInstance
}

func NewPrometheusInstance() *prometheus.PrometheusClient {
	if prometheusInstance == nil {
		prometheusInstance = prometheus.NewPrometheusClient()
		initMetrics(prometheusInstance)
	}
	return prometheusInstance
}

func initMetrics(prometheusInstance *prometheus.PrometheusClient) {
	// Proxy request duration metrics
	prometheusInstance.CreateHistogramVec("proxy_request_duration_histogram_milliseconds_buckets", "proxy request millisecond", []string{"service", "protocol", "method", "status_code", "result", "env"}, []float64{
		100,
		200,
		300,
		400,
		500,
		600,
		700,
		800,
		900,
		1000,
		2000,
		5000,
	})

	// General HTTP request duration metrics
	prometheusInstance.CreateHistogramVec("http_request_duration_histogram_milliseconds", "HTTP request millisecond", []string{"service", "method", "status_code", "result", "env"}, []float64{
		10,
		25,
		50,
		100,
		200,
		500,
		1000,
		2000,
		5000,
	})

	// JWT validation duration metrics
	prometheusInstance.CreateHistogramVec("jwt_validation_duration_histogram_milliseconds", "JWT validation millisecond", []string{"service", "result", "env"}, []float64{
		10,
		25,
		50,
		100,
		250,
		500,
		1000,
	})

	// Key fetch duration metrics
	prometheusInstance.CreateHistogramVec("key_fetch_duration_histogram_milliseconds", "key fetch millisecond", []string{"service", "result", "env"}, []float64{
		100,
		200,
		500,
		1000,
		2000,
		5000,
	})

	// Error count metrics
	prometheusInstance.CreateHistogramVec("error_count_histogram", "error count", []string{"service", "error_type", "env"}, []float64{
		1,
	})

	// Cache operation duration metrics
	prometheusInstance.CreateHistogramVec("cache_operation_duration_histogram_milliseconds", "cache operation millisecond", []string{"service", "method", "result", "env"}, []float64{
		1,
		5,
		10,
		25,
		50,
		100,
		250,
		500,
	})

	// Cache counter metrics
	prometheusInstance.CreateHistogramVec("cache_counter_histogram", "cache counter", []string{"service", "method", "result", "env"}, []float64{
		1,
	})
}

func GetCurrentEnv() string {
	cfg := config.LoadConfigOrPanic()
	return cfg.APPConfig.Env
}