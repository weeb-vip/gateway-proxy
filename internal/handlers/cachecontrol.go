package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EdgeCache is how long a CDN may hold an anonymous query response, and
// the one browser origin such an answer is for.
//
// A CDN keys on the URL, not on the Origin header, so a shared answer can
// carry exactly one Access-Control-Allow-Origin. That is the site's own
// origin (the first CORS allowed origin): a request from it, or from no
// origin at all (a server-side render), may be shared; a request from any
// other origin (the admin panel, a local dev server) is answered fresh and
// marked private, with the CORS header it asked for.
type EdgeCache struct {
	TTLSeconds int
	SWRSeconds int
	Origin     string
}

// Shareable reports whether a request's Origin is one a cached answer may
// serve: none, or the site's own.
func (ec EdgeCache) Shareable(origin string) bool {
	return origin == "" || (ec.Origin != "" && origin == ec.Origin)
}

// Cache-Control for a GraphQL response, decided here because this is the one
// hop that knows who asked.
//
// Only an anonymous GET of a successful, error-free query is shareable:
// "public, s-maxage" lets Cloudflare hold it for everyone. Anything a user
// asked for carries their data and is "private, no-store"; so is every POST
// (mutations, and the registration of a persisted query), and so is any
// answer carrying errors -- above all PersistedQueryNotFound, which cached
// would hand every visitor the error instead of the data.
func CacheControlFor(method, path string, authenticated bool, status int, body []byte, ec EdgeCache) string {
	if path != "/graphql" || method != http.MethodGet || authenticated {
		return "private, no-store"
	}
	if status != http.StatusOK || hasErrors(body) {
		return "no-store"
	}
	return fmt.Sprintf("public, max-age=0, s-maxage=%d, stale-while-revalidate=%d", ec.TTLSeconds, ec.SWRSeconds)
}

// hasErrors: a GraphQL response with a non-empty top-level errors array.
// Unparseable bodies count as errors: nothing unknown gets cached.
func hasErrors(body []byte) bool {
	var payload struct {
		Errors []json.RawMessage `json:"errors"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return true
	}
	return len(payload.Errors) > 0
}

// maxInspectedBody bounds what is read back to decide cacheability; a bigger
// answer is passed through untouched and marked no-store.
const maxInspectedBody = 8 << 20

// authenticated reports whether the proxied request carried a user: the
// director sets x-user-id from a valid token, and a token that did not parse
// is still a token -- the answer may be an auth error, never shared.
func authenticated(r *http.Request) bool {
	if r.Header.Get("x-user-id") != "" || r.Header.Get("Authorization") != "" {
		return true
	}
	_, err := r.Cookie("access_token")
	return err == nil
}

// setCacheControl is the reverse proxy's ModifyResponse: it reads the body
// once (restoring it), decides, and sets the header.
func setCacheControl(ec EdgeCache) func(*http.Response) error {
	return func(resp *http.Response) error {
		req := resp.Request
		if req == nil || req.URL.Path != "/graphql" {
			return nil
		}
		if req.Method != http.MethodGet || authenticated(req) || !ec.Shareable(req.Header.Get("Origin")) {
			resp.Header.Set("Cache-Control", "private, no-store")
			return nil
		}
		if resp.ContentLength > maxInspectedBody {
			resp.Header.Set("Cache-Control", "no-store")
			return nil
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxInspectedBody+1))
		if err != nil {
			return err
		}
		resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))
		if len(body) > maxInspectedBody {
			resp.Header.Set("Cache-Control", "no-store")
			return nil
		}
		resp.Header.Set("Cache-Control", CacheControlFor(req.Method, req.URL.Path, false, resp.StatusCode, body, ec))
		return nil
	}
}
