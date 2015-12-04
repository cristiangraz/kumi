package kumi

import (
	"net/http"
	"sync"

	"github.com/cristiangraz/kumi/cache"
	"golang.org/x/net/context"
)

type (
	// Context holds contextual data for each request and
	// implements the http.ResponseWriter interface.
	Context struct {
		http.ResponseWriter
		Context      context.Context
		Request      *http.Request
		CacheHeaders *cache.Headers
		Query        Query
		Params       Params

		deferred    []func()
		beforeWrite []func()
		engine      *Engine
		writeHeader sync.Once
		handlers    []HandlerFunc
		status      int
	}

	key int
)

var (
	// DefaultContext is the starting context used for each request.
	DefaultContext = context.Background()
)

const (
	panicKey key = iota
	cacheHitKey
	cacheTTLKey
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
		c.status = s
		c.CacheHeaders.SensibleDefaults(c.Header(), c.Status())

		// Run any callbacks
		for _, fn := range c.beforeWrite {
			fn()
		}

		if c.Header().Get("Content-Type") == "" {
			c.Header().Set("Content-Type", "text/plain")
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
	select {
	case <-c.Context.Done():
		return
	default:
		if len(c.handlers) == 0 {
			return
		}

		h := c.handlers[0:1][0]
		c.handlers = c.handlers[1:]

		h(c)
	}
}

// ServeHTTP makes context compatible with the http.Handler interface.
func (c *Context) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	c.Next()
}

// Defer registers defer handlers in LIFO order.
// This works similar to Go's built in defer, but occurs when the
// context is about to be reset rather than when the function
// closes. This is useful for wrapping the response writer
// and closing the writer in the proper order.
func (c *Context) Defer(fn func()) {
	c.deferred = append([]func(){fn}, c.deferred...)
}

// BeforeWrite adds a function to be executed in FIFO order
// just before WriteHeader is called on http.ResponseWriter.
// Useful for conditional writers.
func (c *Context) BeforeWrite(fn func()) {
	c.beforeWrite = append(c.beforeWrite, fn)
}

// newContext creates a new context for the sync pool.
func newContext(rw http.ResponseWriter, r *http.Request, handlers ...HandlerFunc) *Context {
	return &Context{
		Context:        DefaultContext,
		Request:        r,
		ResponseWriter: rw,
		CacheHeaders:   cache.NewHeaders(),
		handlers:       handlers,
		Query:          Query{r},

		writeHeader: sync.Once{},
		status:      0,
		deferred:    []func(){},
		beforeWrite: []func(){},
	}
}

// reset resets the context.
func (c *Context) reset(rw http.ResponseWriter, r *http.Request, handlers ...HandlerFunc) {
	c.Context = DefaultContext
	c.Request = r
	c.ResponseWriter = rw
	c.CacheHeaders = cache.NewHeaders()
	c.handlers = handlers
	c.Query = Query{r}
	c.Params = Params{}

	c.engine = nil
	c.writeHeader = sync.Once{}
	c.status = 0
	c.deferred = []func(){}
	c.beforeWrite = []func(){}
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
