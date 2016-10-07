package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

// CompressibleExtensions are the html extensions to compress.
var (
	gzipWriterPools = map[int]*sync.Pool{}

	rxExtension = regexp.MustCompile(`\.((?:min\.)?[a-zA-Z]{2,4})$`)

	CompressibleExtensions = map[string]struct{}{
		"html":    {},
		"js":      {},
		"min.js":  {},
		"css":     {},
		"min.css": {},
		"txt":     {},
		"xml":     {},
		"json":    {},
		"svg":     {},
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
func (gzw gzipResponseWriter) Write(p []byte) (int, error) {
	return gzw.writer.Write(p)
}

// Compressor middleware with default compression.
// Use CompressorLevel to set a different compression level.
var Compressor = CompressorLevel(gzip.DefaultCompression)

// CompressorLevel returns gzip compressable middleware using a given
// gzip level.
func CompressorLevel(level int) func(http.Handler) http.Handler {
	if level != gzip.NoCompression && level != gzip.BestSpeed && level != gzip.BestCompression && level != gzip.DefaultCompression {
		panic("Invalid compressor level.")
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// check client's accepted encodings
			if encs := acceptedEncodings(r); len(encs) == 0 {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			} else if encs[0] != encGzip {
				next.ServeHTTP(w, r)
				return
			} else if strings.Contains(w.Header().Get("Content-Encoding"), "gzip") { // Don't double-encode
				next.ServeHTTP(w, r)
				return
			}

			// If there is a file extension and it doesn't match compressible
			// extensions, skip
			if exts := rxExtension.FindStringSubmatch(r.URL.Path); len(exts) == 0 {
				next.ServeHTTP(w, r)
				return
			} else if len(exts) == 1 {
				ext := exts[0]
				if _, ok := CompressibleExtensions[ext]; !ok {
					next.ServeHTTP(w, r)
					return
				}
			}

			// cannot accept Range requests for possibly gzipped responses
			r.Header.Del("Range")

			w.Header().Set("Vary", "Accept-Encoding")
			setCompressionHeaders(w.Header())

			gzw := gzipWriterPools[level].Get().(*gzip.Writer)
			gzw.Reset(w)

			w = &gzipResponseWriter{w, gzw}
			next.ServeHTTP(w, r)

			gzw.Close()
			gzipWriterPools[level].Put(gzw)
		}
		return http.HandlerFunc(fn)
	}
}

// Compress ...
func Compress(r io.Reader, w io.Writer) {
	level := gzip.DefaultCompression

	gzw := gzipWriterPools[level].Get().(*gzip.Writer)
	gzw.Reset(w)

	io.Copy(gzw, r)

	gzw.Close()
	gzipWriterPools[level].Put(gzw)
}

// Decompress ...
func Decompress(r io.Reader, w io.Writer) {
	gzr, _ := gzip.NewReader(r)
	io.Copy(w, gzr)
	gzr.Close()
}

// AcceptsEncoding ...
func AcceptsEncoding(r *http.Request) bool {
	if encs := acceptedEncodings(r); len(encs) == 0 {
		return false
	} else if encs[0] != encGzip {
		return false
	}
	return true
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
