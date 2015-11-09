package router

import (
	"net/http"

	"github.com/cristiangraz/kumi"
	"github.com/julienschmidt/httprouter"
)

type (
	// HTTPRouter wraps the httprouter.Router router and meets the
	// kumi.Router interface.
	HTTPRouter struct {
		Router *httprouter.Router
		engine *kumi.Engine
	}
)

// NewHTTPRouter creates a new instance of a default httptreemux router.
// If you need to set custom options, you should instantiate HTTPRouter
// yourself.
func NewHTTPRouter() *HTTPRouter {
	return &HTTPRouter{
		Router: httprouter.New(),
	}
}

// Handle ...
func (router HTTPRouter) Handle(method string, path string, h ...kumi.HandlerFunc) {
	router.Router.Handle(method, path, func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
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
