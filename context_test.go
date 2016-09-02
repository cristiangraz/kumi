package kumi_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cristiangraz/kumi"
)

// Create a custom context that includes a request-scoped logger.
// The logger could be a struct with more methods, in this case it just holds
// a single function.
type CustomContext struct {
	*kumi.RequestContext
	Log func(msg string)
}

func Log(msg string) {
	log.Println(msg)
}

// App will call Context rather than kumi.Context to access our custom context.
func Context(r *http.Request) *CustomContext {
	return kumi.FromContext(r).(*CustomContext)
}

// Test a custom context without any panics.
func TestContext(t *testing.T) {
	// Middleware uses kumi.SetRequestContext to set our custom context.
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = kumi.SetRequestContext(r, &CustomContext{
				RequestContext: kumi.Context(r),
				Log:            Log,
			})
			next.ServeHTTP(w, r)
		})
	}

	k := kumi.New(&Router{})
	k.Use(mw)
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		Context(r).Log("worked")

		if got := Context(r).Query.Get("name"); got != "ctx" {
			t.Fatalf("invalid query params: %s", got)
		}
	})

	r, _ := http.NewRequest("GET", "/?name=ctx", nil)
	w := httptest.NewRecorder()

	k.ServeHTTP(w, r)
}
