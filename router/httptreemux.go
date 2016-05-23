package router

import (
	"net/http"
	"strings"

	"github.com/cristiangraz/kumi"
	"github.com/dimfeld/httptreemux"
)

// HTTPTreeMux wraps the httptreemux.TreeMux router and meets the
// kumi.Router interface.
type HTTPTreeMux struct {
	Router *httptreemux.TreeMux
	engine *kumi.Engine
	routes map[string][]string
}

// NewHTTPTreeMux creates a new instance of a default httptreemux router.
// If you need to set custom options, you should instantiate HTTPTreeMux
// yourself.
func NewHTTPTreeMux() *HTTPTreeMux {
	r := map[string][]string{}
	for _, m := range kumi.HTTPMethods {
		r[m] = []string{}
	}

	return &HTTPTreeMux{
		Router: httptreemux.New(),
		routes: r,
	}
}

// Handle ...
func (router *HTTPTreeMux) Handle(method string, pattern string, h ...kumi.HandlerFunc) {
	router.Router.Handle(method, pattern, func(rw http.ResponseWriter, r *http.Request, p map[string]string) {
		c := router.engine.NewContext(rw, r, h...)
		defer router.engine.ReturnContext(c)

		if len(p) > 0 {
			c.Params = kumi.Params(p)
		}

		c.Next()
	})

	router.routes[method] = append(router.routes[method], pattern)
}

// SetEngine sets the kumi engine on the router.
func (router *HTTPTreeMux) SetEngine(e *kumi.Engine) {
	router.engine = e
}

// Engine retrieves the kumi engine.
func (router *HTTPTreeMux) Engine() *kumi.Engine {
	return router.engine
}

// ServeHTTP ...
func (router *HTTPTreeMux) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	router.Router.ServeHTTP(rw, r)
}

// NotFoundHandler ...
func (router *HTTPTreeMux) NotFoundHandler(h ...kumi.HandlerFunc) {
	router.Router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c := router.engine.NewContext(rw, r, h...)
		defer router.engine.ReturnContext(c)

		c.Next()
	})
}

// MethodNotAllowedHandler ...
func (router *HTTPTreeMux) MethodNotAllowedHandler(h ...kumi.HandlerFunc) {
	router.Router.MethodNotAllowedHandler = func(rw http.ResponseWriter, r *http.Request, methods map[string]httptreemux.HandlerFunc) {
		c := router.engine.NewContext(rw, r, h...)
		defer router.engine.ReturnContext(c)

		var allow []string
		for m := range methods {
			allow = append(allow, m)
		}

		c.Header().Set("Allow", strings.Join(allow, ", "))

		c.Next()
	}
}

// HasRoute ...
func (router *HTTPTreeMux) HasRoute(method string, pattern string) bool {
	for _, p := range router.routes[method] {
		if p == pattern {
			return true
		}
	}

	return false
}
