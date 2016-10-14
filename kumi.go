package kumi

import (
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/justinas/alice"
)

// Engine embeds RouterGroup and provides methods to start the server.
type Engine struct {
	RouterGroup
}

// New creates a new Engine using the given Router.
func New(r Router) *Engine {
	return &Engine{
		RouterGroup: &routerGroup{
			router:     r,
			middleware: alice.New(setup),
		},
	}
}

// Run starts kumi.
func (e *Engine) Run(addr string) error {
	return e.Serve(&http.Server{Addr: addr})
}

// RunTLS starts kumi with a given TLS config.
func (e *Engine) RunTLS(addr string, config *tls.Config) error {
	return e.Serve(&http.Server{
		Addr:      addr,
		TLSConfig: config,
	})
}

// Serve takes one or more http.Server structs and serves those.
// Note: The handler will be set for you if you don't provide one.
// If you are using TLS, http2 will automatically be configured as well.
func (e *Engine) Serve(servers ...*http.Server) error {
	e.prep(servers...)

	return gracehttp.Serve(servers...)
}

// prep breaks out all of the steps of Serve except actually calling
// gracehttp.Serve so we can test.
func (e *Engine) prep(servers ...*http.Server) {
	hasHandler := true
	for _, s := range servers {
		if s.Handler == nil {
			hasHandler = false
			s.Handler = http.DefaultServeMux
		}
	}

	if !hasHandler {
		http.Handle("/", e.RouterGroup)
	}
}

// setup is internal kumi middleware. It wraps http.ResponseWriter with
// ResponseWriter, or with BodylessResponseWriter for HEAD requests.
// It normalizes the Host and sets the URL scheme. In addition, this
// sets the RequestContext.
func setup(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case HEAD:
			w = &BodylessResponseWriter{w}
		default:
			rw := newWriter(w)
			w = rw
			defer writerPool.Put(rw)
		}

		r.Host = strings.ToLower(r.Host)
		r.URL.Host = r.Host
		if r.TLS != nil {
			r.URL.Scheme = "https"
		} else {
			r.URL.Scheme = "http"
		}

		// Set the kumi request context
		rc := newRequestContext(r)
		defer returnContext(rc)
		if p, ok := getParams(r); ok {
			rc.params = p
		}

		next.ServeHTTP(w, setRequestContext(r, rc))
	})
}
