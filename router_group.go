package kumi

import "net/http"

type (
	// Router defines an interface that allows for interchangeable routers.
	Router interface {
		Handle(method string, pattern string, handlers ...HandlerFunc)
		ServeHTTP(http.ResponseWriter, *http.Request)
		SetEngine(*Engine)
		Engine() *Engine
		NotFoundHandler(...HandlerFunc)
	}

	// RouterGroup allows for grouping routes by a base pattern (path) and shared middleware.
	RouterGroup struct {
		pattern  string
		router   Router
		Handlers []HandlerFunc
		engine   *Engine
	}
)

// Group creates a sub-group of the router based on a route prefix. Any middleware
// added to the group will be appended to the parent's middleware
func (g RouterGroup) Group(pattern string, handlers ...Handler) RouterGroup {
	wrapped, err := wrapHandlers(handlers...)
	if err != nil {
		panic(err)
	}

	return RouterGroup{
		pattern:  pattern,
		router:   g.router,
		Handlers: appendHandlers(g.Handlers, wrapped...),
	}
}

// Use adds middleware to any routes used in this RouterGroup.
func (g *RouterGroup) Use(handlers ...Handler) {
	wrapped, err := wrapHandlers(handlers...)
	if err != nil {
		panic(err)
	}

	g.Handlers = appendHandlers(g.Handlers, wrapped...)
}

// Get defines an HTTP GET endpoint with one or more handlers.
func (g RouterGroup) Get(pattern string, handlers ...Handler) {
	g.handle("GET", pattern, handlers...)
}

// Post defines an HTTP POST endpoint with one or more handlers.
func (g RouterGroup) Post(pattern string, handlers ...Handler) {
	g.handle("POST", pattern, handlers...)
}

// Put defines an HTTP PUT endpoint with one or more handlers.
func (g RouterGroup) Put(pattern string, handlers ...Handler) {
	g.handle("PUT", pattern, handlers...)
}

// Patch defines an HTTP PATCH endpoint with one or more handlers.
func (g RouterGroup) Patch(pattern string, handlers ...Handler) {
	g.handle("PATCH", pattern, handlers...)
}

// Head defines an HTTP HEAD endpoint with one or more handlers.
func (g RouterGroup) Head(pattern string, handlers ...Handler) {
	g.handle("HEAD", pattern, handlers...)
}

// Options defines an HTTP OPTIONS endpoint with one or more handlers.
func (g RouterGroup) Options(pattern string, handlers ...Handler) {
	g.handle("OPTIONS", pattern, handlers...)
}

// Delete defines an HTTP DELETE endpoint with one or more handlers.
func (g RouterGroup) Delete(pattern string, handlers ...Handler) {
	g.handle("DELETE", pattern, handlers...)
}

// All is a convenience function that adds a handler to
// GET/HEAD/POST/PUT/PATCH/DELETE methods.
func (g RouterGroup) All(pattern string, handlers ...Handler) {
	for _, method := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"} {
		g.handle(method, pattern, handlers...)
	}
}

// NotFoundHandler ...
func (g RouterGroup) NotFoundHandler(handlers ...Handler) {
	wrapped, err := wrapHandlers(handlers...)
	if err != nil {
		panic(err)
	}

	g.router.NotFoundHandler(appendHandlers(g.Handlers, wrapped...)...)
}

// handle consolidates all of the middleware into a route that satisfies the
// router.Handle interface
func (g RouterGroup) handle(method, pattern string, handlers ...Handler) {
	pattern = g.pattern + pattern
	wrapped, err := wrapHandlers(handlers...)
	if err != nil {
		panic(err)
	}

	g.router.Handle(method, pattern, appendHandlers(g.Handlers, wrapped...)...)
}

// appendHandlers extends a chain, adding the specified handlers
// as the last ones in the request flow. appendHandlers returns a new chain,
// leaving the original one untouched.
func appendHandlers(h []HandlerFunc, append ...HandlerFunc) []HandlerFunc {
	handlers := make([]HandlerFunc, len(h)+len(append))
	copy(handlers, h)
	copy(handlers[len(h):], append)

	return handlers
}
