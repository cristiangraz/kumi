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
	store  *Store
}

var _ kumi.Router = &HTTPTreeMux{}

// NewHTTPTreeMux creates a new instance of a default httptreemux router.
// If you need to set custom options, you should instantiate HTTPTreeMux
// yourself.
func NewHTTPTreeMux() *HTTPTreeMux {
	return &HTTPTreeMux{
		router: httptreemux.New(),
		store:  &Store{},
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
	router.store.Save(method, pattern)
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
func (router *HTTPTreeMux) HasRoute(method string, pattern string) bool {
	return router.store.HasRoute(method, pattern)
}
