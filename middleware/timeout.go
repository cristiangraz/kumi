package middleware

import (
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// Timeout cancels context.Context after a given duration.
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)

			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					w.WriteHeader(http.StatusGatewayTimeout)
				}
			}()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
