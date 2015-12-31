package kumi

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type (
	testRouter struct {
		engine   *Engine
		routes   map[string]map[string][]HandlerFunc
		notFound []HandlerFunc
	}
)

func TestContext(t *testing.T) {
	rw, expected := httptest.NewRecorder(), httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	e := New(&testRouter{})
	c := e.NewContext(rw, r, mw2, mw1, h2)

	c.Next()
	e.ReturnContext(c)

	expected.Write([]byte("mw2mw1h2"))
	if !reflect.DeepEqual(expected.Body, rw.Body) {
		t.Errorf("TestContext: Expected %s, given %s", expected.Body, rw.Body)
	}

	// Run again from engine Context
	// Use haltMw that doesn't call c.Next(). h1 shouldnt' run.
	rw, expected = httptest.NewRecorder(), httptest.NewRecorder()
	c = e.NewContext(rw, r, mw1, haltMw, h1)

	c.Next()
	e.ReturnContext(c)

	expected = httptest.NewRecorder()
	expected.Write([]byte("mw1stopping"))
	if !reflect.DeepEqual(expected.Body, rw.Body) {
		t.Errorf("TestContext: Halt mw. Expected %s, given %s", expected.Body, rw.Body)
	}

	// Test Exception
	err := errors.New("Panic exception")
	c = e.NewContext(rw, r)
	defer e.ReturnContext(c)
	NewContextWithException(c, err)

	if !reflect.DeepEqual(err, Exception(c)) {
		t.Error("TestContext: Expected exceptions to be equal")
	}
}

// Handle ...
func (router *testRouter) Handle(method string, path string, h ...HandlerFunc) {
	if router.routes == nil {
		router.routes = make(map[string]map[string][]HandlerFunc, 1)
	}

	if _, ok := router.routes[method]; !ok {
		router.routes[method] = make(map[string][]HandlerFunc, 1)
	}

	router.routes[method][path] = h
}

func (router *testRouter) Lookup(method string, path string) ([]HandlerFunc, bool) {
	if router.routes == nil {
		return nil, false
	}

	if _, ok := router.routes[method]; !ok {
		return nil, false
	}

	if h, ok := router.routes[method][path]; ok {
		return h, true
	}

	return nil, false
}

func (router *testRouter) SetEngine(e *Engine) {
	router.engine = e
}

func (router *testRouter) Engine() *Engine {
	return router.engine
}

func (router *testRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h, found := router.Lookup(r.Method, r.URL.Path)
	if !found {
		if len(router.notFound) == 0 {
			log.Println("Route not found")
			w.WriteHeader(http.StatusNotFound)
		} else {
			e := router.Engine()
			c := e.NewContext(w, r, router.notFound...)
			defer e.ReturnContext(c)

			c.Next()
		}
		return
	}

	e := router.Engine()
	c := e.NewContext(w, r, h...)
	defer e.ReturnContext(c)

	c.Next()
}

func (router *testRouter) NotFoundHandler(fn ...HandlerFunc) {
	router.notFound = fn
}

// HasRoute ...
func (router *testRouter) HasRoute(method string, pattern string) bool {
	if _, found := router.Lookup(method, pattern); found {
		return true
	}

	return false
}
