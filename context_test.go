package kumi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cristiangraz/kumi"
)

// Test a custom context without any panics.
func TestContext(t *testing.T) {
	k := kumi.New(&Router{})
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if got := kumi.Context(r).Query().Get("name"); got != "ctx" {
			t.Fatalf("invalid query params: %s", got)
		}
	})

	r, _ := http.NewRequest("GET", "/?name=ctx", nil)
	w := httptest.NewRecorder()

	k.ServeHTTP(w, r)
}
