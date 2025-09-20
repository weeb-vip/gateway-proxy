package handlers

import (
	"fmt"
	"github.com/weeb-vip/gateway-proxy/config"
	"github.com/weeb-vip/gateway-proxy/internal/jwt"
	"net/http"
	"net/http/httputil"
)

func GetProxy(config *config.Config, jwtParser jwt.Parser) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{Director: func(request *http.Request) {
		request.URL.Scheme = "http"
		request.URL.Host = config.ProxyURL.Host
		addUserAgentHeader(request, config)
		addRemoteIP(request)
		addJWTData(request, jwtParser)
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

func addJWTData(request *http.Request, parser jwt.Parser) {
	accessTokenCookie, err := request.Cookie("access_token")
	if err != nil {
		return
	}
	token := accessTokenCookie.Value
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

func addUserAgentHeader(request *http.Request, cfg *config.Config) {
	request.Header.Set("x-user-agent", request.UserAgent())
	request.Header.Del("User-Agent")
	request.Header.Set("User-Agent", fmt.Sprintf("reverse-proxy/%s", cfg.Version))
}
