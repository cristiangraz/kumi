package router

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cristiangraz/kumi"
)

func TestHasRoute(t *testing.T) {
	routers := []struct {
		name, route, req string
		router           kumi.Router
		params           kumi.Params
	}{
		{
			name:   "httprouter",
			route:  "/users/:name/:id",
			req:    "/users/httprouter/10",
			router: NewHTTPRouter(),
			params: kumi.Params{"name": "httprouter", "id": "10"},
		},
		{
			name:   "httptreemux",
			route:  "/users/:name/:msg",
			req:    "/users/httptreemux/hello",
			router: NewHTTPTreeMux(),
			params: kumi.Params{"name": "httptreemux", "msg": "hello"},
		},
		{
			name:   "gorilla",
			route:  "/users/{name}",
			req:    "/users/gorilla",
			router: NewGorillaMuxRouter(),
			params: kumi.Params{"name": "gorilla"},
		},
	}

	for _, r := range routers {
		for _, m := range kumi.HTTPMethods {
			if found := r.router.HasRoute(m, r.route); found {
				t.Errorf("TestHasRoute (%s): Expected %s route to return false before route was defined", r.name, m)
			}

			r.router.Handle(m, r.route)

			if found := r.router.HasRoute(m, r.route); !found {
				t.Errorf("TestHasRoute (%s): Expected %s route to be found after it is registered", r.name, m)
			}
		}
	}
}

func TestRouters(t *testing.T) {
	routers := []struct {
		name, route, req string
		router           kumi.Router
		params           kumi.Params
	}{
		{
			name:   "httprouter",
			route:  "/users/:name/:id",
			req:    "/users/httprouter/10",
			router: NewHTTPRouter(),
			params: kumi.Params{"name": "httprouter", "id": "10"},
		},
		{
			name:   "httptreemux",
			route:  "/users/:name/:msg",
			req:    "/users/httptreemux/hello",
			router: NewHTTPTreeMux(),
			params: kumi.Params{"name": "httptreemux", "msg": "hello"},
		},
		{
			name:   "gorilla",
			route:  "/users/{name}",
			req:    "/users/gorilla",
			router: NewGorillaMuxRouter(),
			params: kumi.Params{"name": "gorilla"},
		},
	}

	mw1 := func(c *kumi.Context) {
		c.Write([]byte("mw1"))
		c.Next()
	}

	mw2 := func(c *kumi.Context) {
		c.Write([]byte("mw2"))
		c.Next()
	}

	for _, r := range routers {
		// Kumi sets OPTIONS route for each route. So test first.
		// Kumi sets HEAD route with each GET route, so set before GET.
		methods := []string{"OPTIONS", "HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"}
		for _, m := range methods {
			k := kumi.New(r.router)
			r.router.SetEngine(k)
			if !reflect.DeepEqual(k, r.router.Engine()) {
				t.Errorf("TestRouters (%s): Expected Engine() would return the proper engine after SetEngine call", r.name)
			}

			var fn func(string, ...kumi.Handler)
			switch m {
			case "OPTIONS":
				fn = k.Options
			case "HEAD":
				fn = k.Head
			case "GET":
				fn = k.Get
			case "POST":
				fn = k.Post
			case "PUT":
				fn = k.Put
			case "PATCH":
				fn = k.Patch
			case "DELETE":
				fn = k.Delete
			}

			routed := false
			h := func(c *kumi.Context) {
				routed = true
				c.Write([]byte("routehandler"))
				if !reflect.DeepEqual(r.params, c.Params) {
					t.Errorf("TestRouters (%s): Expected params of %v, given %v", r.name, r.params, c.Params)
				}
			}

			rw := httptest.NewRecorder()
			fn(r.route, mw1, mw2, h)
			req, _ := http.NewRequest(m, r.req, nil)
			k.ServeHTTP(rw, req)

			if !routed {
				t.Errorf("TestRouters (%s): Routing failed for %s request", r.name, m)
			}

			expected := httptest.NewRecorder()
			if m != "HEAD" {
				expected.Write([]byte("mw1mw2routehandler"))
			}

			if !reflect.DeepEqual(expected.Body, rw.Body) {
				t.Errorf("TestRouters (%s): Expected body to equal %s, given %s", r.name, expected.Body, rw.Body)
			}
		}
	}
}
