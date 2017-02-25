package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// An encoding is a supported content coding.
type encoding int

const (
	encIdentity encoding = iota
	encGzip
)

// CompressibleExtensions are the html extensions to compress.
var (
	gzipWriterPools = map[int]*sync.Pool{}

	compressibleContentTypes = map[string]struct{}{
		"text/plain":             {},
		"text/html":              {},
		"text/css":               {},
		"text/javascript":        {},
		"application/javascript": {},
		"application/atom+xml":   {},
		"application/json":       {},
		"image/svg+xml":          {},
	}
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

// Compressor middleware with default compression.
// Use CompressorLevel to set a different compression level.
var Compressor = CompressorLevel(gzip.DefaultCompression)

// CompressorLevel returns gzip compressable middleware using a given
// gzip level.
func CompressorLevel(level int) func(http.Handler) http.Handler {
	switch level {
	case gzip.NoCompression, gzip.BestSpeed, gzip.BestCompression, gzip.DefaultCompression:
		// OK
	default:
		panic("invalid compressor level")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// check client's accepted encodings
			if encs := acceptedEncodings(r); len(encs) == 0 {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			} else if encs[0] != encGzip {
				next.ServeHTTP(w, r)
				return
			}

			// Create a response writer that will defer it's decision to
			// write gzipped content until the Content-Type header
			// can be inspected.
			gzipWriter := &lazyCompressResponseWriter{
				ResponseWriter: w,
				w:              w,
				level:          level,
			}
			defer gzipWriter.Close()

			next.ServeHTTP(gzipWriter, r)
		})
	}
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

type lazyCompressResponseWriter struct {
	http.ResponseWriter
	w     io.Writer
	level int

	wroteHeader  bool // whether or not WriteHeader has been called
	compressable bool // whether or not the response can be compressed
}

// WriteHeader determines if the compressor should be used and writes
// the http status code.
func (w *lazyCompressResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	defer w.ResponseWriter.WriteHeader(code)

	// Use text/plain content-type if one is not provided.
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/plain")
	}

	var contentType string
	parts := strings.Split(w.Header().Get("Content-Type"), ";")
	if len(parts) > 0 {
		contentType = parts[0]
	}

	if _, ok := compressibleContentTypes[contentType]; !ok {
		return
	} else if strings.Contains(w.Header().Get("Content-Encoding"), "gzip") { // Don't double-encode
		return
	}

	// Compressible. Use gzip.Writer.
	gzw := gzipWriterPools[w.level].Get().(*gzip.Writer)
	gzw.Reset(w.ResponseWriter)
	w.w = gzw

	w.Header().Set("Vary", "Accept-Encoding")
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Del("Content-Length")
	w.Header().Del("Accept-Ranges")
}

// Write writes to the gzip response writer if the response is compressible.
func (w *lazyCompressResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.w.Write(p)
}

// Close closes the writer.
func (w *lazyCompressResponseWriter) Close() error {
	if gzw, ok := w.w.(*gzip.Writer); ok {
		gzw.Close()
		gzipWriterPools[w.level].Put(gzw)
	} else if c, ok := w.w.(io.WriteCloser); ok {
		return c.Close()
	}
	return nil
}
