package cache

import (
	"log"
	"net/http"
	"strconv"
)

// IsRequestCacheable checks to see if the request was a GET or HEAD request
// with no Authorization or Cookie headers.
func IsRequestCacheable(r *http.Request) bool {
	if r.Method == "GET" || r.Method == "HEAD" {
		if r.Header.Get("Authorization") == "" && r.Header.Get("Cookie") == "" {
			return true
		}
	}

	return false
}

// IsResponseCacheable checks to see if the response can be cached, and if so, for how long.
// Criteria:
//  - Request must be a GET or HEAD request with no Authorization or Cookie headers.
//  - Response must be a 200, 203, 300, 301, 302, 404, or 410 sttaus code
//  - Response must have a Cache-Control header
//  - Response Cache-Control header cannot contain no-store or private directives
//  - Response Cache-control header must explicitly contain a public directive
//  - TTL gives preference to the s-maxage directive, followed by the max-age directive. Expires
// is not used. If the s-maxage or max-age are set to zero, the response is not cacheable.
func IsResponseCacheable(r *http.Request, rw http.ResponseWriter, status int) (cacheable bool, ttl int) {
	// If the request isn't cacheable, then don't bother evaluating the response.
	if !IsRequestCacheable(r) {
		log.Println("request not cacheable")
		return false, 0
	}

	allowedStatusCodes := map[int]struct{}{200: {}, 203: {}, 300: {}, 301: {}, 302: {}, 404: {}, 410: {}}
	if _, ok := allowedStatusCodes[status]; !ok {
		return false, 0
	}

	h := rw.Header()
	cc := parseCacheControl(h.Get("Cache-Control"))
	if cc.IsEmpty() || cc.Has("no-store") || cc.IsPrivate() || !cc.IsPublic() {
		log.Println(h.Get("Cache-Control"))
		return false, 0
	}

	if cc.Has("s-maxage") {
		sharedMaxAge, err := strconv.Atoi(cc.Get("s-maxage"))
		if err != nil || sharedMaxAge == 0 {
			return false, 0
		}

		return true, sharedMaxAge
	}

	if cc.Has("max-age") {
		maxAge, err := strconv.Atoi(cc.Get("max-age"))
		if err != nil || maxAge == 0 {
			return false, 0
		}

		return true, maxAge
	}

	// Cacheable but we don't know for how long
	return true, 0
}
