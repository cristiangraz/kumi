package kumi

import (
	"context"
	"net/http"
	"sync"
)

// RequestContext returns route params and query params for the
// current request.
type RequestContext interface {
	Params() Params
	Query() *Query
}

type key int

const (
	contextKey key = iota
	paramsKey
)

// Context retrieves the request context.
func Context(r *http.Request) RequestContext {
	return FromContext(r).(RequestContext)
}

// SetRequestContext sets a custom value in kumi's Context slot.
func SetRequestContext(r *http.Request, rc RequestContext) *http.Request {
	ctx := context.WithValue(r.Context(), contextKey, rc)
	return r.WithContext(ctx)
}

// FromContext allows you to create your own Context function that returns a
// specific RequestContext type custom to your application, while still protecting
// the context key within kumi. In that case, use FromContext instead of Context.
func FromContext(r *http.Request) interface{} {
	return r.Context().Value(contextKey)
}

// SetParams sets Params in the context for kumi to access. These will be
// moved to the RequestContext immediately after the router sets them.
// This should generally only be called from a Router.
func SetParams(r *http.Request, p Params) *http.Request {
	ctx := context.WithValue(r.Context(), paramsKey, p)
	return r.WithContext(ctx)
}

func getParams(r *http.Request) (Params, bool) {
	p, ok := r.Context().Value(paramsKey).(Params)
	return p, ok
}

type requestContext struct {
	params Params
	query  *Query
}

var _ RequestContext = &requestContext{}

// Params returns the request params.
func (r *requestContext) Params() Params {
	return r.params
}

// Query returns the query params for the request.
func (r *requestContext) Query() *Query {
	return r.query
}

var requestContextPool = &sync.Pool{
	New: func() interface{} {
		return &requestContext{}
	},
}

// newRequestContext returns a new RequestContext from a sync.Pool.
func newRequestContext(r *http.Request) *requestContext {
	rc := requestContextPool.Get().(*requestContext)
	rc.params = nil
	rc.query = &Query{request: r}

	return rc
}

// returnContext returns the RequestContext to the sync.Pool.
func returnContext(rc *requestContext) {
	requestContextPool.Put(rc)
}

// Cache returns cache.Headers for setting Cache-Control headers.
// func (c *RequestContext) Cache() *cache.Headers {
// 	if c.cache == nil {
// 		c.cache = cache.New()
// 	}
// 	return c.cache
// }
