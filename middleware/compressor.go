package middleware

import (
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/cristiangraz/kumi"
	"github.com/klauspost/compress/gzip"
)

type (
	gzipResponseWriter struct {
		http.ResponseWriter
		writer io.Writer
	}

	// An encoding is a supported content coding.
	encoding int
)

const (
	encIdentity encoding = iota
	encGzip
)

var (
	gzipWriterPools = map[int]*sync.Pool{}

	// CompressibleContentTypes is a list of content types that the compressor
	// will compress. The list is taken from the article found at
	// https://www.fastly.com/blog/new-gzip-settings-and-deciding-what-to-compress
	CompressibleContentTypes = map[string]struct{}{
		"text/html":                     {},
		"application/x-javascript":      {},
		"text/css":                      {},
		"application/javascript":        {},
		"text/javascript":               {},
		"text/plain":                    {},
		"text/xml":                      {},
		"application/json":              {},
		"application/vnd.ms-fontobject": {},
		"application/x-font-opentype":   {},
		"application/x-font-truetype":   {},
		"application/x-font-ttf":        {},
		"application/xml":               {},
		"font/eot":                      {},
		"font/opentype":                 {},
		"font/otf":                      {},
		"image/svg+xml":                 {},
		"image/vnd.microsoft.icon":      {},
	}

	acceptEncoding  = "Accept-Encoding"
	contentEncoding = "Content-Encoding"
	vary            = "Vary"
)

func init() {
	for _, level := range []int{gzip.NoCompression, gzip.BestSpeed, gzip.BestCompression, gzip.DefaultCompression} {
		gzipWriterPools[level] = &sync.Pool{
			New: func() interface{} {
				w, _ := gzip.NewWriterLevel(nil, level)
				return w
			},
		}
	}
}

// Write writes to the gzip response writer if the response is compressible.
func (gzw *gzipResponseWriter) Write(p []byte) (int, error) {
	return gzw.writer.Write(p)
}

// Compressor middleware with default compression.
// Use CompressorLevel to set a different compression level.
var Compressor = CompressorLevel(gzip.DefaultCompression)

// CompressorLevel returns gzip compressable middleware using a given
// gzip level.
func CompressorLevel(level int) kumi.HandlerFunc {
	if level != gzip.NoCompression && level != gzip.BestSpeed && level != gzip.BestCompression && level != gzip.DefaultCompression {
		panic("Invalid compressor level.")
	}

	return func(c *kumi.Context) {
		c.Header().Set("Vary", "Accept-Encoding")

		// check client's accepted encodings
		encs := acceptedEncodings(c.Request)
		if len(encs) == 0 {
			c.WriteHeader(http.StatusNotAcceptable)
			return
		}

		if encs[0] != encGzip {
			c.Next()
			return
		}

		// cannot accept Range requests for possibly gzipped responses
		c.Request.Header.Del("Range")
		c.BeforeWrite(func() {
			if isResponseCompressible(c.Header()) {
				setCompressionHeaders(c.Header())

				gzw := gzipWriterPools[level].Get().(*gzip.Writer)
				gzw.Reset(c.ResponseWriter)

				c.ResponseWriter = &gzipResponseWriter{c.ResponseWriter, gzw}

				c.Defer(func() {
					gzw.Close()
					gzipWriterPools[level].Put(gzw)
				})
			}
		})

		c.Next()
	}
}

// IsResponseCompressible returns true if the response has a Content-Type
// found in the CompressibleContentTypes map.
func isResponseCompressible(h http.Header) bool {
	ct := h.Get("Content-Type")
	mt, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return false
	}

	if _, ok := CompressibleContentTypes[mt]; ok {
		return true
	}

	return false
}

// SetHeaders sets gzip headers.
func setCompressionHeaders(h http.Header) {
	h.Set(contentEncoding, "gzip")
	h.Del("Content-Length")
	h.Del("Accept-Ranges")
}

// acceptedEncodings returns the supported content codings that are
// accepted by the request r. It returns a slice of encodings in
// client preference order.
//
// If the Sec-WebSocket-Key header is present then compressed content
// encodings are not considered.
//
// Source: https://github.com/xi2/httpgzip
//
// ref: http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html
func acceptedEncodings(r *http.Request) []encoding {
	h := r.Header.Get("Accept-Encoding")
	swk := r.Header.Get("Sec-WebSocket-Key")
	if h == "" {
		return []encoding{encIdentity}
	}
	gzip := float64(-1)    // -1 means not accepted, 0 -> 1 means value of q
	identity := float64(0) // -1 means not accepted, 0 -> 1 means value of q
	for _, s := range strings.Split(h, ",") {
		f := strings.Split(s, ";")
		f0 := strings.ToLower(strings.Trim(f[0], " "))
		q := float64(1.0)
		if len(f) > 1 {
			f1 := strings.ToLower(strings.Trim(f[1], " "))
			if strings.HasPrefix(f1, "q=") {
				if flt, err := strconv.ParseFloat(f1[2:], 32); err == nil {
					if flt >= 0 && flt <= 1 {
						q = flt
					}
				}
			}
		}
		if (f0 == "gzip" || f0 == "*") && q > gzip && swk == "" {
			gzip = q
		}
		if (f0 == "gzip" || f0 == "*") && q == 0 {
			gzip = -1
		}
		if (f0 == "identity" || f0 == "*") && q > identity {
			identity = q
		}
		if (f0 == "identity" || f0 == "*") && q == 0 {
			identity = -1
		}
	}
	switch {
	case gzip == -1 && identity == -1:
		return []encoding{}
	case gzip == -1:
		return []encoding{encIdentity}
	case identity == -1:
		return []encoding{encGzip}
	case identity > gzip:
		return []encoding{encIdentity, encGzip}
	default:
		return []encoding{encGzip, encIdentity}
	}
}
