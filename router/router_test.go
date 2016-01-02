package router

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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

func TestNotFoundHandlers(t *testing.T) {
	routers := []struct {
		name   string
		router kumi.Router
	}{
		{
			name:   "httprouter",
			router: NewHTTPRouter(),
		},
		{
			name:   "httptreemux",
			router: NewHTTPTreeMux(),
		},
		{
			name:   "gorilla",
			router: NewGorillaMuxRouter(),
		},
	}

	nfh := func(c *kumi.Context) {
		c.Header().Set("X-Not-Found-Handler", "True")
		c.WriteHeader(http.StatusNotFound)
	}

	mw := func(c *kumi.Context) {
		c.Header().Set("X-Middleware-Ran", "True")
		c.Next()
	}

	for _, r := range routers {
		k := kumi.New(r.router)

		// Set Global middleware to run
		k.Use(mw)

		for _, inheritMiddleware := range []bool{true, false} {
			// Set NotFoundHandler
			k.NotFoundHandler(inheritMiddleware, nfh)

			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/not-found-path", nil)

			c := k.NewContext(rec, req)
			k.ServeHTTP(c, c.Request)
			k.ReturnContext(c)

			if rec.Code != http.StatusNotFound {
				t.Errorf("TestNotFoundHandlers (%s): Expected not found handler to return 404, given %d", r.name, rec.Code)
			}

			// Ensure NFH ran
			if rec.Header().Get("X-Not-Found-Handler") != "True" {
				t.Errorf("TestNotFoundHandlers (%s): Expected X-Not-Found-Handler header", r.name)
			}

			// Ensure global middleware ran on NFH when inheritMiddleware = true
			expectedMw := ""
			if inheritMiddleware {
				expectedMw = "True"
			}
			if rec.Header().Get("X-Middleware-Ran") != expectedMw {
				t.Errorf("TestNotFoundHandlers (%s): Expected X-Middleware-Ran header", r.name)
			}
		}
	}
}

func TestMethodNotAllowedHandlers(t *testing.T) {
	routers := []struct {
		name   string
		router kumi.Router
	}{
		{
			name:   "httprouter",
			router: NewHTTPRouter(),
		},
		{
			name:   "httptreemux",
			router: NewHTTPTreeMux(),
		},
		{
			name:   "gorilla",
			router: NewGorillaMuxRouter(),
		},
	}

	mnah := func(c *kumi.Context) {
		c.Header().Set("X-Method-Not-Allowed-Handler", "True")
		c.WriteHeader(http.StatusMethodNotAllowed)
	}

	mw := func(c *kumi.Context) {
		c.Header().Set("X-Middleware-Ran", "True")
		c.Next()
	}

	expected := []string{"PATCH", "DELETE"}
	for _, r := range routers {
		k := kumi.New(r.router)

		// Set Global middleware to run
		k.Use(mw)

		k.Get("/")
		k.Post("/bla/bla")
		k.Patch("/path")
		k.Delete("/path")

		for _, inheritMiddleware := range []bool{true, false} {
			// Set MethodNotAllowedHandler
			k.MethodNotAllowedHandler(inheritMiddleware, mnah)

			rec := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/path", nil)

			c := k.NewContext(rec, req)
			k.ServeHTTP(c, c.Request)
			k.ReturnContext(c)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected method not allowed handler to return 405, given %d", r.name, rec.Code)
			}

			// Ensure NFH ran
			if rec.Header().Get("X-Method-Not-Allowed-Handler") != "True" {
				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected X-Method-Not-Allowed-Handler header", r.name)
			}

			if rec.Header().Get("Allow") == "" {
				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected Allow header. None given", r.name)
			}

			given := strings.Split(rec.Header().Get("Allow"), ", ")
			if len(given) != 2 {
				t.Errorf("TestmMethodNotAllowedHandlers (%s): Expected allow header with 2 methods. %d given", r.name, len(given))
			}

			if !((given[0] == "PATCH" && given[1] == "DELETE") || (given[0] == "DELETE" && given[1] == "PATCH")) {
				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected allow header to contain %q, given %q", r.name, expected, given)
			}

			// Ensure global middleware ran on NFH when inheritMiddleware = true
			expectedMw := ""
			if inheritMiddleware {
				expectedMw = "True"
			}
			if rec.Header().Get("X-Middleware-Ran") != expectedMw {
				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected X-Middleware-Ran header", r.name)
			}
		}
	}
}
