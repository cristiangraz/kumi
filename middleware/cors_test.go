package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cristiangraz/kumi"
	"github.com/cristiangraz/kumi/middleware"
	"github.com/cristiangraz/kumi/router"
)

func TestCors(t *testing.T) {
	tests := []struct {
		options      *middleware.CorsOptions
		reqHeaders   map[string]string
		method       string
		allowMethods []string
		headers      map[string]string
		statusCode   int
	}{
		{
			// No Cors Response when no config
			options: &middleware.CorsOptions{},
			method:  "GET",
			headers: map[string]string{
				"Vary": "",
				"Access-Control-Allow-Origin":      "",
				"Access-Control-Allow-Methods":     "",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test allow all origin
			options:    &middleware.CorsOptions{AllowOrigin: []string{"*"}},
			reqHeaders: map[string]string{"Origin": "http://kumi.io"},
			method:     "GET",
			headers: map[string]string{
				"Vary": "",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test allow specific origin
			options:    &middleware.CorsOptions{AllowOrigin: []string{"http://kumi.io"}},
			reqHeaders: map[string]string{"Origin": "http://kumi.io"},
			method:     "GET",
			headers: map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test disallowed origin
			options:    &middleware.CorsOptions{AllowOrigin: []string{"http://kumi.io"}},
			reqHeaders: map[string]string{"Origin": "http://baz.com"},
			method:     "GET",
			headers: map[string]string{
				"Vary": "",
				"Access-Control-Allow-Origin":      "",
				"Access-Control-Allow-Methods":     "",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test allowed method
			options: &middleware.CorsOptions{
				AllowOrigin: []string{"http://kumi.io"},
			},
			reqHeaders: map[string]string{
				"Origin":                        "http://kumi.io",
				"Access-Control-Request-Method": "PUT",
			},
			method:       "OPTIONS",
			allowMethods: []string{"PUT", "DELETE"},
			headers: map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test allow headers
			options: &middleware.CorsOptions{
				AllowOrigin:  []string{"http://kumi.io"},
				AllowHeaders: []string{"Origin"},
			},
			reqHeaders: map[string]string{
				"Origin":                         "http://kumi.io",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "origin",
			},
			method:       "OPTIONS",
			allowMethods: []string{"GET", "POST"},
			headers: map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "GET, POST, HEAD, OPTIONS",
				"Access-Control-Allow-Headers":     "Origin",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test allow headers mirrors when AllowHeaders is not set
			// and Access-Control-Request-Headers is sent with request.
			options: &middleware.CorsOptions{
				AllowOrigin: []string{"http://kumi.io"},
			},
			reqHeaders: map[string]string{
				"Origin":                         "http://kumi.io",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "origin",
			},
			method:       "OPTIONS",
			allowMethods: []string{"GET", "POST"},
			headers: map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "GET, POST, HEAD, OPTIONS",
				"Access-Control-Allow-Headers":     "origin",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test exposed header
			options: &middleware.CorsOptions{
				AllowOrigin:   []string{"http://kumi.io"},
				ExposeHeaders: []string{"X-Header-1", "X-Header-2"},
			},
			reqHeaders: map[string]string{
				"Origin": "http://kumi.io",
			},
			method: "GET",
			headers: map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "X-Header-1, X-Header-2",
			},
		},
		{
			// Test max age
			options: &middleware.CorsOptions{
				AllowOrigin: []string{"http://kumi.io"},
				MaxAge:      time.Duration(24) * time.Hour,
			},
			reqHeaders: map[string]string{
				"Origin": "http://kumi.io",
			},
			method: "GET",
			headers: map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "86400",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test allow credentials
			options: &middleware.CorsOptions{
				AllowOrigin:      []string{"http://kumi.io"},
				AllowCredentials: true,
			},
			reqHeaders: map[string]string{
				"Origin":                        "http://kumi.io",
				"Access-Control-Request-Method": "GET",
			},
			method:       "OPTIONS",
			allowMethods: []string{"GET"},
			headers: map[string]string{
				"Allow": "GET, HEAD, OPTIONS",
				"Vary":  "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "GET, HEAD, OPTIONS",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test HEAD and OPTIONS in allow methods doesn't lead to duplicates
			options: &middleware.CorsOptions{
				AllowOrigin: []string{"*"},
			},
			reqHeaders: map[string]string{
				"Origin":                        "http://kumi.io",
				"Access-Control-Request-Method": "GET",
			},
			method:       "OPTIONS",
			allowMethods: []string{"GET", "HEAD", "OPTIONS"},
			headers: map[string]string{
				"Allow": "GET, HEAD, OPTIONS",
				"Vary":  "",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "GET, HEAD, OPTIONS",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test OPTIONS request with no Origin
			options: &middleware.CorsOptions{
				AllowOrigin: []string{"*"},
			},
			reqHeaders:   map[string]string{},
			method:       "OPTIONS",
			allowMethods: []string{"GET"},
			headers: map[string]string{
				"Allow": "GET, HEAD, OPTIONS",
				"Vary":  "",
				"Access-Control-Allow-Origin":      "",
				"Access-Control-Allow-Methods":     "",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
			statusCode: http.StatusNoContent,
		},
	}

	for i, tt := range tests {
		w := httptest.NewRecorder()
		r := MustNewRequest(tt.method, "/", nil)
		for k, v := range tt.reqHeaders {
			r.Header.Add(k, v)
		}

		k := kumi.New(router.NewHTTPRouter())
		k.AutoOptionsMethod()
		k.Use(tt.options.Allow(tt.allowMethods...))
		k.Get("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == kumi.OPTIONS {
				t.Fatalf("(%d) handler should not run", i)
			}
		})
		k.ServeHTTP(w, r)

		resHeaders := w.Header()
		for name, value := range tt.headers {
			if actual := strings.Join(resHeaders[name], ", "); actual != value {
				t.Errorf("(%d): Invalid header %q, wanted %q, got %q", i, name, value, actual)
			}
		}

		if tt.statusCode > 0 && w.Code != tt.statusCode {
			t.Errorf("(%d): Invalid status code, wanted %d, got %d", i, tt.statusCode, w.Code)
		}
	}
}

func TestCors_Preflight(t *testing.T) {
	mw := (&middleware.CorsOptions{
		AllowOrigin: []string{"*"},
	}).Allow(kumi.GET)

	w := httptest.NewRecorder()
	r := MustNewRequest("OPTIONS", "/", nil)
	r.Header.Set("Origin", "http://kumi.io")

	k := kumi.New(router.NewHTTPRouter())
	k.AutoOptionsMethod()
	k.Group(mw).Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	k.ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("TestCorsPreflight: Expected OPTIONS Preflight request to return cors. %d given", w.Code)
	}

	expected := map[string]string{
		"Allow": "GET, HEAD, OPTIONS",
		"Vary":  "",
		"Access-Control-Allow-Origin":      "http://kumi.io",
		"Access-Control-Allow-Methods":     "GET, HEAD, OPTIONS",
		"Access-Control-Allow-Headers":     "",
		"Access-Control-Allow-Credentials": "",
		"Access-Control-Max-Age":           "",
		"Access-Control-Expose-Headers":    "",
	}

	resHeaders := w.Header()
	for name, value := range expected {
		if actual := strings.Join(resHeaders[name], ", "); actual != value {
			t.Errorf("TestCorsPreflight: Invalid header %q, wanted %q, got %q", name, value, actual)
		}
	}
}

// MustNewRequest returns a new HTTP request. Panic on error.
func MustNewRequest(method, urlStr string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		panic(err)
	}
	return req
}
