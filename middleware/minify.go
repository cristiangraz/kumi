package middleware

import (
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"sync"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/xml"
)

type (
	// minifyResponseWriter provides a response writer for minification.
	minifyResponseWriter struct {
		http.ResponseWriter
		io.WriteCloser
		minifier    *minify.M
		allowed     map[string]struct{}
		initialized bool
	}
)

var (
	minifyResponseWriterPool = &sync.Pool{
		New: func() interface{} {
			return &minifyResponseWriter{}
		},
	}
)

// reset the response writer pulled from the pool
func (m *minifyResponseWriter) reset(w http.ResponseWriter, minifier *minify.M, allowed map[string]struct{}) {
	m.ResponseWriter = w
	m.minifier = minifier
	m.allowed = allowed
	m.WriteCloser = nil
	m.initialized = false
}

// Write defers to the initializer on first run to
// see if the content should be minified or not.
func (m *minifyResponseWriter) Write(b []byte) (int, error) {
	if !m.initialized {
		m.initialize()
	}

	if m.WriteCloser == nil {
		return m.ResponseWriter.Write(b)
	}

	return m.WriteCloser.Write(b)
}

// initialize checks for a valid content-type in the allowed list of
// content types and initializes the correct minifier if found.
// If the response has a no-transform value in Cache-Control,
// nothing is minified.
func (m *minifyResponseWriter) initialize() {
	m.initialized = true
	hdr := m.ResponseWriter.Header()

	cc := hdr.Get("Cache-Control")
	if strings.Contains(cc, "no-transform") {
		return
	}

	ct, _, err := mime.ParseMediaType(hdr.Get("Content-Type"))
	if err != nil {
		return
	}

	if _, ok := m.allowed[ct]; !ok {
		return
	}

	m.WriteCloser = m.minifier.Writer(ct, m.ResponseWriter)
}

// closes the minifier.
func (m *minifyResponseWriter) close() {
	if m.WriteCloser == nil {
		return
	}

	if err := m.WriteCloser.Close(); err != nil {
		log.Println("Minification Error: Err:", err)
	}
}

// Minify returns minify middleware that will minify css, javascript, and json
var Minify = MinifyTypes("text/css", "text/javascript", "application/json", "text/xml")

// MinifyTypes returns a custom minifier.
func MinifyTypes(contentTypes ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(contentTypes))
	for _, t := range contentTypes {
		allowed[t] = struct{}{}
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("application/json", json.Minify)
	m.AddFunc("text/xml", xml.Minify)

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			mrw := minifyResponseWriterPool.Get().(*minifyResponseWriter)
			mrw.reset(w, m, allowed)

			defer func() {
				mrw.close()
				minifyResponseWriterPool.Put(mrw)
			}()

			next.ServeHTTP(mrw, r)
		}
		return http.HandlerFunc(fn)
	}
}
