package middleware

import (
	"context"
	"log"
	"net/http"
)

// CloseNotify cancels ctx when the underlying connection has gone away.
// It can be used to cancel long operations on the server when the client
// disconnects before the response is ready.
func CloseNotify(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cn, ok := w.(http.CloseNotifier) // Cancel the context if the client closes the connection
		if !ok {
			panic("CloseNotify middleware expects http.ResponseWriter to implement http.CloseNotifier interface")
		}

		ch := cn.CloseNotify()

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		go func() {
			select {
			case <-ctx.Done():
				return
			case <-ch:
				log.Println("request was closed")
				cancel()
				w.WriteHeader(499)
				return
			}
		}()

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
