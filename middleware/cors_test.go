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
		options    *middleware.CorsOptions // the cors config
		reqHeaders map[string]string       // the request headers to send
		handlers   []string                // the HTTP methods that should be registered for handling
		method     string                  // the http method to request
		headers    map[string]string       // expected headers
		statusCode int                     // expected status code
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
			method:   "OPTIONS",
			handlers: []string{"PUT", "DELETE"},
			headers: map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "PUT, OPTIONS, DELETE",
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
			method:   "OPTIONS",
			handlers: []string{"GET", "POST"},
			headers: map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "GET, HEAD, POST, OPTIONS",
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
			method:   "OPTIONS",
			handlers: []string{"GET", "POST"},
			headers: map[string]string{
				"Vary": "Origin",
				"Access-Control-Allow-Origin":      "http://kumi.io",
				"Access-Control-Allow-Methods":     "GET, HEAD, POST, OPTIONS",
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
			method:   "OPTIONS",
			handlers: []string{"GET"},
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
			method:   "OPTIONS",
			handlers: []string{"GET"},
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
			reqHeaders: map[string]string{},
			method:     "OPTIONS",
			handlers:   []string{"GET"},
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

	h := func(w http.ResponseWriter, r *http.Request) {}

	for i, tt := range tests {
		w := httptest.NewRecorder()
		r := MustNewRequest(tt.method, "/", nil)
		for k, v := range tt.reqHeaders {
			r.Header.Add(k, v)
		}

		rtr := router.NewHTTPRouter()
		k := kumi.New(rtr)
		k.AutoOptionsMethod()
		k.Use(middleware.Cors(rtr, tt.options))

		if len(tt.handlers) == 0 {
			tt.handlers = []string{tt.method}
		}

		for _, method := range tt.handlers {
			switch method {
			case kumi.GET:
				k.Get("/", h)
			case kumi.HEAD:
				k.Head("/", h)
			case kumi.POST:
				k.Post("/", h)
			case kumi.PUT:
				k.Put("/", h)
			case kumi.PATCH:
				k.Patch("/", h)
			case kumi.OPTIONS:
				k.Options("/", h)
			case kumi.DELETE:
				k.Delete("/", h)
			}
		}
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
	w := httptest.NewRecorder()
	r := MustNewRequest("OPTIONS", "/", nil)
	r.Header.Set("Origin", "http://kumi.io")

	rtr := router.NewHTTPRouter()
	k := kumi.New(rtr)
	k.AutoOptionsMethod()
	k.Use(middleware.Cors(rtr, &middleware.CorsOptions{
		AllowOrigin: []string{"*"},
	}))
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
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

// Ensures multiple origins can be matched to return the correct
// Access-Control-Allow-Origin.
func TestCors_MultipleOrigins(t *testing.T) {
	w := httptest.NewRecorder()
	r := MustNewRequest("OPTIONS", "/", nil)
	r.Header.Set("Origin", "http://bar.com")

	rtr := router.NewHTTPRouter()
	k := kumi.New(rtr)
	k.AutoOptionsMethod()
	k.Use(middleware.Cors(rtr, &middleware.CorsOptions{
		AllowOrigin: []string{"http://foo.com", "http://bar.com"},
	}))
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	k.ServeHTTP(w, r)

	if h := w.Header().Get("Access-Control-Allow-Origin"); h != "http://bar.com" {
		t.Fatalf("unexpected access control allow orign: %s", h)
	}
}

// Ensures no Access-Control-Allow-Origin is set the the Origin sent is
// not in the list of allowed origins.
func TestCors_OriginNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := MustNewRequest("OPTIONS", "/", nil)
	r.Header.Set("Origin", "http://other.com")

	rtr := router.NewHTTPRouter()
	k := kumi.New(rtr)
	k.AutoOptionsMethod()
	k.Use(middleware.Cors(rtr, &middleware.CorsOptions{
		AllowOrigin: []string{"http://foo.com", "http://bar.com"},
	}))
	k.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	k.ServeHTTP(w, r)

	if h := w.Header().Get("Access-Control-Allow-Origin"); h != "" {
		t.Fatalf("unexpected access control allow orign: %s", h)
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
