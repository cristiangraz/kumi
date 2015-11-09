package kumi

import (
	"fmt"
	"net/http"
)

type (
	// Handler is the generic HTTP Handler
	Handler interface{}

	// HandlerFunc is the HTTP handler for kumi middleware.
	HandlerFunc func(*Context)
)

// h takes a Handler and returns a HandlerFunc
func wrapHandler(handler Handler) (HandlerFunc, error) {
	switch h := handler.(type) {
	case func(*Context):
		return h, nil
	case HandlerFunc:
		return h, nil
	case func(http.ResponseWriter, *http.Request):
		return func(c *Context) {
			h(c, c.Request)

			// No handler is passed. Have to trigger it here.
			c.Next()
		}, nil
	case http.Handler:
		return func(c *Context) {
			h.ServeHTTP(c, c.Request)

			// No handler is passed. Have to trigger it here.
			c.Next()
		}, nil
	case func(http.Handler) http.Handler:
		// Note: handler is responsible for calling ServeHTTP on the Context
		// to continue or halt execution.
		return func(c *Context) {
			h(c).ServeHTTP(c, c.Request)
		}, nil
	default:
		return nil, fmt.Errorf("Expected http.HandlerFunc, http.Handler, or kumi.HandlerFunc. Given %T", h)
	}
}

// Wraps one or more handlers.
func wrapHandlers(handlers ...Handler) ([]HandlerFunc, error) {
	var wrapped []HandlerFunc
	for _, handler := range handlers {
		h, err := wrapHandler(handler)
		if err != nil {
			return nil, err
		}

		wrapped = append(wrapped, h)
	}

	return wrapped, nil
}
