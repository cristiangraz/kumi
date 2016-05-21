package kumi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/cristiangraz/kumi/cache"
)

type (
	testRouter struct {
		engine           *Engine
		routes           map[string]map[string][]HandlerFunc
		notFound         []HandlerFunc
		methodNotAllowed []HandlerFunc
	}
)

func TestContext(t *testing.T) {
	rw, expected := httptest.NewRecorder(), httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	e := New(&testRouter{})
	c := e.NewContext(rw, r, mw2, mw1, h2)

	if c.Status() != http.StatusOK {
		t.Errorf("TestContext: Expected zero-value status to return %d, given %d", http.StatusOK, c.Status())
	}

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

func BenchmarkContext(b *testing.B) {
	rw := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	b.ReportAllocs()
	b.ResetTimer()

	rtr := &testRouter{}
	e := New(rtr)
	for i := 0; i < b.N; i++ {
		c := e.NewContext(rw, r, mw1)

		c.Next()
		e.ReturnContext(c)
	}
}

func TestContext_ContentTypeReturnedWhenNoneSpecified(t *testing.T) {
	rec := httptest.NewRecorder()

	k := New(&testRouter{})
	r, _ := http.NewRequest("GET", "/", nil)
	c := k.NewContext(rec, r, func(c *Context) {
		c.Write([]byte("hello"))
	})
	c.Next()
	k.ReturnContext(c)

	if rec.Body.String() != "hello" {
		t.Errorf("TestContentTypeReturnedWhenNoneSpecified: want=%s, actual=%s", "hello", rec.Body.String())
	}

	if actual := rec.Header().Get("Content-Type"); actual != "text/plain" {
		t.Fatalf("unexpected content-type, want=%s, actual=%s", "text/plain", actual)
	}
}

func TestContext_HeadRequestsDontReturnBody(t *testing.T) {
	rec := httptest.NewRecorder()

	k := New(&testRouter{})
	r, _ := http.NewRequest("HEAD", "/", nil)
	c := k.NewContext(rec, r, func(c *Context) {
		c.Write([]byte("hello"))
	})
	c.Next()

	if _, ok := c.ResponseWriter.(*BodylessResponseWriter); !ok {
		t.Error("TestHeadRequestsDontReturnBody: Expected HEAD request to use BodylessResponseWriter")
	}

	k.ReturnContext(c)

	if len(rec.Body.Bytes()) > 0 {
		t.Error("TestHeadRequestsDontReturnBody: Didn't expect any bytes to be written")
	}

	if actual := rec.Header().Get("Content-Type"); actual != "" {
		t.Fatalf("expected content-type to be empty: %v", actual)
	}
}

func TestContext_NoContentResponsesDontReturnBody(t *testing.T) {
	rec := httptest.NewRecorder()

	k := New(&testRouter{})
	r, _ := http.NewRequest("GET", "/", nil)
	c := k.NewContext(rec, r, func(c *Context) {
		c.WriteHeader(http.StatusNoContent)
		c.Write([]byte("hello"))
	})
	c.Next()

	if _, ok := c.ResponseWriter.(*BodylessResponseWriter); !ok {
		t.Error("TestNoContentResponsesDontReturnBody: Expected 204 request to use BodylessResponseWriter")
	}

	k.ReturnContext(c)

	if len(rec.Body.Bytes()) > 0 {
		t.Error("TestNoContentResponsesDontReturnBody: Didn't expect any bytes to be written")
	}

	if actual := rec.Header().Get("Content-Type"); actual != "" {
		t.Fatalf("expected content-type to be empty: %v", actual)
	}
}

// Even if the content-type is set, it should not be returned when
// no body is written.
func TestContext_ContentTypeRemovedWhenNoBody(t *testing.T) {
	rec := httptest.NewRecorder()

	k := New(&testRouter{})
	r, _ := http.NewRequest("GET", "/", nil)
	c := k.NewContext(rec, r, func(c *Context) {
		c.Header().Set("Content-Type", "application/json")
		c.WriteHeader(http.StatusNoContent)
		c.Write([]byte("hello"))
	})
	c.Next()
	k.ReturnContext(c)

	if actual := rec.Header().Get("Content-Type"); actual != "" {
		t.Fatalf("expected content-type to be empty: %v", actual)
	}
}

func TestContext_CacheControl(t *testing.T) {
	rec := httptest.NewRecorder()

	k := New(&testRouter{})
	r, _ := http.NewRequest("GET", "/", nil)
	c := k.NewContext(rec, r, func(c *Context) {
		c.Write([]byte("hello"))
	})
	c.Next()
	k.ReturnContext(c)

	// When no cache-control header is sent, SensibleDefaults should be used
	cc := cache.New()
	if want := cc.NoCache().SetPrivate().String(); rec.Header().Get("Cache-Control") != want {
		t.Errorf("TestContext_CacheControl: want=%s, actual=%s", want, rec.Header().Get("Cache-Control"))
	}
}

func TestContext_ManualCacheControl(t *testing.T) {
	rec := httptest.NewRecorder()

	want := "public, s-maxage=30"

	k := New(&testRouter{})
	r, _ := http.NewRequest("GET", "/", nil)
	c := k.NewContext(rec, r, func(c *Context) {
		c.Header().Set("Cache-Control", want)
		c.Write([]byte("hello"))
	})
	c.Next()
	k.ReturnContext(c)

	// When cache-control header is sent, it shouldn't be changed or overridden
	if rec.Header().Get("Cache-Control") != want {
		t.Errorf("TestContext_CacheControl: want=%s, actual=%s", want, rec.Header().Get("Cache-Control"))
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
		methods := router.getMethods(r)
		if len(methods) > 0 {
			// 405 Not Allowed
			if len(router.methodNotAllowed) == 0 {
				w.Header().Set("Allow", strings.Join(methods, ", "))
				w.WriteHeader(http.StatusMethodNotAllowed)
			} else {
				e := router.Engine()
				c := e.NewContext(w, r, router.methodNotAllowed...)
				defer e.ReturnContext(c)

				c.Header().Set("Allow", strings.Join(methods, ", "))

				c.Next()
			}

			return
		}

		// 404
		if len(router.notFound) == 0 {
			w.WriteHeader(http.StatusNotFound)
		} else {
			e := router.Engine()
			c := e.NewContext(w, r, router.notFound...)
			defer e.ReturnContext(c)

			c.Next()
		}
	}

	e := router.Engine()
	c := e.NewContext(w, r, h...)
	defer e.ReturnContext(c)

	c.Next()
}

func (router *testRouter) NotFoundHandler(fn ...HandlerFunc) {
	router.notFound = fn
}

func (router *testRouter) MethodNotAllowedHandler(fn ...HandlerFunc) {
	router.methodNotAllowed = fn
}

// HasRoute ...
func (router *testRouter) HasRoute(method string, pattern string) bool {
	if _, found := router.Lookup(method, pattern); found {
		return true
	}

	return false
}

// getMethods ...
func (router *testRouter) getMethods(r *http.Request) (methods []string) {
	for _, m := range HTTPMethods {
		if _, found := router.Lookup(m, r.URL.Path); found {
			methods = append(methods, m)
		}
	}

	return methods
}
