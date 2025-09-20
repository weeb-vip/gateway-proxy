package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/weeb-vip/gateway-proxy/config"
)

// CORS creates a CORS middleware with configuration
func CORS(cfg *config.Config) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return corsHandler(h, cfg)
	}
}

func corsHandler(next http.Handler, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			setCORSHeaders(w, origin, cfg)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Create a response writer that intercepts headers
		corsWriter := &corsResponseWriter{
			ResponseWriter: w,
			origin:         origin,
			cfg:            cfg,
		}

		next.ServeHTTP(corsWriter, r)
	})
}

type corsResponseWriter struct {
	http.ResponseWriter
	origin      string
	cfg         *config.Config
	headersSent bool
}

func (c *corsResponseWriter) WriteHeader(statusCode int) {
	if !c.headersSent {
		// Remove any existing CORS headers from backend
		c.Header().Del("Access-Control-Allow-Origin")
		c.Header().Del("Access-Control-Allow-Credentials")
		c.Header().Del("Access-Control-Allow-Methods")
		c.Header().Del("Access-Control-Allow-Headers")
		c.Header().Del("Access-Control-Max-Age")

		// Set our CORS headers
		setCORSHeaders(c.ResponseWriter, c.origin, c.cfg)
		c.headersSent = true
	}

	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *corsResponseWriter) Write(data []byte) (int, error) {
	if !c.headersSent {
		c.WriteHeader(http.StatusOK)
	}
	return c.ResponseWriter.Write(data)
}

func setCORSHeaders(w http.ResponseWriter, origin string, cfg *config.Config) {
	// Check if origin is allowed
	if isOriginAllowed(origin, cfg.CORSAllowedOrigins) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	if cfg.CORSAllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

	if cfg.CORSMaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.CORSMaxAge))
	}
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// Support wildcard subdomains (e.g., "*.example.com")
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}
