package kumi

import (
	"net/http"
	"sync"

	"github.com/cristiangraz/kumi/cache"
	"golang.org/x/net/context"
)

// Context holds contextual data for each request and
// implements the http.ResponseWriter interface.
type Context struct {
	http.ResponseWriter
	context.Context
	Request      *http.Request
	CacheHeaders *cache.Headers
	Query        Query
	Params       Params

	engine      *Engine
	writeHeader sync.Once
	handlers    []HandlerFunc
	status      int
}

type key int

const (
	panicKey key = iota
)

// Status returns the http status code. If none has been set,
// http.StatusOK (200) will be returned.
func (c *Context) Status() int {
	if c.status == 0 {
		return http.StatusOK
	}

	return c.status
}

// WriteHeader prepares the response once.
func (c *Context) WriteHeader(s int) {
	c.writeHeader.Do(func() {
		hasBody := true
		if s == http.StatusNoContent {
			hasBody = false
			c.ResponseWriter = &BodylessResponseWriter{c.ResponseWriter}
		} else if _, ok := c.ResponseWriter.(*BodylessResponseWriter); ok {
			hasBody = false
		}

		c.status = s
		c.CacheHeaders.SensibleDefaults(c.Header(), c.Status())

		ct := c.Header().Get("Content-Type")
		if hasBody && ct == "" {
			c.Header().Set("Content-Type", "text/plain")
		} else if !hasBody {
			c.Header().Del("Content-Type")
		}

		c.ResponseWriter.WriteHeader(s)
	})
}

// Writes the response.
func (c *Context) Write(p []byte) (int, error) {
	c.WriteHeader(http.StatusOK)
	return c.ResponseWriter.Write(p)
}

// Next runs the next handler in the chain. It should be called from all of your
// middleware except the last http handler. If you don't call it from your handler,
// no additional handlers will be called.
func (c *Context) Next() {
	if len(c.handlers) == 0 {
		return
	}

	h := c.handlers[0]
	c.handlers = c.handlers[1:]

	h(c)
}

// ServeHTTP makes context compatible with the http.Handler interface.
func (c *Context) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	c.Next()
}

// newContext creates a new context for the sync pool.
func newContext(rw http.ResponseWriter, r *http.Request, handlers ...HandlerFunc) *Context {
	var once sync.Once

	return &Context{
		Context:        context.Background(),
		Request:        r,
		ResponseWriter: rw,
		CacheHeaders:   cache.NewHeaders(),
		handlers:       handlers,
		Query:          Query{r},

		writeHeader: once,
		status:      0,
	}
}

// reset resets the context.
func (c *Context) reset(rw http.ResponseWriter, r *http.Request, handlers ...HandlerFunc) {
	var params Params
	var once sync.Once

	c.Context = context.Background()
	c.Request = r
	c.ResponseWriter = rw
	c.CacheHeaders = cache.NewHeaders()
	c.handlers = handlers
	c.Query = Query{r}
	c.Params = params

	c.writeHeader = once
	c.status = 0
}

// NewContextWithException adds an exception to the context.
func NewContextWithException(c *Context, exception interface{}) {
	c.Context = context.WithValue(c.Context, panicKey, exception)
}

// Exception gets the "v" in panic(v). The panic details.
// Only PanicHandler will receive a context you can use this with.
func Exception(c *Context) interface{} {
	return c.Context.Value(panicKey)
}
