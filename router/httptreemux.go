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
	router *httptreemux.TreeMux
}

var _ kumi.Router = &HTTPTreeMux{}

// NewHTTPTreeMux creates a new instance of a default httptreemux router.
// If you need to set custom options, you should instantiate HTTPTreeMux
// yourself.
func NewHTTPTreeMux() *HTTPTreeMux {
	return &HTTPTreeMux{
		router: httptreemux.New(),
	}
}

// Handle ...
func (router *HTTPTreeMux) Handle(method string, pattern string, next http.Handler) {
	router.router.Handle(method, pattern, func(w http.ResponseWriter, r *http.Request, p map[string]string) {
		if len(p) > 0 {
			r = kumi.SetParams(r, p)
		}
		next.ServeHTTP(w, r)
	})
}

// ServeHTTP ...
func (router *HTTPTreeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.router.ServeHTTP(w, r)
}

// NotFoundHandler ...
func (router *HTTPTreeMux) NotFoundHandler(h http.Handler) {
	router.router.NotFoundHandler = func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

// MethodNotAllowedHandler ...
func (router *HTTPTreeMux) MethodNotAllowedHandler(h http.Handler) {
	router.router.MethodNotAllowedHandler = func(w http.ResponseWriter, r *http.Request, methods map[string]httptreemux.HandlerFunc) {
		allow := make([]string, len(methods))
		var i int
		for m := range methods {
			allow[i] = m
			i++
		}
		w.Header().Set("Allow", strings.Join(allow, ", "))

		h.ServeHTTP(w, r)
	}
}

// HasRoute ...
func (router *HTTPTreeMux) HasRoute(method string, path string) bool {
	req, _ := http.NewRequest(method, path, nil)
	_, found := router.router.Lookup(nil, req)
	return found
}
