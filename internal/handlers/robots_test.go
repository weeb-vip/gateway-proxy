package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/weeb-vip/gateway-proxy/config"
)

func TestRobotsDisallowsEverything(t *testing.T) {
	w := httptest.NewRecorder()
	Robots(w, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	if got := w.Body.String(); got != "User-agent: *\nDisallow: /\n" {
		t.Errorf("body %q", got)
	}
	if got := w.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type %q", got)
	}
	if got := w.Header().Get("Cache-Control"); got != "public, max-age=86400" {
		t.Errorf("Cache-Control %q", got)
	}
}

func TestRobotsHeadHasNoBody(t *testing.T) {
	w := httptest.NewRecorder()
	Robots(w, httptest.NewRequest(http.MethodHead, "/robots.txt", nil))

	if w.Code != http.StatusOK || w.Body.Len() != 0 {
		t.Errorf("status %d, body %q; want 200 and empty", w.Code, w.Body.String())
	}
}

func TestRobotsRejectsOtherMethods(t *testing.T) {
	w := httptest.NewRecorder()
	Robots(w, httptest.NewRequest(http.MethodPost, "/robots.txt", nil))

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status %d, want 405", w.Code)
	}
}

// Through the real reverse proxy: the tag is added and Cache-Control is still
// decided as before.
func TestProxiedAnswersAreNoIndex(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer upstream.Close()
	u, _ := url.Parse(upstream.URL)
	proxy := GetProxy(&config.Config{ProxyURL: u, AuthMode: "both", EdgeCacheTTLSeconds: 300, EdgeCacheSWRSeconds: 600}, noParser{})

	rec := httptest.NewRecorder()
	proxy.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/graphql", nil))

	if got := rec.Header().Get("X-Robots-Tag"); got != "noindex" {
		t.Errorf("X-Robots-Tag %q", got)
	}
	if cc := rec.Header().Get("Cache-Control"); !strings.HasPrefix(cc, "public, max-age=0, s-maxage=300") {
		t.Errorf("Cache-Control %q", cc)
	}
}
