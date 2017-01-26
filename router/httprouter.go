package router

import (
	"net/http"
	"strings"

	"github.com/cristiangraz/kumi"
	"github.com/julienschmidt/httprouter"
)

// HTTPRouter wraps the httprouter.router router and meets the
// kumi.Router interface.
type HTTPRouter struct {
	router *httprouter.Router
}

var _ kumi.Router = &HTTPRouter{}

// NewHTTPRouter creates a new instance of HTTPRouter.
func NewHTTPRouter() *HTTPRouter {
	return &HTTPRouter{
		router: httprouter.New(),
	}
}

// Handle implements httprouter.Handler and converts the params to Params accessible
// in the RequestContext.
func (router *HTTPRouter) Handle(method string, pattern string, next http.Handler) {
	router.router.Handle(method, pattern, func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		if len(params) > 0 {
			p := make(map[string]string, len(params))
			for _, v := range params {
				p[v.Key] = v.Value
			}
			r = kumi.SetParams(r, p)
		}

		next.ServeHTTP(w, r)
	})
}

// ServeHTTP calls httprouter's ServeHTTP method.
func (router *HTTPRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.router.ServeHTTP(w, r)
}

// NotFoundHandler registers a handler to execute when no route is matched.
func (router *HTTPRouter) NotFoundHandler(h http.Handler) {
	router.router.NotFound = h
}

// MethodNotAllowedHandler registers a handler to execute when the requested
// method is not allowed.
func (router *HTTPRouter) MethodNotAllowedHandler(h http.Handler) {
	router.router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods := make([]string, 0, len(kumi.HTTPMethods))
		for _, m := range kumi.HTTPMethods {
			if h, _, _ := router.router.Lookup(m, r.URL.Path); h != nil {
				methods = append(methods, m)
			}
		}
		w.Header().Set("Allow", strings.Join(methods, ", "))
		h.ServeHTTP(w, r)
	})
}

// HasRoute returns true if the router has registered a route with that
// method and pattern.
func (router *HTTPRouter) HasRoute(method string, path string) bool {
	h, _, _ := router.router.Lookup(method, path)
	return h != nil
}
