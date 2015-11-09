package router

import (
	"net/http"

	"github.com/cristiangraz/kumi"
	"github.com/dimfeld/httptreemux"
)

type (
	// HTTPTreeMux wraps the httptreemux.TreeMux router and meets the
	// kumi.Router interface.
	HTTPTreeMux struct {
		Router *httptreemux.TreeMux
		engine *kumi.Engine
	}
)

// NewHTTPTreeMux creates a new instance of a default httptreemux router.
// If you need to set custom options, you should instantiate HTTPTreeMux
// yourself.
func NewHTTPTreeMux() *HTTPTreeMux {
	return &HTTPTreeMux{
		Router: httptreemux.New(),
	}
}

// Handle ...
func (router HTTPTreeMux) Handle(method string, path string, h ...kumi.HandlerFunc) {
	router.Router.Handle(method, path, func(rw http.ResponseWriter, r *http.Request, p map[string]string) {
		e := router.Engine()
		c := e.NewContext(rw, r, h...)
		defer e.ReturnContext(c)

		if len(p) > 0 {
			c.Params = kumi.Params(p)
		}

		c.Next()
	})
}

// SetEngine sets the kumi engine on the router.
func (router *HTTPTreeMux) SetEngine(e *kumi.Engine) {
	router.engine = e
}

// Engine retrieves the kumi engine.
func (router HTTPTreeMux) Engine() *kumi.Engine {
	return router.engine
}

// ServeHTTP ...
func (router HTTPTreeMux) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	router.Router.ServeHTTP(rw, r)
}
