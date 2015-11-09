package cache

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type (
	// Headers is used to generate cache-control headers.
	Headers struct {
		directives map[string]string
	}
)

var (
	rxCacheControlHeader = regexp.MustCompile(`([a-zA-Z][a-zA-Z_-]*)\s*(?:=(?:"([^"]*)"|([^ \t",;]*)))?`)
)

// NewHeaders returns a Headers struct.
func NewHeaders() *Headers {
	return &Headers{
		directives: map[string]string{},
	}
}

// NewHeadersString returns a Headers struct from a cache-control header.
func NewHeadersString(cc string) *Headers {
	return parseCacheControl(cc)
}

// IsEmpty checks if there are zero directives.
func (h *Headers) IsEmpty() bool {
	return len(h.directives) == 0
}

// AddDirective adds a cache-control directive with no value
// i.e public, private, must-revalidate
func (h *Headers) AddDirective(name string) *Headers {
	h.AddDirectiveValue(name, "")
	return h
}

// AddDirectiveValue adds a cache-control directive with a value
// i.e. max-age, s-maxage
func (h *Headers) AddDirectiveValue(name string, value string) *Headers {
	h.directives[name] = value
	return h
}

// RemoveDirective removes a directive
func (h *Headers) RemoveDirective(name string) *Headers {
	delete(h.directives, name)
	return h
}

// SetPublic lets user agents know this is a public response.
func (h *Headers) SetPublic() *Headers {
	return h.AddDirective("public").RemoveDirective("private")
}

// SetPrivate lets user agents know this is a private response.
func (h *Headers) SetPrivate() *Headers {
	return h.AddDirective("private").RemoveDirective("public")
}

// IsPublic checks to see if the cache-control header includes the public directive.
func (h *Headers) IsPublic() bool {
	return h.Has("public")
}

// IsPrivate checks to see if the cache-control header includes the public directive.
func (h *Headers) IsPrivate() bool {
	return h.Has("private")
}

// Has returns bool true if the directive exists.
func (h *Headers) Has(name string) bool {
	_, ok := h.directives[name]
	return ok
}

// Get returns a directive value by name.
func (h *Headers) Get(name string) string {
	if v, ok := h.directives[name]; ok {
		return v
	}

	return ""
}

// NoTransform sets a no-transform directive.
func (h *Headers) NoTransform() *Headers {
	return h.AddDirective("no-transform")
}

// NoCache adds a no-cache directive
func (h *Headers) NoCache() *Headers {
	return h.AddDirective("no-cache")
}

// MustRevalidate adds the must-revalidate directive.
func (h *Headers) MustRevalidate() *Headers {
	return h.AddDirective("must-revalidate")
}

// SetMaxAge sets a max age for the response.
func (h *Headers) SetMaxAge(age int) *Headers {
	return h.AddDirectiveValue("max-age", strconv.Itoa(age))
}

// SetSharedMaxAge sets a shared max age for the response.
func (h *Headers) SetSharedMaxAge(age int) *Headers {
	return h.AddDirectiveValue("s-maxage", strconv.Itoa(age))
}

// Strings returns the cache-control header as a string
func (h *Headers) String() string {
	if len(h.directives) == 0 {
		return ""
	}

	var parts []string
	keys, values := sortMap(h.directives, false)
	for i, k := range keys {
		v := values[i]
		if v == "" {
			parts = append(parts, k)
		} else {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return strings.Join(parts, ", ")
}

// Add adds the Cache-Control header to an http.Header
func (h *Headers) Add(header http.Header) {
	header.Set("Cache-Control", h.String())
}

// SensibleDefaults sets sensible defaults -- taking into consideration any existing
// Cache-Control headers and other cache-headers.
// If http.Header has Cache-Control headers set, those will take precedence over anything
// in Headers. Follow's Symfony's guidelines for defaults:
// http://symfony.com/doc/current/book/http_cache.html#caching-rules-and-defaults
// @todo support 301 headers as cacheable
func (h *Headers) SensibleDefaults(header http.Header, status int) {
	cch := NewHeadersString(header.Get("Cache-Control"))
	if cch.IsEmpty() {
		cch = h
	}

	if cch.IsEmpty() {
		if status == 301 {
			cch.SetPublic().SetSharedMaxAge(60)
			cch.Add(header)

			return
		}

		if header.Get("Expires") == "" && header.Get("ETag") == "" && header.Get("Last-Modified") == "" {
			cch.AddDirective("no-cache")
		}

		if header.Get("Expires") != "" || header.Get("ETag") != "" || header.Get("Last-Modified") != "" {
			cch.AddDirective("private").MustRevalidate()
		}
	}

	if !cch.IsPublic() && !cch.IsPrivate() && !cch.Has("s-maxage") {
		cch.SetPrivate()
	}

	cch.Add(header)
}

// parseCacheControl parses a cache-control header
func parseCacheControl(cc string) *Headers {
	if cc == "" {
		return NewHeaders()
	}

	matches := rxCacheControlHeader.FindAllStringSubmatch(cc, -1)
	h := NewHeaders()
	for _, v := range matches {
		h.AddDirectiveValue(v[1], v[3])
	}

	return h
}
