package router_test

import (
	"testing"

	"github.com/cristiangraz/kumi/router"
)

func TestStore(t *testing.T) {
	s := &router.Store{}
	for _, pattern := range []string{"/", "/a", "/b", "/a/b/", "/a/:b"} {
		s.Save("GET", pattern)
	}

	tests := []struct {
		pattern string
		valid   bool
	}{
		{"/", true},
		{"/a", true},
		{"/b", true},
		{"/a/b/", true},
		{"/a/:b", true},
		{"/c", false},
		{"/a/b/c", false},
		{"/A", false},
		{"/B", false},
		{"/A/B/", false},
		{"/A/:B", false},
	}

	for i, tt := range tests {
		if got := s.HasRoute("GET", tt.pattern); got != tt.valid {
			t.Fatalf("(%d) expected %v, got %c", i, tt.valid, got)
		}
	}
}
