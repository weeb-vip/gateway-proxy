package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/weeb-vip/gateway-proxy/config"
)

func TestAnOriginlessGetOfGraphQLGetsTheSitesOwnAllowOrigin(t *testing.T) {
	cfg := &config.Config{CORSAllowedOrigins: []string{"https://weeb.vip", "https://admin.weeb.vip"}, CORSAllowCredentials: true}
	h := CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("{}")) }))

	serve := func(method, path, origin string) http.Header {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		h.ServeHTTP(rec, req)
		return rec.Result().Header
	}

	// A server-side render: no Origin, but its answer may be cached and
	// served to the site's browsers, so it carries the site's header.
	if got := serve("GET", "/graphql?extensions=x", "").Get("Access-Control-Allow-Origin"); got != "https://weeb.vip" {
		t.Errorf("origin-less GET /graphql: %q", got)
	}
	// Everything else is as before: the asking origin when allowed, nothing when not.
	if got := serve("GET", "/graphql", "https://admin.weeb.vip").Get("Access-Control-Allow-Origin"); got != "https://admin.weeb.vip" {
		t.Errorf("admin origin: %q", got)
	}
	if got := serve("POST", "/graphql", "").Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("origin-less POST: %q", got)
	}
	if got := serve("GET", "/healthcheck", "").Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("origin-less other path: %q", got)
	}
}
