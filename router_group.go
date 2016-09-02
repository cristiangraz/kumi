package kumi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/justinas/alice"
)

// HTTPMethods provides an array of HTTP methods for looping over.
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

// MiddlewareFn is a type for handling upstream/downstream middleware.
type MiddlewareFn func(http.Handler) http.Handler

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
	Group(pattern string, middleware ...MiddlewareFn) RouterGroup
	Use(middleware ...MiddlewareFn)
	Get(pattern string, handlers ...Handler)
	Post(pattern string, handlers ...Handler)
	Put(pattern string, handlers ...Handler)
	Patch(pattern string, handlers ...Handler)
	Head(pattern string, handlers ...Handler)
	Options(pattern string, handlers ...Handler)
	Delete(pattern string, handlers ...Handler)
	All(pattern string, handlers ...Handler)
	NotFoundHandler(http.HandlerFunc)
	MethodNotAllowedHandler(http.HandlerFunc)

	// SetCors sets a middleware to handle CORS headers.
	// This ensures OPTIONS endpoints are automatically created if not defined,
	// and that NotFound endpoints return CORS headers.
	SetCors(MiddlewareFn)

	ServeHTTP(http.ResponseWriter, *http.Request)
}

// routerGroup implements RouterGroup.
type routerGroup struct {
	pattern    string
	router     Router
	middleware alice.Chain
	cors       MiddlewareFn
}

var _ RouterGroup = &routerGroup{}

// Group creates a sub-group of the router based on a route prefix. Any middleware
// added to the group will be appended to the parent's middleware.
func (g *routerGroup) Group(pattern string, middleware ...MiddlewareFn) RouterGroup {
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
func (g *routerGroup) Use(middleware ...MiddlewareFn) {
	c := make([]alice.Constructor, len(middleware))
	for i := range middleware {
		c[i] = alice.Constructor(middleware[i])
	}

	g.middleware = g.middleware.Append(c...)
}

// Get defines an HTTP GET endpoint with one or more handlers.
// It will also register a HEAD endpoint. Kumi will automatically
// use a bodyless response writer.
func (g *routerGroup) Get(pattern string, handlers ...Handler) {
	g.handle(GET, pattern, handlers...)
}

// Post defines an HTTP POST endpoint with one or more handlers.
func (g *routerGroup) Post(pattern string, handlers ...Handler) {
	g.handle(POST, pattern, handlers...)
}

// Put defines an HTTP PUT endpoint with one or more handlers.
func (g *routerGroup) Put(pattern string, handlers ...Handler) {
	g.handle(PUT, pattern, handlers...)
}

// Patch defines an HTTP PATCH endpoint with one or more handlers.
func (g *routerGroup) Patch(pattern string, handlers ...Handler) {
	g.handle(PATCH, pattern, handlers...)
}

// Head defines an HTTP HEAD endpoint with one or more handlers.
// Kumi defines this automatically for all GET routes. If you want
// to define your own Head handler, define it before defining
// the Get handler for the same pattern.
func (g *routerGroup) Head(pattern string, handlers ...Handler) {
	g.handle(HEAD, pattern, handlers...)
}

// Options defines an HTTP OPTIONS endpoint with one or more handlers.
// If you are using CORS, Kumi defines this automatically for all routes.
// If you want to define your own Options handler, define it before defining
// other methods against the same pattern.
func (g *routerGroup) Options(pattern string, handlers ...Handler) {
	g.handle(OPTIONS, pattern, handlers...)
}

// Delete defines an HTTP DELETE endpoint with one or more handlers.
func (g *routerGroup) Delete(pattern string, handlers ...Handler) {
	g.handle(DELETE, pattern, handlers...)
}

// All is a convenience function that adds a handler to
// GET/HEAD/POST/PUT/PATCH/DELETE methods.
// Note HEAD/OPTIONS are set in the handle method automatically.
func (g *routerGroup) All(pattern string, handlers ...Handler) {
	for _, method := range HTTPMethods {
		g.handle(method, pattern, handlers...)
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

// SetCors sets the MiddlewareFn that handles CORS headers.
// This is registered independendently so kumi can handle some CORS
// conveniences for the application (creating OPTIONS routes and running
// CORS on 404 requests).
func (g *routerGroup) SetCors(m MiddlewareFn) {
	g.cors = m
}

// ServeHTTP ...
func (g *routerGroup) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.router.ServeHTTP(w, r)
}

// handle consolidates all of the middleware into a route that satisfies the
// router.Handle interface
func (g *routerGroup) handle(method, pattern string, handlers ...Handler) {
	middleware, handler, err := combine(handlers...)
	if err != nil {
		panic(err)
	}

	h := g.middleware.Append(middleware...).ThenFunc(handler)
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

// combine takes one or more handlers and combines them into an http.Handler.
// combine expects at least one handler. The last handler should always be http.HandlerFunc.
// One or more middleware can be combined before the final handler.
func combine(handlers ...Handler) ([]alice.Constructor, http.HandlerFunc, error) {
	if len(handlers) == 0 {
		return nil, nil, errors.New("one or more handlers are required")
	}

	handler := handlers[len(handlers)-1]
	middleware := handlers[:len(handlers)-1]

	var h http.HandlerFunc
	switch handler.(type) {
	case http.HandlerFunc:
		h = handler.(http.HandlerFunc)
	case func(http.ResponseWriter, *http.Request):
		h = http.HandlerFunc(handler.(func(http.ResponseWriter, *http.Request)))
	default:
		return nil, nil, fmt.Errorf("invalid handler: %T", handler)
	}

	if len(handlers) == 1 {
		return nil, h, nil
	}

	c := make([]alice.Constructor, len(middleware))
	for i := range middleware {
		switch mw := middleware[i].(type) {
		case func(http.Handler) http.Handler:
			c[i] = mw
		case MiddlewareFn:
			c[i] = alice.Constructor(mw)
		default:
			panic(fmt.Sprintf("invalid middleware: %T", mw))
		}
	}
	return c, h, nil
}

// MiddlewareFunc wraps an http.HandlerFunc so it implements MiddlewareFn.
// Do not use this if you are wrapping ResponseWriter or using r.WithContext -
// both values need to be passed to fn.ServeHTTP in order to be accessible downstream.
func MiddlewareFunc(fn http.HandlerFunc) MiddlewareFn {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(w, r)
			next.ServeHTTP(w, r)
		})
	}
}
