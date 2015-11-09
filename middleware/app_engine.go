package middleware

import (
	"github.com/cristiangraz/kumi"
	"google.golang.org/appengine"
)

// AppEngine sets up the context for use with app engine.
func AppEngine() kumi.HandlerFunc {
	return func(c *kumi.Context) {
		c.Context = appengine.WithContext(c.Context, c.Request)

		c.Next()
	}
}
