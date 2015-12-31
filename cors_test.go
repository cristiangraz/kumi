package kumi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var testHandler HandlerFunc

func init() {
	testHandler, _ = wrapHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
}

func TestCors(t *testing.T) {
	tests := []struct {
		options      *CorsOptions
		routeOptions *CorsOptions
		reqHeaders   map[string]string
		method       string
		headers      map[string]string
	}{
		{
			// No Cors Response when no config
			options: &CorsOptions{},
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
			options:    &CorsOptions{AllowOrigin: []string{"*"}},
			reqHeaders: map[string]string{"Origin": "http://kumi.io"},
			method:     "GET",
			headers: map[string]string{
				"Vary": "",
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Methods":     "",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
		{
			// Test allow specific origin
			options:    &CorsOptions{AllowOrigin: []string{"http://kumi.io"}},
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
			options:    &CorsOptions{AllowOrigin: []string{"http://kumi.io"}},
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
			options: &CorsOptions{
				AllowOrigin:  []string{"http://kumi.io"},
				AllowMethods: []string{"PUT", "DELETE"},
			},
			reqHeaders: map[string]string{
				"Origin":                        "http://kumi.io",
				"Access-Control-Request-Method": "PUT",
			},
			method: "OPTIONS",
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
			options: &CorsOptions{
				AllowOrigin:  []string{"http://kumi.io"},
				AllowMethods: []string{"GET", "POST"},
				AllowHeaders: []string{"Origin"},
			},
			reqHeaders: map[string]string{
				"Origin":                         "http://kumi.io",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "origin",
			},
			method: "OPTIONS",
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
			options: &CorsOptions{
				AllowOrigin:  []string{"http://kumi.io"},
				AllowMethods: []string{"GET", "POST"},
			},
			reqHeaders: map[string]string{
				"Origin":                         "http://kumi.io",
				"Access-Control-Request-Method":  "GET",
				"Access-Control-Request-Headers": "origin",
			},
			method: "OPTIONS",
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
			options: &CorsOptions{
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
			// Test allow credentials
			options: &CorsOptions{
				AllowOrigin:      []string{"http://kumi.io"},
				AllowCredentials: true,
				AllowMethods:     []string{"GET"},
			},
			reqHeaders: map[string]string{
				"Origin":                        "http://kumi.io",
				"Access-Control-Request-Method": "GET",
			},
			method: "OPTIONS",
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
			// Test merge route options with global
			options: &CorsOptions{
				AllowOrigin:  []string{"http://kumi.io"},
				AllowMethods: []string{"GET"},
			},
			routeOptions: &CorsOptions{
				AllowOrigin:      []string{"*"},
				AllowCredentials: true,
				AllowMethods:     []string{"PUT", "DELETE"},
			},
			reqHeaders: map[string]string{
				"Origin":                        "http://kumi.io",
				"Access-Control-Request-Method": "GET",
			},
			method: "OPTIONS",
			headers: map[string]string{
				"Allow": "PUT, DELETE, OPTIONS",
				"Vary":  "",
				"Access-Control-Allow-Origin":      "*",
				"Access-Control-Allow-Methods":     "PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "",
				"Access-Control-Expose-Headers":    "",
			},
		},
	}

	for i, tt := range tests {
		k := New(&dummyRouter{})
		k.SetGlobalCors(tt.options)

		rec := httptest.NewRecorder()
		req, _ := http.NewRequest(tt.method, "/foo", nil)
		for k, v := range tt.reqHeaders {
			req.Header.Add(k, v)
		}

		c := k.NewContext(rec, req, k.CorsOptions(tt.routeOptions), testHandler)
		c.Next()
		k.ReturnContext(c)

		resHeaders := rec.Header()
		for name, value := range tt.headers {
			if actual := strings.Join(resHeaders[name], ", "); actual != value {
				t.Errorf("TestCors (%d): Invalid header %q, wanted %q, got %q", i, name, value, actual)
			}
		}
	}
}
