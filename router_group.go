package kumi

import (
	"net/http"

	"github.com/justinas/alice"
)

// HTTPMethods is a list of HTTP methods kumi supports.
var HTTPMethods = []string{GET, HEAD, POST, PUT, PATCH, OPTIONS, DELETE}

// HTTP method constants.
const (
	GET     = "GET"
	HEAD    = "HEAD"
	POST    = "POST"
	PUT     = "PUT"
	PATCH   = "PATCH"
	DELETE  = "DELETE"
	OPTIONS = "OPTIONS"
)

// Handler is a generic HTTP handler.
type Handler interface{}

// Router defines an interface that allows for interchangeable routers.
type Router interface {
	Handle(method string, pattern string, handler http.Handler)
	ServeHTTP(http.ResponseWriter, *http.Request)
	NotFoundHandler(http.Handler)

	// MethodNotAllowedHandler registers handlers for MethodNotAllowed
	// responses. The router is responsible for setting the Allow response
	// header here.
	MethodNotAllowedHandler(http.Handler)
	HasRoute(method string, pattern string) bool
}

// RouterGroup wraps the Router interface to provide route grouping by
// a base pattern path and shared middleware.
type RouterGroup interface {
	// Group generates a new RouterGroup from the current RouterGroup that
	// shares middleware. Any middleware appended to the
	// group will be appended to it's parent's middleware.
	Group(middleware ...func(http.Handler) http.Handler) RouterGroup

	// GroupPath generates a new RouterGroup from the current RouterGroup that
	// shares a base path and middleware. Any middleware appended to the
	// group will be appended to it's parent's middleware.
	GroupPath(pattern string, middleware ...func(http.Handler) http.Handler) RouterGroup

	// Use applies middleware to all handlers defined after the Use call in
	// this RouterGroup or it's descendants.
	Use(middleware ...func(http.Handler) http.Handler)

	// Defines a handler and optional middleware for a GET request at pattern.
	Get(pattern string, handler http.HandlerFunc)

	// Defines a handler and optional middleware for a POST request at pattern.
	Post(pattern string, handler http.HandlerFunc)

	// Defines a handler and optional middleware for a PUT request at pattern.
	Put(pattern string, handler http.HandlerFunc)

	// Defines a handler and optional middleware for a PATCH request at pattern.
	Patch(pattern string, handler http.HandlerFunc)

	// Defines a handler and optional middleware for a HEAD request at pattern.
	// Kumi defines this automatically for all GET routes. If you want
	// to define your own Head handler, define it before defining
	// the Get handler for the same pattern.
	Head(pattern string, handler http.HandlerFunc)

	// Defines a handler and optional middleware for a OPTIONS request at pattern.
	Options(pattern string, handler http.HandlerFunc)

	// Defines a handler and optional middleware for a DELETE request at pattern.
	Delete(pattern string, handler http.HandlerFunc)

	// Defines a handler and optional middleware for all
	// HTTP method requests at pattern.
	All(pattern string, handler http.HandlerFunc)

	// NotFoundHandler registers a handler to run when no matching route is found.
	NotFoundHandler(http.HandlerFunc)

	// MethodNotAllowedHandler registers a handler to run when a route is valid,
	// but not for the requested HTTP method.
	MethodNotAllowedHandler(http.HandlerFunc)

	// SetCors sets a middleware to handle CORS headers.
	// This ensures OPTIONS endpoints are automatically created if not defined,
	// and that NotFound endpoints return CORS headers.
	SetCors(func(http.Handler) http.Handler)

	// ServeHTTP implements the http.Handler interface.
	ServeHTTP(http.ResponseWriter, *http.Request)
}

// routerGroup implements RouterGroup.
type routerGroup struct {
	pattern    string
	router     Router
	middleware alice.Chain
	cors       func(http.Handler) http.Handler
}

var _ RouterGroup = &routerGroup{}

// Group creates a sub-group of the router based on a route prefix. Any middleware
// added to the group will be appended to the parent's middleware.
func (g *routerGroup) Group(middleware ...func(http.Handler) http.Handler) RouterGroup {
	c := make([]alice.Constructor, len(middleware))
	for i := range middleware {
		c[i] = alice.Constructor(middleware[i])
	}

	return &routerGroup{
		router:     g.router,
		middleware: g.middleware.Append(c...),
	}
}

// GroupPath creates a sub-group of the router based on a route prefix. Any middleware
// added to the group will be appended to the parent's middleware.
func (g *routerGroup) GroupPath(pattern string, middleware ...func(http.Handler) http.Handler) RouterGroup {
	c := make([]alice.Constructor, len(middleware))
	for i := range middleware {
		c[i] = alice.Constructor(middleware[i])
	}

	return &routerGroup{
		pattern:    pattern,
		router:     g.router,
		middleware: g.middleware.Append(c...),
	}
}

// Use adds middleware to any routes used in this RouterGroup.
func (g *routerGroup) Use(middleware ...func(http.Handler) http.Handler) {
	c := make([]alice.Constructor, len(middleware))
	for i := range middleware {
		c[i] = alice.Constructor(middleware[i])
	}

	g.middleware = g.middleware.Append(c...)
}

// Get defines an HTTP GET endpoint with one or more handlers.
// It will also register a HEAD endpoint. Kumi will automatically
// use a bodyless response writer.
func (g *routerGroup) Get(pattern string, handler http.HandlerFunc) {
	g.handle(GET, pattern, handler)
}

// Post defines an HTTP POST endpoint with one or more handlers.
func (g *routerGroup) Post(pattern string, handler http.HandlerFunc) {
	g.handle(POST, pattern, handler)
}

// Put defines an HTTP PUT endpoint with one or more handlers.
func (g *routerGroup) Put(pattern string, handler http.HandlerFunc) {
	g.handle(PUT, pattern, handler)
}

// Patch defines an HTTP PATCH endpoint with one or more handlers.
func (g *routerGroup) Patch(pattern string, handler http.HandlerFunc) {
	g.handle(PATCH, pattern, handler)
}

// Head defines an HTTP HEAD endpoint with one or more handlers.
// Kumi defines this automatically for all GET routes. If you want
// to define your own Head handler, define it before defining
// the Get handler for the same pattern.
func (g *routerGroup) Head(pattern string, handler http.HandlerFunc) {
	g.handle(HEAD, pattern, handler)
}

// Options defines an HTTP OPTIONS endpoint with one or more handlers.
// If you are using CORS, Kumi defines this automatically for all routes.
// If you want to define your own Options handler, define it before defining
// other methods against the same pattern.
func (g *routerGroup) Options(pattern string, handler http.HandlerFunc) {
	g.handle(OPTIONS, pattern, handler)
}

// Delete defines an HTTP DELETE endpoint with one or more handlers.
func (g *routerGroup) Delete(pattern string, handler http.HandlerFunc) {
	g.handle(DELETE, pattern, handler)
}

// All is a convenience function that adds a handler to
// GET/HEAD/POST/PUT/PATCH/DELETE methods.
// Note HEAD/OPTIONS are set in the handle method automatically.
func (g *routerGroup) All(pattern string, handler http.HandlerFunc) {
	for _, method := range HTTPMethods {
		g.handle(method, pattern, handler)
	}
}

// NotFoundHandler runs when no route is found.
// inhermitMiddleware determines if the global and group middleware chain
// should run on a not found request. You can optionally set to false and
// include a custom middleware chain in the handlers parameters.
func (g *routerGroup) NotFoundHandler(handler http.HandlerFunc) {
	// TODO: If middleware is inherited, won't this run automatically?
	if g.cors != nil {
		g.router.NotFoundHandler(g.middleware.Append(alice.Constructor(g.cors)).ThenFunc(handler))
		return
	}
	g.router.NotFoundHandler(g.middleware.ThenFunc(handler))
}

// MethodNotAllowedHandler runs when a route exists at the current
// path -- but not for the request method used.
// inhermitMiddleware determines if the global and group middleware chain
// should run on a method not allowed request. You can optionally set to
// false and include a custom middleware chain in the handlers parameters.
func (g *routerGroup) MethodNotAllowedHandler(handler http.HandlerFunc) {
	g.router.MethodNotAllowedHandler(g.middleware.ThenFunc(handler))
}

// SetCors sets the func(http.Handler) http.Handler that handles CORS headers.
// This is registered independendently so kumi can handle some CORS
// conveniences for the application (creating OPTIONS routes and running
// CORS on 404 requests).
func (g *routerGroup) SetCors(m func(http.Handler) http.Handler) {
	g.cors = m
}

// ServeHTTP ...
func (g *routerGroup) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.router.ServeHTTP(w, r)
}

// handle consolidates all of the middleware into a route that satisfies the
// router.Handle interface
func (g *routerGroup) handle(method, pattern string, handler http.HandlerFunc) {
	if handler == nil {
		panic("cannot send a nil http.HandlerFunc")
	}

	h := g.middleware.ThenFunc(handler)
	pattern = g.pattern + pattern

	g.router.Handle(method, pattern, h)

	// Add HEAD to all GET routes if no route is already defined.
	if method == GET && !g.router.HasRoute(HEAD, pattern) {
		g.router.Handle(HEAD, pattern, h)
	}

	// Add OPTIONS to all CORS routes if no route is already defined.
	if g.cors != nil && method != OPTIONS && !g.router.HasRoute(OPTIONS, pattern) {
		g.router.Handle(OPTIONS, pattern, h)
	}
}

// MiddlewareFunc wraps an http.HandlerFunc so it implements func(http.Handler) http.Handler.
// Do not use this if you are wrapping ResponseWriter or using r.WithContext -
// both values need to be passed to fn.ServeHTTP in order to be accessible downstream.
func MiddlewareFunc(fn http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r)
			next.ServeHTTP(w, r)
		})
	}
}
