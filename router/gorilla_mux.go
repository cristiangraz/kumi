package router

import (
	"net/http"
	"strings"

	"github.com/cristiangraz/kumi"
	"github.com/gorilla/mux"
)

// GorillaMuxRouter wraps the mux.Router router and meets the
// kumi.Router interface.
type GorillaMuxRouter struct {
	router   *mux.Router
	store    *Store
	notFound http.Handler
}

var _ kumi.Router = &GorillaMuxRouter{}

// NewGorillaMuxRouter creates a new instance of a default mux.Router.
// If you need to set custom options, you should instantiate GorillaMuxRouter
// yourself.
func NewGorillaMuxRouter() *GorillaMuxRouter {
	return &GorillaMuxRouter{
		router: mux.NewRouter(),
		store:  &Store{},
	}
}

// Handle ...
func (router *GorillaMuxRouter) Handle(method string, pattern string, next http.Handler) {
	router.router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if p := mux.Vars(r); len(p) > 0 {
			r = kumi.SetParams(r, p)
		}
		next.ServeHTTP(w, r)
	}).Methods(method)

	router.store.Save(method, pattern)
}

// ServeHTTP ...
func (router *GorillaMuxRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.router.ServeHTTP(w, r)
}

// NotFoundHandler ...
func (router *GorillaMuxRouter) NotFoundHandler(h http.Handler) {
	router.notFound = h
	router.router.NotFoundHandler = h
}

// MethodNotAllowedHandler registers handlers to respond to Method Not
// Allowed requests. Because Gorilla Mux does not support this natively,
// this method registers a NotFoundHandler that looks for route matches
// to determine if the 404 has matches against other methods. If so,
// the MethodNotAllowed handlers run. Otherwise, the NotFound handlers run.
func (router *GorillaMuxRouter) MethodNotAllowedHandler(next http.Handler) {
	router.router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods := router.getMethods(r)
		if len(methods) > 0 {
			w.Header().Set("Allow", strings.Join(methods, ", "))
		} else {
			// 404
			if router.notFound != nil {
				// 404 handler is defined by user
				next = router.notFound
			} else {
				// Fallback 404
				http.NotFoundHandler().ServeHTTP(w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// getMethods ...
func (router *GorillaMuxRouter) getMethods(r *http.Request) (methods []string) {
	var reqCopy http.Request
	for _, m := range kumi.HTTPMethods {
		reqCopy = *r
		reqCopy.Method = m

		var routeMatch mux.RouteMatch
		if router.router.Match(&reqCopy, &routeMatch) && routeMatch.Route != nil {
			methods = append(methods, m)
		}
	}

	return methods
}

// HasRoute returns true if the router has registered a route with that
// method and pattern.
func (router *GorillaMuxRouter) HasRoute(method string, pattern string) bool {
	return router.store.HasRoute(method, pattern)
}
