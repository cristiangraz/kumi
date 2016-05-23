package cache

import (
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

type nullInt64 struct {
	Int64 int64
	Valid bool
}

// Headers is used to generate cache-control headers.
type Headers struct {
	b               []byte
	public          bool
	private         bool
	maxAge          nullInt64 // 0 is a valid max-age
	sharedMaxAge    nullInt64 // 0 is a valid s-maxage
	noCache         bool
	noStore         bool
	noTransform     bool
	mustRevalidate  bool
	proxyRevalidate bool
}

// Cache-Control directives.
var (
	private         = []byte("private")
	public          = []byte("public")
	maxAge          = []byte("max-age")
	sharedMaxAge    = []byte("s-maxage")
	noCache         = []byte("no-cache")
	noStore         = []byte("no-store")
	noTransform     = []byte("no-transform")
	mustRevalidate  = []byte("must-revalidate")
	proxyRevalidate = []byte("proxy-revalidate")
)

var pool = &sync.Pool{
	New: func() interface{} {
		return &Headers{
			b: make([]byte, 0, defaultByteBufferSize),
		}
	},
}

const defaultByteBufferSize = 128

// New returns a Headers struct pulled from a sync pool.
func New() *Headers {
	v := pool.Get()
	if v == nil {
		return &Headers{
			b: make([]byte, 0, defaultByteBufferSize),
		}
	}
	return v.(*Headers)
}

// NewString returns a Headers struct from a cache-control header.
func NewString(cc string) *Headers {
	h := New()
	h.Parse(cc)
	return h
}

// Release returns *Headers back to the pool.
func Release(h *Headers) {
	var ma nullInt64
	var sma nullInt64
	h.b = h.b[:0]
	h.public = false
	h.private = false
	h.maxAge = ma
	h.sharedMaxAge = sma
	h.noCache = false
	h.noStore = false
	h.noTransform = false
	h.mustRevalidate = false
	h.proxyRevalidate = false

	pool.Put(h)
}

// IsEmpty checks if there are zero directives.
func (h *Headers) IsEmpty() bool {
	if h.public == true {
		return false
	} else if h.private == true {
		return false
	} else if h.maxAge.Valid {
		return false
	} else if h.sharedMaxAge.Valid {
		return false
	} else if h.noCache == true {
		return false
	} else if h.noStore == true {
		return false
	} else if h.noTransform == true {
		return false
	} else if h.mustRevalidate == true {
		return false
	} else if h.proxyRevalidate == true {
		return false
	}
	return true
}

// SetPublic lets user agents know this is a public response.
func (h *Headers) SetPublic() *Headers {
	h.public = true
	h.private = false
	return h
}

// SetPrivate lets user agents know this is a private response.
func (h *Headers) SetPrivate() *Headers {
	h.private = true
	h.public = false
	return h
}

// IsPublic checks to see if the cache-control header includes the public directive.
func (h *Headers) IsPublic() bool {
	return h.public
}

// IsPrivate checks to see if the cache-control header includes the public directive.
func (h *Headers) IsPrivate() bool {
	return h.private
}

// NoTransform sets a no-transform directive.
func (h *Headers) NoTransform() *Headers {
	h.noTransform = true
	return h
}

// NoCache adds a no-cache directive
func (h *Headers) NoCache() *Headers {
	h.noCache = true
	return h
}

// NoStore adds a no-store directive
func (h *Headers) NoStore() *Headers {
	h.noStore = true
	return h
}

// MustRevalidate adds the must-revalidate directive.
func (h *Headers) MustRevalidate() *Headers {
	h.mustRevalidate = true
	return h
}

// ProxyRevalidate adds the proxy-revalidate directive.
func (h *Headers) ProxyRevalidate() *Headers {
	h.proxyRevalidate = true
	return h
}

// SetMaxAge sets a max age for the response.
func (h *Headers) SetMaxAge(age int64) *Headers {
	h.maxAge = nullInt64{Int64: age, Valid: true}
	return h
}

// SetSharedMaxAge sets a shared max age for the response.
func (h *Headers) SetSharedMaxAge(age int64) *Headers {
	h.sharedMaxAge = nullInt64{Int64: age, Valid: true}
	return h
}

// convenience byte slices
var (
	equalSign = []byte("=")
	separate  = []byte(", ")
)

// Strings returns the cache-control header as a string
func (h *Headers) String() string {
	// Because there is a finite number of fields, the fields are appended in
	// alphabetical order so we don't need a sorting algorithm.
	// The fields are appended to a byte buffer to minimize allocations.
	if h.maxAge.Valid {
		if len(h.b) > 0 {
			h.b = append(h.b, separate...)
		}
		h.b = appendByteSlices(h.b, maxAge, equalSign)
		h.b = strconv.AppendInt(h.b, h.maxAge.Int64, 10)
	}
	if h.mustRevalidate {
		if len(h.b) > 0 {
			h.b = append(h.b, separate...)
		}
		h.b = append(h.b, mustRevalidate...)
	}
	if h.noCache {
		if len(h.b) > 0 {
			h.b = append(h.b, separate...)
		}
		h.b = append(h.b, noCache...)
	}
	if h.noStore {
		if len(h.b) > 0 {
			h.b = append(h.b, separate...)
		}
		h.b = append(h.b, noStore...)
	}
	if h.noTransform {
		if len(h.b) > 0 {
			h.b = append(h.b, separate...)
		}
		h.b = append(h.b, noTransform...)
	}
	if h.proxyRevalidate {
		if len(h.b) > 0 {
			h.b = append(h.b, separate...)
		}
		h.b = append(h.b, proxyRevalidate...)
	}
	if h.public {
		if len(h.b) > 0 {
			h.b = append(h.b, separate...)
		}
		h.b = append(h.b, public...)
	}
	if h.private {
		if len(h.b) > 0 {
			h.b = append(h.b, separate...)
		}
		h.b = append(h.b, private...)
	}
	if h.sharedMaxAge.Valid {
		if len(h.b) > 0 {
			h.b = append(h.b, separate...)
		}
		h.b = appendByteSlices(h.b, sharedMaxAge, equalSign)
		h.b = strconv.AppendInt(h.b, h.sharedMaxAge.Int64, 10)
	}

	if len(h.b) == 0 {
		return ""
	}
	return string(h.b)
}

func appendByteSlices(bb []byte, b ...[]byte) []byte {
	for _, buf := range b {
		bb = append(bb, buf...)
	}
	return bb
}

// SensibleDefaults sets sensible defaults for the Cache-Control header.
// Follow's Symfony's guidelines for defaults:
// http://symfony.com/doc/current/book/http_cache.html#caching-rules-and-defaults
func (h *Headers) SensibleDefaults(header http.Header, status int) string {
	if h.IsEmpty() {
		if status == http.StatusMovedPermanently {
			h.SetPublic().SetSharedMaxAge(60)
			return h.String()
		}

		if header.Get("Expires") == "" && header.Get("ETag") == "" && header.Get("Last-Modified") == "" {
			h.NoCache()
		} else if header.Get("Expires") != "" || header.Get("ETag") != "" || header.Get("Last-Modified") != "" {
			h.SetPrivate().MustRevalidate()
		}
	}

	if !h.public && !h.private && !h.sharedMaxAge.Valid {
		h.SetPrivate()
	}
	return h.String()
}

var rxCacheControlHeader = regexp.MustCompile(`([a-zA-Z][a-zA-Z_-]*)\s*(?:=(?:"([^"]*)"|([^ \t",;]*)))?`)

// Parse parses a cache-control header
func (h *Headers) Parse(cc string) {
	if cc == "" {
		return
	}

	matches := rxCacheControlHeader.FindAllStringSubmatch(cc, -1)
	for _, v := range matches {
		switch v[1] {
		case "public":
			h.SetPublic()
		case "private":
			h.SetPrivate()
		case "max-age":
			i, _ := strconv.ParseInt(v[3], 10, 64)
			h.maxAge = nullInt64{Int64: i, Valid: true}
		case "s-maxage":
			i, _ := strconv.ParseInt(v[3], 10, 64)
			h.sharedMaxAge = nullInt64{Int64: i, Valid: true}
		case "no-cache":
			h.noCache = true
		case "no-store":
			h.noStore = true
		case "no-transform":
			h.noTransform = true
		case "must-revalidate":
			h.mustRevalidate = true
		}
	}
}
