package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/weeb-vip/gateway-proxy/config"
	"github.com/weeb-vip/gateway-proxy/internal/jwt"
)

var ec = EdgeCache{TTLSeconds: 300, SWRSeconds: 600}

func TestOnlyAnAnonymousErrorFreeGetIsShareable(t *testing.T) {
	ok := []byte(`{"data":{"anime":[]}}`)
	cases := []struct {
		name          string
		method, path  string
		authenticated bool
		status        int
		body          []byte
		want          string
	}{
		{"anonymous query", "GET", "/graphql", false, 200, ok, "public, max-age=0, s-maxage=300, stale-while-revalidate=600"},
		{"signed-in query", "GET", "/graphql", true, 200, ok, "private, no-store"},
		{"a POST, whoever sent it", "POST", "/graphql", false, 200, ok, "private, no-store"},
		{"persisted query not found", "GET", "/graphql", false, 200, []byte(`{"errors":[{"message":"PersistedQueryNotFound"}],"data":null}`), "no-store"},
		{"empty errors array is fine", "GET", "/graphql", false, 200, []byte(`{"errors":[],"data":{}}`), "public, max-age=0, s-maxage=300, stale-while-revalidate=600"},
		{"not a 200", "GET", "/graphql", false, 500, ok, "no-store"},
		{"not json", "GET", "/graphql", false, 200, []byte(`<html>`), "no-store"},
		{"another path", "GET", "/healthcheck", false, 200, ok, "private, no-store"},
	}
	for _, c := range cases {
		if got := CacheControlFor(c.method, c.path, c.authenticated, c.status, c.body, ec); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

// Through the real reverse proxy: the header lands on the response and the
// body still reaches the client intact.
func TestTheProxySetsCacheControlAndKeepsTheBody(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("extensions") == "" {
			w.Write([]byte(`{"errors":[{"message":"PersistedQueryNotFound"}]}`))
			return
		}
		w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer upstream.Close()
	u, _ := url.Parse(upstream.URL)
	cfg := &config.Config{ProxyURL: u, AuthMode: "both", EdgeCacheTTLSeconds: 300, EdgeCacheSWRSeconds: 600}
	proxy := GetProxy(cfg, noParser{})

	get := func(target string, cookie string) *http.Response {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", target, nil)
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		proxy.ServeHTTP(rec, req)
		return rec.Result()
	}

	res := get("/graphql?extensions=%7B%7D", "")
	body, _ := io.ReadAll(res.Body)
	if cc := res.Header.Get("Cache-Control"); !strings.HasPrefix(cc, "public, max-age=0, s-maxage=300") {
		t.Errorf("anonymous: Cache-Control %q", cc)
	}
	if string(body) != `{"data":{"ok":true}}` {
		t.Errorf("body altered: %s", body)
	}

	res = get("/graphql?extensions=%7B%7D", "access_token=abc")
	if cc := res.Header.Get("Cache-Control"); cc != "private, no-store" {
		t.Errorf("signed in: Cache-Control %q", cc)
	}

	res = get("/graphql", "")
	if cc := res.Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("error answer: Cache-Control %q", cc)
	}
}

// A parser that accepts nothing: the request above still counts as
// authenticated because it carried a token, which is the point.
type noParser struct{}

func (noParser) Parse(string) (*jwt.ParsedJWT, error) { return nil, io.EOF }
