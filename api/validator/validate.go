package validator

import (
	"github.com/cristiangraz/kumi"
	"golang.org/x/net/context"
)

const (
	dstKey int = iota
)

// Validate takes a schema and provides validation as middleware.
// If the validation fails your handler will not be called.
// If validation succeeds the payload will be stored in the context.
func Validate(schema Schema, dst interface{}) kumi.HandlerFunc {
	return func(c *kumi.Context) {
		payload := dst
		if schema.Valid(payload, c, c.Request) {
			withContext(c, payload)
			c.Next()
		}
	}
}

// withContext stores the payload in the context.
func withContext(c *kumi.Context, dst interface{}) {
	c.Context = context.WithValue(c.Context, dstKey, dst)
}

// FromContext retrieves the payload from the context.
func FromContext(c *kumi.Context) interface{} {
	return c.Context.Value(dstKey)
}
