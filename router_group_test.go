package kumi

import (
	"fmt"
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
		router := &dummyRouter{}
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
				t.Errorf("TestRouterGroup: Expected registered route to be found")
			}

			for i, e := range expected {
				if !funcEqual(h[i], e) {
					t.Errorf("TestRouterGroup (%s): Expected handlers to match", m)
				}
			}
		}
	}
}

func TestRouterGroupAll(t *testing.T) {
	router := &dummyRouter{}
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

func TestInvalidHandlers(t *testing.T) {
	wrong := func() {}
	k := New(&dummyRouter{})
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

func funcEqual(a, b Handler) bool {
	av := reflect.ValueOf(&a).Elem()
	bv := reflect.ValueOf(&b).Elem()

	return av.InterfaceData() == bv.InterfaceData()
}
