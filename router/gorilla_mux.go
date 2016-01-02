package router

import (
	"net/http"
	"strings"

	"github.com/cristiangraz/kumi"
	"github.com/gorilla/mux"
)

type (
	// GorillaMuxRouter wraps the mux.Router router and meets the
	// kumi.Router interface.
	GorillaMuxRouter struct {
		Router   *mux.Router
		engine   *kumi.Engine
		routes   map[string][]string
		notFound []kumi.HandlerFunc
	}
)

// NewGorillaMuxRouter creates a new instance of a default mux.Router.
// If you need to set custom options, you should instantiate GorillaMuxRouter
// yourself.
func NewGorillaMuxRouter() *GorillaMuxRouter {
	r := map[string][]string{}
	for _, m := range kumi.HTTPMethods {
		r[m] = []string{}
	}

	return &GorillaMuxRouter{
		Router: mux.NewRouter(),
		routes: r,
	}
}

// Handle ...
func (router GorillaMuxRouter) Handle(method string, pattern string, h ...kumi.HandlerFunc) {
	router.Router.HandleFunc(pattern, func(rw http.ResponseWriter, r *http.Request) {
		e := router.Engine()
		c := e.NewContext(rw, r, h...)
		defer e.ReturnContext(c)

		if p := mux.Vars(r); len(p) > 0 {
			c.Params = kumi.Params(p)
		}

		c.Next()
	}).Methods(method)

	router.routes[method] = append(router.routes[method], pattern)
}

// SetEngine sets the kumi engine on the router.
func (router *GorillaMuxRouter) SetEngine(e *kumi.Engine) {
	router.engine = e
}

// Engine retrieves the kumi engine.
func (router GorillaMuxRouter) Engine() *kumi.Engine {
	return router.engine
}

// ServeHTTP ...
func (router GorillaMuxRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	router.Router.ServeHTTP(rw, r)
}

// NotFoundHandler ...
func (router *GorillaMuxRouter) NotFoundHandler(h ...kumi.HandlerFunc) {
	router.notFound = h
	router.Router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		e := router.Engine()
		c := e.NewContext(rw, r, h...)
		defer e.ReturnContext(c)

		c.Next()
	})
}

// MethodNotAllowedHandler registers handlers to respond to Method Not
// Allowed requests. Because Gorilla Mux does not support this natively,
// this method registers a NotFoundHandler that looks for route matches
// to determine if the 404 has matches against other methods. If so,
// the MethodNotAllowed handlers run. Otherwise, the NotFound handlers run.
func (router *GorillaMuxRouter) MethodNotAllowedHandler(h ...kumi.HandlerFunc) {
	router.Router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		methods := router.getMethods(r)
		if len(methods) > 0 {
			// At least one match against an HTTP method. 405 Not Allowed
			e := router.Engine()
			c := e.NewContext(rw, r, h...)
			defer e.ReturnContext(c)

			c.Header().Set("Allow", strings.Join(methods, ", "))

			c.Next()
		} else {
			// 404
			if len(router.notFound) > 0 {
				// 404 handler is defined by user
				e := router.Engine()
				c := e.NewContext(rw, r, router.notFound...)
				defer e.ReturnContext(c)

				c.Next()
			} else {
				// Fallback 404
				e := router.Engine()
				c := e.NewContext(rw, r, func(c *kumi.Context) {
					http.NotFoundHandler().ServeHTTP(c, c.Request)
				})
				defer e.ReturnContext(c)

				c.Next()
			}
		}
	})
}

// getMethods ...
func (router *GorillaMuxRouter) getMethods(r *http.Request) (methods []string) {
	match := &mux.RouteMatch{}
	var reqCopy http.Request
	for _, m := range kumi.HTTPMethods {
		reqCopy = *r
		reqCopy.Method = m

		if router.Router.Match(&reqCopy, match) {
			methods = append(methods, m)
		}
	}

	return methods
}

// HasRoute returns true if the router has registered a route with that
// method and pattern.
func (router *GorillaMuxRouter) HasRoute(method string, pattern string) bool {
	for _, p := range router.routes[method] {
		if p == pattern {
			return true
		}
	}

	return false
}
