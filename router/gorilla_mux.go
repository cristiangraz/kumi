package router

import (
	"net/http"

	"github.com/cristiangraz/kumi"
	"github.com/gorilla/mux"
)

type (
	// GorillaMuxRouter wraps the mux.Router router and meets the
	// kumi.Router interface.
	GorillaMuxRouter struct {
		Router *mux.Router
		engine *kumi.Engine
		routes map[string][]string
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
func (router GorillaMuxRouter) NotFoundHandler(h ...kumi.HandlerFunc) {
	router.Router.NotFoundHandler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		e := router.Engine()
		c := e.NewContext(rw, r, h...)
		defer e.ReturnContext(c)

		c.Next()
	})
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
