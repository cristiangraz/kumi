package kumi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/zenazn/goji/web/mutil"
)

type (
	fakeCacher struct {
		Cacher
	}

	alwaysFoundCacher struct {
		Cacher
	}

	dummyRouter struct {
		engine *Engine
		routes map[string]map[string][]HandlerFunc
	}

	// response responds to cache Check
	cacheResponse struct {
		found   bool
		status  int
		headers map[string]string
		body    io.Reader
	}

	multiResponseWriter struct {
		http.ResponseWriter
		b *bytes.Buffer
	}
)

func (mrw multiResponseWriter) Write(p []byte) (int, error) {
	return mrw.b.Write(p)
}

func TestContext(t *testing.T) {
	rw, expected := httptest.NewRecorder(), httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)

	e := New(&dummyRouter{})
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

func TestDefer(t *testing.T) {
	rec, expected := httptest.NewRecorder(), httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	e := New(&dummyRouter{})
	str := ""
	var deferrer = func(name string) HandlerFunc {
		return func(ctx *Context) {
			ctx.Defer(func() {
				str += name
			})

			ctx.Write([]byte(name))

			ctx.Next()
		}
	}

	c := e.NewContext(rec, r, mw1, deferrer("defer1"), deferrer("defer2"), mw1)
	c.Next()

	expected.Write([]byte("mw1defer1defer2mw1"))
	if !reflect.DeepEqual(expected.Body, rec.Body) {
		t.Errorf("TestDefer error: Expected %q, given %q", expected.Body, rec.Body)
	}

	if str != "" {
		t.Errorf("TestDefer: Expected defer not to run until context is returned.")
	}

	e.ReturnContext(c)

	// Defer should run in LIFO Order
	if str != "defer2defer1" {
		t.Error("TestDefer: Expected defers to run in proper order")
	}
}

func TestOnBeforeWrite(t *testing.T) {
	rec := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	e := New(&dummyRouter{})
	var conditionalWriter = func(c *Context) {
		c.BeforeWrite(func() {
			c.ResponseWriter = mutil.WrapWriter(c.ResponseWriter)
		})

		c.Next()
	}

	var mwstatus = func(c *Context) {
		c.WriteHeader(http.StatusExpectationFailed)
	}

	c := e.NewContext(rec, r, conditionalWriter, mwstatus)
	if _, ok := c.ResponseWriter.(mutil.WriterProxy); ok {
		t.Error("TestOnBeforeWrite: Expected initial response writer would not be mutil.WriterProxy")
	}

	c.Next()

	mw, ok := c.ResponseWriter.(mutil.WriterProxy)
	if !ok {
		t.Errorf("TestOnBeforeWrite: Expected before write to run")
	}

	if mw.Status() != http.StatusExpectationFailed {
		t.Errorf("TestOnBeforeWrite: Expected to capture status of %d, given %d", http.StatusExpectationFailed, mw.Status())
	}

	e.ReturnContext(c)
}

// func TestContextEventsCacheHit(t *testing.T) {
// 	finishRequest := false
// 	rec, expected := httptest.NewRecorder(), httptest.NewRecorder()
// 	r, _ := http.NewRequest("GET", "/", nil)
//
// 	wrap := func(event string) HandlerFunc {
// 		return func(ctx *Context) {
// 			// EventFinishRequest happens after response has been sent.
// 			// Nothing will be written, so need a way to verify it ran.
// 			if event == EventFinishRequest {
// 				finishRequest = true
// 			}
//
// 			ctx.Writer().Write([]byte(event))
// 		}
// 	}
//
// 	e := New(&dummyRouter{})
// 	e.SetCacher(&alwaysFoundCacher{})
// 	e.SetCompressor(&fakeCompressor{})
//
// 	c := e.NewContext(rec, r, mw1)
//
// 	for _, event := range []string{EventRequestStart, EventFinishRequest, EventResponse} {
// 		if err := e.AddListener(event, wrap(event)); err != nil {
// 			t.Errorf("TestContextEventsCacheHit: Error adding event listener. Error: %s", err)
// 		}
// 	}
//
// 	c.Start(e)
// 	e.ReturnContext(c)
//
// 	// On Cache HITs:
// 	// (E) Request starts, cache is checked, (E) before send response, send response, (E) finish request
// 	expected.Write([]byte(fmt.Sprint(EventRequestStart, "cache HIT", EventResponse, "mw1")))
// 	if !reflect.DeepEqual(expected.Body, rec.Body) {
// 		t.Errorf("TestContextEventsCacheHit: Expected %s, given %s", expected.Body, rec.Body)
// 	}
//
// 	if rec.Header().Get("X-Compressor-Fake") != "Running" {
// 		t.Error("TestContextEventsCacheHit: Expected compressor to run")
// 	}
//
// 	// Finish request runs after write has sent. Track that it executed here.
// 	if !finishRequest {
// 		t.Errorf("TestContextEventsCacheHit: Expected %s to run", EventFinishRequest)
// 	}
// }

func (c *alwaysFoundCacher) Check(ctx *Context) CacheResponse {
	ctx.Write([]byte("cache HIT"))

	return cacheResponse{
		found:  true,
		status: 200,
		body:   bytes.NewBufferString("t"),
	}
}

func (c *fakeCacher) Store(ctx *Context, ttl int) {
	// Write to the context to test io.Writer implementation.
	fmt.Fprint(ctx, "store")
}

// Handle ...
func (router *dummyRouter) Handle(method string, path string, h ...HandlerFunc) {
	if router.routes == nil {
		router.routes = make(map[string]map[string][]HandlerFunc, 1)
	}

	if _, ok := router.routes[method]; !ok {
		router.routes[method] = make(map[string][]HandlerFunc, 1)
	}

	router.routes[method][path] = h
}

func (router *dummyRouter) Lookup(method string, path string) ([]HandlerFunc, bool) {
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

func (router *dummyRouter) SetEngine(e *Engine) {
	router.engine = e
}

func (router *dummyRouter) Engine() *Engine {
	return router.engine
}

func (router *dummyRouter) ServeHTTP(rw http.ResponseWriter, r *http.Request) {}

// Found returns true if the entry was found in the cache.
func (cr cacheResponse) Found() bool {
	return cr.found
}

// Status returns the status code for the entry.
// If none is store in the cache, http.StatusOK is reeturned.
func (cr cacheResponse) Status() int {
	if cr.status == 0 {
		return http.StatusOK
	}

	return cr.status
}

// Headers return the headers stored in cache for the response.
func (cr cacheResponse) Headers() map[string]string {
	return cr.headers
}

// Body returns an io.Reader of the body.
func (cr cacheResponse) Body() io.Reader {
	return cr.body
}
