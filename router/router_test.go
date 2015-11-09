package router

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cristiangraz/kumi"
)

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
		e := kumi.New(r.router)
		r.router.SetEngine(e)
		if !reflect.DeepEqual(e, r.router.Engine()) {
			t.Errorf("TestRouters (%s): Expected Engine() would return the proper engine after SetEngine call", r.name)
		}

		funcs := map[string]func(string, ...kumi.Handler){
			"GET":     e.Get,
			"POST":    e.Post,
			"PUT":     e.Put,
			"PATCH":   e.Patch,
			"HEAD":    e.Head,
			"OPTIONS": e.Options,
			"DELETE":  e.Delete,
		}

		for m, fn := range funcs {
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
			r.router.ServeHTTP(rw, req)

			if !routed {
				t.Errorf("TestRouters (%s): Routing failed for %s request", r.name, m)
			}

			expected := httptest.NewRecorder()
			expected.Write([]byte("mw1mw2routehandler"))
			if !reflect.DeepEqual(expected.Body, rw.Body) {
				t.Errorf("TestRouters (%s): Expected body to equal %s, given %s", r.name, expected.Body, rw.Body)
			}
		}
	}
}
