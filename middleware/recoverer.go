package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/cristiangraz/kumi"
)

// Recoverer returns a recoverer function to recover from panics.
func Recoverer(c *kumi.Context) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			kumi.NewContextWithException(c, err)
			http.Error(c, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}()

	c.Next()
}
