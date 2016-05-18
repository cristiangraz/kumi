package kumi

import (
	"crypto/tls"
	"net/http"
	"strings"
	"sync"

	"github.com/facebookgo/grace/gracehttp"
	"golang.org/x/net/context"
)

// Engine is the glue that holds everything together.
type Engine struct {
	RouterGroup

	// DefaultContext is the starting context used for each request.
	DefaultContext context.Context

	pool sync.Pool

	// Global CORS settings. This is used only if you attach the
	// Engine.CorsOptions handler to your handler chain of
	// each route or route group.
	// Additionally you can provide route-specific overrides
	// for any of the settings in CorsOptions.
	cors *CorsOptions
}

// BodylessResponseWriter wraps http.ResponseWriter, discarding
// anything written to the body.
type BodylessResponseWriter struct {
	http.ResponseWriter
}

// New creates a new Engine using the given Router.
func New(r Router) *Engine {
	e := &Engine{
		RouterGroup:    RouterGroup{},
		DefaultContext: context.Background(),
		pool: sync.Pool{
			New: func() interface{} {
				return newContext(nil, nil, nil)
			},
		},
	}

	r.SetEngine(e)
	e.RouterGroup.router = r

	return e
}

// NewContext retrieves a context from the pool and sets it for the request.
func (e *Engine) NewContext(rw http.ResponseWriter, r *http.Request, handlers ...HandlerFunc) *Context {
	if r.Method == HEAD {
		rw = &BodylessResponseWriter{rw}
	}

	c := e.pool.Get().(*Context)
	c.reset(rw, r, handlers...)

	// Set the starting context.Context from the engine's default
	c.Context = e.DefaultContext

	// Set URL host and scheme
	r.Host = strings.ToLower(r.Host)
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
	if r.TLS != nil {
		r.URL.Scheme = "https"
	}

	return c
}

// ReturnContext returns a context back to the pool. Before the context is
// returned deferred functions are run. This is automatically done
// by each of the router implementations and should only be used if you are
// integrating a route with Kumi.
func (e *Engine) ReturnContext(c *Context) {
	e.pool.Put(c)
}

// ServeHTTP is the ServeHTTP call for the router. Useful for testing.
func (e *Engine) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	e.router.ServeHTTP(rw, r)
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

func (brw BodylessResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
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
		http.Handle("/", e.router)
	}
}
