package middleware

import (
	"net/http"
	"time"

	"github.com/cristiangraz/kumi"
	"golang.org/x/net/context"
)

// Timeout cancels context.Context after a given duration.
func Timeout(timeout time.Duration) kumi.HandlerFunc {
	return func(c *kumi.Context) {
		var cancel context.CancelFunc
		c.Context, cancel = context.WithTimeout(c.Context, timeout)

		defer func() {
			cancel()
			if c.Context.Err() == context.DeadlineExceeded {
				c.WriteHeader(http.StatusGatewayTimeout)
			}
		}()

		c.Next()
	}
}
