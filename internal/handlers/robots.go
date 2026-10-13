package handlers

import "net/http"

// robotsTxt asks every crawler to stay out. gateway.weeb.vip is an API, not
// a site: there is nothing here to index, and Googlebot was spending its
// crawl budget on GraphQL URLs it found in the site's rendered pages.
const robotsTxt = "User-agent: *\nDisallow: /\n"

// Robots answers /robots.txt from this process; it is never proxied to the
// router.
func Robots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write([]byte(robotsTxt))
}

// setNoIndex marks every proxied answer as not for a search index, for the
// crawler that fetches a URL without reading robots.txt first.
func setNoIndex(resp *http.Response) error {
	resp.Header.Set("X-Robots-Tag", "noindex")
	return nil
}
