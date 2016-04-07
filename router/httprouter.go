package router

import (
	"net/http"
	"strings"

	"github.com/cristiangraz/kumi"
	"github.com/julienschmidt/httprouter"
)

// HTTPRouter wraps the httprouter.Router router and meets the
// kumi.Router interface.
type HTTPRouter struct {
	Router *httprouter.Router
	engine *kumi.Engine
}

// NewHTTPRouter creates a new instance of a default httptreemux router.
// If you need to set custom options, you should instantiate HTTPRouter
// yourself.
func NewHTTPRouter() *HTTPRouter {
	return &HTTPRouter{
		Router: httprouter.New(),
	}
}

// Handle ...
func (router *HTTPRouter) Handle(method string, pattern string, h ...kumi.HandlerFunc) {
	router.Router.Handle(method, pattern, func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		e := router.Engine()
		c := e.NewContext(rw, r, h...)
		defer e.ReturnContext(c)

		if len(params) > 0 {
			p := make(map[string]string, len(params))
			for _, v := range params {
				p[v.Key] = v.Value
			}

			c.Params = kumi.Params(p)
		}

		c.Next()
	})
}

// SetEngine sets the kumi engine on the router.
func (router *HTTPRouter) SetEngine(e *kumi.Engine) {
	router.engine = e
}

// Engine retrieves the kumi engine.
func (router *HTTPRouter) Engine() *kumi.Engine {
	return router.engine
}

// ServeHTTP ...
func (router *HTTPRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	router.Router.ServeHTTP(rw, r)
}

// NotFoundHandler ...
func (router *HTTPRouter) NotFoundHandler(h ...kumi.HandlerFunc) {
	router.Router.NotFound = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		e := router.Engine()
		c := e.NewContext(rw, r, h...)
		defer e.ReturnContext(c)

		c.Next()
	})
}

// MethodNotAllowedHandler ...
func (router *HTTPRouter) MethodNotAllowedHandler(h ...kumi.HandlerFunc) {
	router.Router.MethodNotAllowed = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		e := router.Engine()
		c := e.NewContext(rw, r, h...)
		defer e.ReturnContext(c)

		methods := router.getMethods(r)
		c.Header().Set("Allow", strings.Join(methods, ", "))

		c.Next()
	})
}

// getMethods ...
func (router *HTTPRouter) getMethods(r *http.Request) (methods []string) {
	for _, m := range kumi.HTTPMethods {
		if h, _, _ := router.Router.Lookup(m, r.URL.Path); h != nil {
			methods = append(methods, m)
		}
	}

	return methods
}

// HasRoute returns true if the router has registered a route with that
// method and pattern.
func (router *HTTPRouter) HasRoute(method string, pattern string) bool {
	h, _, _ := router.Router.Lookup(method, pattern)
	return h != nil
}
