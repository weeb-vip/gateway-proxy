package handlers

import (
	"fmt"
	"github.com/weeb-vip/gateway-proxy/config"
	"github.com/weeb-vip/gateway-proxy/internal/jwt"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
	"net/http/httputil"
	"strings"
)

func GetProxy(config *config.Config, jwtParser jwt.Parser) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = config.ProxyURL.Host
		addUserAgentHeader(request, config)
		addRemoteIP(request)
		addJWTData(request, jwtParser, config.AuthMode)
		addTraceHeaders(request)
		// log all headers
		for name, headers := range request.Header {
			for _, h := range headers {
				fmt.Printf("Header: %v, Value: %v\n", name, h)
			}
		}
		if config.OverrideOrigin != nil && request.Header.Get("Origin") != *config.OverrideOrigin {
			request.Header.Set("Origin", *config.OverrideOrigin)
		}
	}}
}

func addJWTData(request *http.Request, parser jwt.Parser, authMode string) {
	var token string

	switch authMode {
	case "cookie":
		// Only check cookie
		accessTokenCookie, err := request.Cookie("access_token")
		if err != nil {
			return
		}
		token = accessTokenCookie.Value
	case "header":
		// Only check Authorization header
		authorizationHeader := request.Header.Get("Authorization")
		if authorizationHeader != "" && strings.HasPrefix(authorizationHeader, "Bearer ") {
			token = authorizationHeader[len("Bearer "):]
		}
	default: // "both" or any other value defaults to both
		// First, try to get token from Authorization header
		authorizationHeader := request.Header.Get("Authorization")
		if authorizationHeader != "" && strings.HasPrefix(authorizationHeader, "Bearer ") {
			token = authorizationHeader[len("Bearer "):]
		} else {
			// Fallback to cookie if Authorization header is not present
			accessTokenCookie, err := request.Cookie("access_token")
			if err != nil {
				return
			}
			token = accessTokenCookie.Value
		}
	}

	if token == "" {
		return
	}

	info, err := parser.Parse(token)
	if err != nil {
		return
	}
	if info.Subject != nil {
		request.Header.Set("x-user-id", *info.Subject)
	}
	if info.Purpose != nil {
		request.Header.Set("x-token-purpose", *info.Purpose)
	}
	request.Header.Add("x-raw-token", token)
}

func addRemoteIP(request *http.Request) {
	request.Header.Set("x-remote-ip", request.Header.Get("x-forwarded-for"))
	request.Header.Del("x-forwarded-for")
}

func addTraceHeaders(request *http.Request) {
	// Inject trace context into outgoing request headers
	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	propagator.Inject(request.Context(), propagation.HeaderCarrier(request.Header))
}

func addUserAgentHeader(request *http.Request, cfg *config.Config) {
	request.Header.Set("x-user-agent", request.UserAgent())
	request.Header.Del("User-Agent")
	request.Header.Set("User-Agent", fmt.Sprintf("reverse-proxy/%s", cfg.Version))
}
