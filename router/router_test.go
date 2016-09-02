package router_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
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
			if !reflect.DeepEqual(kumi.Context(r).Params, rt.params) {
				t.Fatalf("unexpected params: %v", kumi.Context(r).Params)
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

//
// func TestNotFoundHandlers(t *testing.T) {
// 	routers := []struct {
// 		name   string
// 		router kumi.Router
// 	}{
// 		{
// 			name:   "httprouter",
// 			router: NewHTTPRouter(),
// 		},
// 		{
// 			name:   "httptreemux",
// 			router: NewHTTPTreeMux(),
// 		},
// 		{
// 			name:   "gorilla",
// 			router: NewGorillaMuxRouter(),
// 		},
// 	}
//
// 	nfh := func(c *kumi.Context) {
// 		c.Header().Set("X-Not-Found-Handler", "True")
// 		c.WriteHeader(http.StatusNotFound)
// 	}
//
// 	mw := func(c *kumi.Context) {
// 		c.Header().Set("X-Middleware-Ran", "True")
// 		c.Next()
// 	}
//
// 	for _, r := range routers {
// 		k := kumi.New(r.router)
//
// 		// Set Global middleware to run
// 		k.Use(mw)
//
// 		for _, inheritMiddleware := range []bool{true, false} {
// 			// Set NotFoundHandler
// 			k.NotFoundHandler(inheritMiddleware, nfh)
//
// 			rec := httptest.NewRecorder()
// 			req, _ := http.NewRequest("GET", "/not-found-path", nil)
//
// 			c := k.NewContext(rec, req)
// 			k.ServeHTTP(c, c.Request)
// 			k.ReturnContext(c)
//
// 			if rec.Code != http.StatusNotFound {
// 				t.Errorf("TestNotFoundHandlers (%s): Expected not found handler to return 404, given %d", r.name, rec.Code)
// 			}
//
// 			// Ensure NFH ran
// 			if rec.Header().Get("X-Not-Found-Handler") != "True" {
// 				t.Errorf("TestNotFoundHandlers (%s): Expected X-Not-Found-Handler header", r.name)
// 			}
//
// 			// Ensure global middleware ran on NFH when inheritMiddleware = true
// 			expectedMw := ""
// 			if inheritMiddleware {
// 				expectedMw = "True"
// 			}
// 			if rec.Header().Get("X-Middleware-Ran") != expectedMw {
// 				t.Errorf("TestNotFoundHandlers (%s): Expected X-Middleware-Ran header", r.name)
// 			}
// 		}
// 	}
// }
//
// func TestMethodNotAllowedHandlers(t *testing.T) {
// 	routers := []struct {
// 		name   string
// 		router kumi.Router
// 	}{
// 		{
// 			name:   "httprouter",
// 			router: NewHTTPRouter(),
// 		},
// 		{
// 			name:   "httptreemux",
// 			router: NewHTTPTreeMux(),
// 		},
// 		{
// 			name:   "gorilla",
// 			router: NewGorillaMuxRouter(),
// 		},
// 	}
//
// 	mnah := func(c *kumi.Context) {
// 		c.Header().Set("X-Method-Not-Allowed-Handler", "True")
// 		c.WriteHeader(http.StatusMethodNotAllowed)
// 	}
//
// 	mw := func(c *kumi.Context) {
// 		c.Header().Set("X-Middleware-Ran", "True")
// 		c.Next()
// 	}
//
// 	expected := []string{"PATCH", "DELETE"}
// 	for _, r := range routers {
// 		k := kumi.New(r.router)
//
// 		// Set Global middleware to run
// 		k.Use(mw)
//
// 		k.Get("/")
// 		k.Post("/bla/bla")
// 		k.Patch("/path")
// 		k.Delete("/path")
//
// 		for _, inheritMiddleware := range []bool{true, false} {
// 			// Set MethodNotAllowedHandler
// 			k.MethodNotAllowedHandler(inheritMiddleware, mnah)
//
// 			rec := httptest.NewRecorder()
// 			req, _ := http.NewRequest("GET", "/path", nil)
//
// 			c := k.NewContext(rec, req)
// 			k.ServeHTTP(c, c.Request)
// 			k.ReturnContext(c)
//
// 			if rec.Code != http.StatusMethodNotAllowed {
// 				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected method not allowed handler to return 405, given %d", r.name, rec.Code)
// 			}
//
// 			// Ensure NFH ran
// 			if rec.Header().Get("X-Method-Not-Allowed-Handler") != "True" {
// 				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected X-Method-Not-Allowed-Handler header", r.name)
// 			}
//
// 			if rec.Header().Get("Allow") == "" {
// 				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected Allow header. None given", r.name)
// 			}
//
// 			given := strings.Split(rec.Header().Get("Allow"), ", ")
// 			if len(given) != 2 {
// 				t.Errorf("TestmMethodNotAllowedHandlers (%s): Expected allow header with 2 methods. %d given", r.name, len(given))
// 			}
//
// 			if !((given[0] == "PATCH" && given[1] == "DELETE") || (given[0] == "DELETE" && given[1] == "PATCH")) {
// 				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected allow header to contain %q, given %q", r.name, expected, given)
// 			}
//
// 			// Ensure global middleware ran on NFH when inheritMiddleware = true
// 			expectedMw := ""
// 			if inheritMiddleware {
// 				expectedMw = "True"
// 			}
// 			if rec.Header().Get("X-Middleware-Ran") != expectedMw {
// 				t.Errorf("TestMethodNotAllowedHandlers (%s): Expected X-Middleware-Ran header", r.name)
// 			}
// 		}
// 	}
// }
