package router_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/cristiangraz/kumi"
	"github.com/cristiangraz/kumi/router"
)

func TestHTTPRouter(t *testing.T) {
	testRouter(t, routerTest{
		router: func() kumi.Router {
			return router.NewHTTPRouter()
		},
		route:  "/users/:name/:id",
		url:    "/users/httprouter/10",
		params: kumi.Params{"name": "httprouter", "id": "10"},
	})
}

func TestHTTPRouter_NotFoundHandler(t *testing.T) {
	testRouterNotFoundHandler(t, router.NewHTTPRouter())
}

func TestHTTPTreeMux(t *testing.T) {
	testRouter(t, routerTest{
		router: func() kumi.Router {
			return router.NewHTTPTreeMux()
		},
		route:  "/users/:name/:msg",
		url:    "/users/httptreemux/hello",
		params: kumi.Params{"name": "httptreemux", "msg": "hello"},
	})
}

func TestHTTPTreeMux_NotFoundHandler(t *testing.T) {
	testRouterNotFoundHandler(t, router.NewHTTPTreeMux())
}

func TestGorilla(t *testing.T) {
	testRouter(t, routerTest{
		router: func() kumi.Router {
			return router.NewGorillaMuxRouter()
		},
		route:  "/users/{name}",
		url:    "/users/gorilla",
		params: kumi.Params{"name": "gorilla"},
	})
}

func TestGorilla_NotFoundHandler(t *testing.T) {
	testRouterNotFoundHandler(t, router.NewGorillaMuxRouter())
}

type routerTest struct {
	router     func() kumi.Router
	route, url string
	params     kumi.Params
}

func testRouter(t *testing.T, rt routerTest) {
	for _, method := range kumi.HTTPMethods {
		rtr := rt.router()
		k := kumi.New(rtr)
		if found := rtr.HasRoute(method, rt.route); found {
			t.Fatal("no route should be found")
		}

		var ran bool
		h := func(w http.ResponseWriter, r *http.Request) {
			ran = true
			if !reflect.DeepEqual(kumi.Context(r).Params(), rt.params) {
				t.Fatalf("unexpected params: %v", kumi.Context(r).Params())
			}
		}

		switch method {
		case kumi.GET:
			k.Get(rt.route, h)
		case kumi.HEAD:
			k.Head(rt.route, h)
		case kumi.POST:
			k.Post(rt.route, h)
		case kumi.PUT:
			k.Put(rt.route, h)
		case kumi.PATCH:
			k.Patch(rt.route, h)
		case kumi.OPTIONS:
			k.Options(rt.route, h)
		case kumi.DELETE:
			k.Delete(rt.route, h)
		}

		r, _ := http.NewRequest(method, rt.url, nil)
		w := httptest.NewRecorder()
		k.ServeHTTP(w, r)

		if !ran {
			t.Fatalf("expected handler to run")
		} else if found := rtr.HasRoute(method, rt.route); !found {
			t.Fatal("expected route to be found")
		}
	}
}

func testRouterNotFoundHandler(t *testing.T, router kumi.Router) {
	a := tagMiddleware("a")
	b := tagMiddleware("b")

	var ran bool
	fn := func(w http.ResponseWriter, r *http.Request) {
		ran = true
	}

	k := kumi.New(router)
	k.Use(a, b)
	k.NotFoundHandler(fn)

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	k.ServeHTTP(w, r)

	if !ran {
		t.Fatal("handler did not run")
	} else if w.Body.String() != "abBA" {
		t.Fatalf("middleware stack did not run")
	}
}

// A constructor for middleware that writes a "tag" to the ResponseWriter
// for testing middleware ordering. Credit github.com/justinas/alice
// This variation writes the tag before and after to verify middleware flow.
func tagMiddleware(tag string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(strings.ToLower(tag)))
			h.ServeHTTP(w, r)
			w.Write([]byte(strings.ToUpper(tag)))
		})
	}
}

func TestMethodNotAllowedHandlers(t *testing.T) {
	routers := []struct {
		name   string
		router kumi.Router
		param  string
	}{
		{
			name:   "httprouter",
			router: router.NewHTTPRouter(),
			param:  ":id",
		},
		{
			name:   "httptreemux",
			router: router.NewHTTPTreeMux(),
			param:  ":id",
		},
		{
			name:   "gorilla",
			router: router.NewGorillaMuxRouter(),
			param:  "{id}",
		},
	}

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-Ran", "True")
			next.ServeHTTP(w, r)
		})
	}

	for _, r := range routers {
		k := kumi.New(r.router)

		// Set Global middleware to run
		k.Use(mw)

		k.Get("/", func(w http.ResponseWriter, r *http.Request) {})
		k.Post("/bla/bla", func(w http.ResponseWriter, r *http.Request) {})
		k.Patch("/path/"+r.param, func(w http.ResponseWriter, r *http.Request) {})
		k.Delete("/path/"+r.param, func(w http.ResponseWriter, r *http.Request) {})

		// Set MethodNotAllowedHandler
		k.MethodNotAllowedHandler(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Method-Not-Allowed-Handler", "True")
			w.WriteHeader(http.StatusMethodNotAllowed)
		})

		req, _ := http.NewRequest("GET", "/path/10", nil)
		w := httptest.NewRecorder()
		k.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("(%s): unexpected status code: %d", r.name, w.Code)
		} else if w.Header().Get("X-Method-Not-Allowed-Handler") != "True" { // Ensure NFH ran
			t.Fatalf("(%s): expected X-Method-Not-Allowed-Handler header", r.name)
		} else if w.Header().Get("Allow") == "" {
			t.Fatalf("(%s): expected Allow header", r.name)
		}

		a := strings.Split(w.Header().Get("Allow"), ", ")
		sort.Strings(a)
		if !reflect.DeepEqual(a, []string{"DELETE", "PATCH"}) {
			t.Fatalf("(%s):unexpected methods: %#v", r.name, a)
		}

		// Ensure global middleware ran.
		if w.Header().Get("X-Middleware-Ran") != "True" {
			t.Fatalf("(%s): expected X-Middleware-Ran header", r.name)
		}
	}
}

func TestHasRoute(t *testing.T) {
	routers := []struct {
		name   string
		router kumi.Router
		param  string
	}{
		{
			name:   "httprouter",
			router: router.NewHTTPRouter(),
			param:  ":id",
		},
		{
			name:   "httptreemux",
			router: router.NewHTTPTreeMux(),
			param:  ":id",
		},
		{
			name:   "gorilla",
			router: router.NewGorillaMuxRouter(),
			param:  "{id}",
		},
	}

	for _, r := range routers {
		k := kumi.New(r.router)

		k.Get("/", func(w http.ResponseWriter, r *http.Request) {})
		k.Post("/bla/bla", func(w http.ResponseWriter, r *http.Request) {})
		k.Patch("/path/"+r.param, func(w http.ResponseWriter, r *http.Request) {})
		k.Delete("/path/"+r.param, func(w http.ResponseWriter, r *http.Request) {})
		k.Options("/path/"+r.param, func(w http.ResponseWriter, r *http.Request) {})

		for _, method := range kumi.HTTPMethods {
			switch method {
			case "PATCH", "DELETE", "OPTIONS":
				if !k.HasRoute(method, "/path/10") {
					t.Errorf("(%s) expected %s to have route", r.name, method)
				}
			default:
				if k.HasRoute(method, "/path/10") {
					t.Errorf("(%s) expected %s to not have route", r.name, method)
				}
			}
		}

		if !k.HasRoute("GET", "/") {
			t.Fatalf("(%s) expected route to be found", r.name)
		} else if !k.HasRoute("POST", "/bla/bla") {
			t.Fatalf("(%s) expected route to be found", r.name)
		}
	}
}
