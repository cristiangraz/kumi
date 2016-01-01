package kumi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	// Middleware and handler functions used in a variety of tests.
	mw1 = func(c *Context) {
		c.Write([]byte("mw1"))
		c.Next()
	}

	mw2 = func(c *Context) {
		c.Write([]byte("mw2"))
		c.Next()
	}

	haltMw = func(c *Context) {
		c.Write([]byte("stopping"))
	}

	h1 = func(c *Context) {
		c.Write([]byte("h1"))
	}

	h2 = func(c *Context) {
		c.Write([]byte("h2"))
	}
)

func TestRouterGroup(t *testing.T) {
	suite := []struct {
		global, handlers, groupHandlers []Handler
		expected                        []Handler
		groupPrefix                     string
	}{
		{
			handlers: []Handler{mw1, mw2},
			expected: []Handler{mw1, mw2},
		},
		{
			handlers: []Handler{mw2, h1},
			expected: []Handler{mw2, h1},
		},
		{
			global:   []Handler{mw1, mw2},
			handlers: []Handler{h2},
			expected: []Handler{mw1, mw2, h2},
		},
		// Grouping
		{
			groupPrefix:   "/articles",
			groupHandlers: []Handler{mw1},
			handlers:      []Handler{h2},
			expected:      []Handler{mw1, h2},
		},
		{
			global:        []Handler{mw2},
			groupPrefix:   "/articles",
			groupHandlers: []Handler{mw1},
			handlers:      []Handler{h1},
			expected:      []Handler{mw2, mw1, h1},
		},
		{
			global:      []Handler{mw2},
			groupPrefix: "/articles",
			handlers:    []Handler{h1},
			expected:    []Handler{mw2, h1},
		},
		{
			groupPrefix: "/articles",
			handlers:    []Handler{h1},
			expected:    []Handler{h1},
		},
	}

	for _, s := range suite {
		router := &testRouter{}
		k := New(router)
		wh, err := wrapHandlers(s.expected...)
		if err != nil {
			t.Fatal("TestRouterGroup: Expected handlers were not valid for tests.")
		}
		expected := appendHandlers(wh)

		if len(s.global) > 0 {
			k.Use(s.global...)
		}

		funcs := map[string]func(string, ...Handler){
			"GET":     k.Get,
			"POST":    k.Post,
			"PUT":     k.Put,
			"PATCH":   k.Patch,
			"HEAD":    k.Head,
			"OPTIONS": k.Options,
			"DELETE":  k.Delete,
		}

		var g RouterGroup
		if s.groupPrefix != "" {
			g = k.Group(s.groupPrefix, s.groupHandlers...)
			funcs = map[string]func(string, ...Handler){
				"GET":     g.Get,
				"POST":    g.Post,
				"PUT":     g.Put,
				"PATCH":   g.Patch,
				"HEAD":    g.Head,
				"OPTIONS": g.Options,
				"DELETE":  g.Delete,
			}
		}

		for m, fn := range funcs {
			fn("/abc", s.handlers...)
			h, ok := router.Lookup(m, fmt.Sprintf("%s%s", s.groupPrefix, "/abc"))
			if !ok {
				t.Errorf("TestRouterGroup (%s): Expected registered route to be found", m)
			}

			for i, e := range expected {
				if !funcEqual(h[i], e) {
					t.Errorf("TestRouterGroup (%s): Expected handlers to match", m)
				}
			}
		}
	}
}

func TestRouterGroupGETCreatedHEAD(t *testing.T) {
	router := &testRouter{}
	k := New(router)

	handlers := []Handler{h2}
	k.Get("/hello", handlers...)

	_, ok := router.Lookup("GET", "/hello")
	if !ok {
		t.Errorf("TestRouterGroupGETCreatedHEAD: Expected GET route to be found")
	}

	_, ok = router.Lookup("HEAD", "/hello")
	if !ok {
		t.Errorf("TestRouterGroupGETCreatedHEAD: Expected HEAD route to be found")
	}
}

func TestRouterGroupAll(t *testing.T) {
	router := &testRouter{}
	k := New(router)

	handlers := []Handler{h1}
	wh, err := wrapHandlers(handlers...)
	if err != nil {
		t.Fatal("TestRouterGroupAll: Expected handlers were not valid for tests.")
	}
	expected := appendHandlers(wh)

	k.All("/all", handlers...)
	for _, m := range []string{"GET", "HEAD", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"} {
		h, ok := router.Lookup(m, "/all")
		if !ok {
			t.Errorf("TestRouterGroupAll (%s): Expected registered route to be found", m)
		}

		if !funcEqual(h[0], expected[0]) {
			t.Errorf("TestRouterGroupAll (%s): Expected handlers to match", m)
		}
	}
}

func TestGetRegistersHead(t *testing.T) {
	k := New(&testRouter{})

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("HEAD", "/foo", nil)

	k.Get("/foo", testHandler)
	k.ServeHTTP(rec, req)

	if rec.Code >= 200 && rec.Code < 300 {
		return
	}

	t.Errorf("TestGetRegistersHead: Expected HEAD route to be found when GET route was registered. Status: %d", rec.Code)
}

func TestInvalidHandlers(t *testing.T) {
	wrong := func() {}
	k := New(&testRouter{})
	suite := []struct {
		fn func()
	}{
		{
			fn: func() {
				k.Group("/", wrong)
			},
		},
		{
			fn: func() {
				k.Use(wrong)
			},
		},
		{
			fn: func() {
				k.handle("GET", "/", mw1, wrong)
			},
		},
	}

	for _, s := range suite {
		func() {
			recovered := false
			defer func() {
				if err := recover(); err != nil {
					recovered = true
				}
			}()

			s.fn()

			if !recovered {
				t.Errorf("TestInvalidHandlers: Expected invalid handler to panic")
			}
		}()
	}
}

func TestNotFoundHandler(t *testing.T) {
	for _, inheritMiddleware := range []bool{true, false} {
		r := &testRouter{}
		k := New(r)

		nfh := func(c *Context) {
			c.Header().Set("X-Not-Found-Handler", "True")
			c.WriteHeader(http.StatusNotFound)
		}

		mw := func(c *Context) {
			c.Header().Set("X-Middleware-Ran", "True")
			c.Next()
		}

		// Set Global middleware to run
		k.Use(mw)

		// Set NotFoundHandler
		k.NotFoundHandler(inheritMiddleware, nfh)

		// NotFoundHandler should include global middleware
		expectedLength := 1
		if inheritMiddleware {
			expectedLength = 2
		}
		if len(r.notFound) != expectedLength {
			t.Error("TestNotFoundHandler: Expected not found handler to have two routes")
		}

		rec := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/not-found-path", nil)

		c := k.NewContext(rec, req)
		k.ServeHTTP(c, c.Request)
		k.ReturnContext(c)

		if rec.Code != http.StatusNotFound {
			t.Errorf("Expected not found handler to return 404, given %d", rec.Code)
		}

		if rec.Header().Get("X-Not-Found-Handler") != "True" {
			t.Error("TestNotFoundHandler: Expected X-Not-Found-Handler header")
		}

		// Ensure global middleware ran on NFH when inheritMiddleware = true
		expectedMw := ""
		if inheritMiddleware {
			expectedMw = "True"
		}
		if rec.Header().Get("X-Middleware-Ran") != expectedMw {
			t.Error("TestNotFoundHandler: Expected X-Middleware-Ran header")
		}
	}
}

func funcEqual(a, b Handler) bool {
	av := reflect.ValueOf(&a).Elem()
	bv := reflect.ValueOf(&b).Elem()

	return av.InterfaceData() == bv.InterfaceData()
}
