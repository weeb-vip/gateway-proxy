package handlers_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/weeb-vip/gateway-proxy/config"
	"github.com/weeb-vip/gateway-proxy/internal/handlers"
	"github.com/weeb-vip/gateway-proxy/internal/jwt"
	"net/http/httptest"
	"net/url"
	"testing"
)

type resultFactory func(token string) (*jwt.ParsedJWT, error)
type mockParser struct {
	token         *jwt.ParsedJWT
	err           error
	resultFactory resultFactory
}

func (r mockParser) Parse(token string) (*jwt.ParsedJWT, error) {
	if r.resultFactory != nil {
		return r.resultFactory(token)
	}
	return r.token, r.err
}

func TestGetProxy(t *testing.T) {
	t.Run("it changes the host to target host", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/", nil)
		request.URL.Host = "https://example.com"
		request.Header.Set("User-Agent", "my-application")
		proxyURL, _ := url.Parse("http://localhost:8080")
		handlers.GetProxy(&config.Config{ProxyURL: proxyURL}, mockParser{}).Director(request)

		assert.Equal(t, "localhost:8080", request.URL.Host)
	})
	t.Run("adds x-user-agent", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Set("User-Agent", "my-application")
		proxyURL, _ := url.Parse("http://localhost:8080")
		handlers.GetProxy(&config.Config{ProxyAddress: "http://localhost:8080", ProxyURL: proxyURL}, mockParser{}).Director(request)

		assert.Equal(t, "my-application", request.Header.Get("x-user-agent"))
	})
	t.Run("adds x-remote-ip", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Add("x-forwarded-for", "192.168.1.1")
		proxyURL, _ := url.Parse("http://localhost:8080")
		handlers.GetProxy(&config.Config{ProxyAddress: "http://localhost:8080", ProxyURL: proxyURL}, mockParser{}).Director(request)

		assert.Equal(t, "192.168.1.1", request.Header.Get("x-remote-ip"))
	})
	t.Run("adds user agent of proxy", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Add("user-agent", "my-agent")
		request.Header.Add("x-forwarded-for", "192.168.1.1")
		proxyURL, _ := url.Parse("http://localhost:8080")
		handlers.GetProxy(&config.Config{ProxyURL: proxyURL, Version: "x.y.z"}, mockParser{}).Director(request)

		assert.Equal(t, "my-agent", request.Header.Get("x-user-agent"))
		assert.Equal(t, fmt.Sprintf("reverse-proxy/%s", "x.y.z"), request.UserAgent())
	})

	t.Run("passes correct token into JWT parser and adds correct information to request headers", func(t *testing.T) {
		request := httptest.NewRequest("GET", "/", nil)
		request.Header.Add("Authorization", "Bearer 123")
		request.Header.Add("user-agent", "my-agent")
		request.Header.Add("x-forwarded-for", "192.168.1.1")
		proxyURL, _ := url.Parse("http://localhost:8080")
		handlers.GetProxy(&config.Config{ProxyURL: proxyURL, Version: "x.y.z"}, mockParser{resultFactory: func(token string) (*jwt.ParsedJWT, error) {
			if token == "123" {
				return &jwt.ParsedJWT{
					Subject:  getPointer("Subject"),
					Audience: getPointer("Audience"),
					Purpose:  getPointer("Purpose"),
				}, nil
			}
			panic("authorization token not passed correctly")
		}}).Director(request)

		assert.Equal(t, "Subject", request.Header.Get("x-user-id"))
		assert.Equal(t, "Purpose", request.Header.Get("x-token-purpose"))
		assert.Equal(t, "123", request.Header.Get("x-raw-token"))
	})
}

func getPointer[T any](input T) *T {
	return &input
}
