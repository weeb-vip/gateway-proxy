package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/weeb-vip/gateway-proxy/config"
	"github.com/weeb-vip/gateway-proxy/internal/cache"
	"github.com/weeb-vip/gateway-proxy/internal/logger"
	"github.com/weeb-vip/gateway-proxy/metrics"
	"github.com/weeb-vip/gateway-proxy/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type cacheResponseWriter struct {
	http.ResponseWriter
	body    *bytes.Buffer
	headers map[string][]string
	status  int
}

func (rw *cacheResponseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func (rw *cacheResponseWriter) WriteHeader(status int) {
	rw.status = status
	// Copy headers before writing
	for k, v := range rw.ResponseWriter.Header() {
		rw.headers[k] = v
	}
	rw.ResponseWriter.WriteHeader(status)
}

func GraphQLCacheMiddleware(cache *cache.GraphQLCache, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.Get()
			metricsClient := metrics.GetAppMetrics()

			// Start timing for cache lookup
			startTime := time.Now()

			// Only cache POST requests (GraphQL mutations/queries)
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}

			// Check Content-Type for GraphQL
			contentType := r.Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "application/graphql") {
				next.ServeHTTP(w, r)
				return
			}

			// Extract user token from access_token cookie
			var userToken string
			if accessTokenCookie, err := r.Cookie("access_token"); err == nil {
				userToken = accessTokenCookie.Value
			}

			// If no user token, don't cache (could be public queries)
			if userToken == "" {
				log.Debug().Msg("No user token found, skipping cache")
				metricsClient.CacheCounterMetric("skip_no_token")
				next.ServeHTTP(w, r)
				return
			}

			// Read request body
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				log.Error().Err(err).Msg("Failed to read request body")
				next.ServeHTTP(w, r)
				return
			}

			// Restore request body for downstream handlers
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			// Generate cache key
			cacheKey := cache.GenerateKey(userToken, string(bodyBytes))

			// Check if response is cached with tracing
			tracer := tracing.GetTracer(r.Context())
			ctx, span := tracer.Start(r.Context(), "cache.lookup",
				trace.WithAttributes(
					attribute.String("cache.key", cacheKey),
					attribute.String("cache.operation", "get"),
				),
			)
			defer span.End()

			if entry, found := cache.Get(cacheKey); found {
				cacheLookupDuration := time.Since(startTime).Seconds()
				metricsClient.CacheMetric(cacheLookupDuration, "hit")
				metricsClient.CacheCounterMetric("hit")

				span.SetAttributes(
					attribute.String("cache.result", "hit"),
					attribute.String("cache.age", time.Since(entry.Timestamp).String()),
				)

				log.Debug().
					Str("cache_key", cacheKey).
					Msg("Cache hit - serving cached response")

				// Set cached headers
				for k, v := range entry.Headers {
					for _, header := range v {
						w.Header().Add(k, header)
					}
				}

				// Add cache status header
				w.Header().Set("X-Cache-Status", "HIT")
				w.Header().Set("X-Cache-Age", time.Since(entry.Timestamp).String())

				// Write cached response
				w.WriteHeader(http.StatusOK)
				w.Write(entry.Response)
				return
			}

			span.SetAttributes(attribute.String("cache.result", "miss"))
			_ = ctx // Use ctx if needed elsewhere

			// Cache miss - capture response
			cacheLookupDuration := time.Since(startTime).Seconds()
			metricsClient.CacheMetric(cacheLookupDuration, "miss")
			metricsClient.CacheCounterMetric("miss")

			log.Debug().
				Str("cache_key", cacheKey).
				Msg("Cache miss - executing request")

			rw := &cacheResponseWriter{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				headers:        make(map[string][]string),
				status:         http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			// Only cache successful responses (2xx status codes)
			if rw.status >= 200 && rw.status < 300 {
				// Check if response is cacheable (avoid caching mutations)
				requestBody := string(bodyBytes)
				if !isCacheableRequest(requestBody) {
					log.Debug().
						Str("cache_key", cacheKey).
						Msg("Request not cacheable (likely mutation)")
					metricsClient.CacheCounterMetric("skip_mutation")
					w.Header().Set("X-Cache-Status", "SKIP")
					return
				}

				cache.Set(cacheKey, rw.body.Bytes(), rw.headers)
				metricsClient.CacheCounterMetric("set")
				log.Debug().
					Str("cache_key", cacheKey).
					Msg("Response cached")
			}

			w.Header().Set("X-Cache-Status", "MISS")
		})
	}
}

// isCacheableRequest determines if a GraphQL request should be cached
// Generally, we only want to cache queries, not mutations
func isCacheableRequest(requestBody string) bool {
	// Convert to lowercase for case-insensitive matching
	body := strings.ToLower(requestBody)

	// If it contains "mutation", it's likely not cacheable
	if strings.Contains(body, "mutation") {
		return false
	}

	// If it contains "query" or looks like a query, it's probably cacheable
	if strings.Contains(body, "query") || strings.Contains(body, "{") {
		return true
	}

	// Default to not caching if we can't determine
	return false
}