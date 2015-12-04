package middleware

import (
	"net/http"

	"github.com/cristiangraz/kumi"
	"golang.org/x/net/context"
)

// CloseNotify cancels ctx when the underlying connection has gone away.
// It can be used to cancel long operations on the server when the client
// disconnects before the response is ready.
func CloseNotify(c *kumi.Context) {
	// Cancel the context if the client closes the connection
	cn, ok := c.ResponseWriter.(http.CloseNotifier)
	if !ok {
		panic("middleware.CloseNotify expects http.ResponseWriter to implement http.CloseNotifier interface")
	}

	var cancel context.CancelFunc
	c.Context, cancel = context.WithCancel(c.Context)
	defer cancel()

	go func() {
		select {
		case <-c.Context.Done():
			return
		case <-cn.CloseNotify():
			c.WriteHeader(499)
			cancel()
			return
		}
	}()

	c.Next()
}
