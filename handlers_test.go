package kumi

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type sHTTP struct{}

func (s sHTTP) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rw.Write([]byte("serveHTTP"))
}

func TestHandlers(t *testing.T) {
	h1 := func(rw http.ResponseWriter, r *http.Request) {}
	if _, err := wrapHandler(h1); err != nil {
		t.Errorf("TestHandlers: http.HandlerFunc. Error: %s", err)
	}

	h2 := func(c *Context) {}
	if _, err := wrapHandler(h2); err != nil {
		t.Errorf("TestHandlers: kumi.HandlerFunc. Error: %s", err)
	}

	h3 := sHTTP{}
	if _, err := wrapHandler(h3); err != nil {
		t.Errorf("TestHandlers: http.Handler. Error: %s", err)
	}

	h4 := func() {}
	if _, err := wrapHandler(h4); err == nil {
		t.Errorf("TestHandlers: Expected invalid handler to return an error. None given.")
	}

	if _, err := wrapHandlers(h2, h4); err == nil {
		t.Errorf("TestHandlers: Expected invalid handler on wrapHandlers to return an error. None given.")
	}
}

func TestNetHTTPCompatibility(t *testing.T) {
	// Calls the next handler
	hf := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("X-ServeHTTP-Middleware", "Yes")
			h.ServeHTTP(rw, r)
		})
	}

	// Doesn't call the next handler
	hf2 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("X-ServeHTTP-Middleware-2", "Yes")
			rw.Write([]byte("stopping here"))
		})
	}

	// http.HandlerFunc
	hf3 := func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("X-Response-Writer-Request", "YES")
		rw.Write([]byte("hello"))
	}

	suite := []struct {
		handlers []Handler
		expected string
		headers  map[string]string
	}{
		{
			handlers: []Handler{hf, hf3},
			expected: "hello",
			headers: map[string]string{
				"X-ServeHTTP-Middleware":    "Yes",
				"X-Response-Writer-Request": "YES",
			},
		},
		{
			handlers: []Handler{hf2, hf3},
			expected: "stopping here",
			headers: map[string]string{
				"X-ServeHTTP-Middleware-2": "Yes",
			},
		},
		{
			handlers: []Handler{sHTTP{}, hf},
			expected: "serveHTTP",
			headers: map[string]string{
				"X-ServeHTTP-Middleware": "Yes",
			},
		},
	}

	for _, s := range suite {
		rec, expected := httptest.NewRecorder(), httptest.NewRecorder()
		r, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatalf("TestNetHTTPCompatibility: Error creating request. Error: %s", err)
		}

		wh, err := wrapHandlers(s.handlers...)
		if err != nil {
			t.Fatal("TestNetHTTPCompatibility: Error in expected handlers")
		}

		e := New(&dummyRouter{})
		c := e.NewContext(rec, r, wh...)
		c.Next()
		e.ReturnContext(c)

		expected.Write([]byte(s.expected))
		if !reflect.DeepEqual(expected.Body, rec.Body) {
			t.Errorf("TestNetHTTPCompatibility: Expected body of %s, given %s", expected.Body, rec.Body)
		}

		for k, v := range s.headers {
			if h := rec.Header().Get(k); h != v {
				t.Errorf("TestNetHTTPCompatibility: Expected header %s to equal %s, given %s", k, v, h)
			}
		}
	}
}
