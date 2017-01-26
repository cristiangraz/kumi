package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cristiangraz/kumi"
)

// CorsOptions provides settings for CORS.
type CorsOptions struct {
	// Configures the Access-Control-Allow-Origin header.
	// Set to "*" to allow all.
	AllowOrigin []string

	// Configures the Access-Control-Allow-Credentials header.
	// Sets the value to "true" if set to bool true.
	AllowCredentials bool

	// Sets the Access-Control-Expose-Headers header.
	ExposeHeaders []string

	// Configures the Access-Control-Max-Age header.
	MaxAge time.Duration

	// Configures the Access-Control-Allow-Headers header.
	// If this is empty, deafults to reflecting the headers specified
	// in the request's Access-Control-Request-Headers.
	AllowHeaders []string
}

// Cors handles CORS requests by setting the appropriate
// response headers.
func Cors(checker kumi.RouteChecker, opt *CorsOptions) func(next http.Handler) http.Handler {
	if opt == nil {
		panic("CORS options required")
	}
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == kumi.OPTIONS { // All OPTIONS requests should set the Allow header.
				w.Header().Set("Allow", allowedMethods(checker, r))
			}

			origin := r.Header.Get("Origin")
			if origin == "" { // Not a CORS requests
				if r.Method == kumi.OPTIONS {
					w.WriteHeader(http.StatusNoContent)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			var validOrigin bool
			for _, ao := range opt.AllowOrigin {
				if ao == "*" {
					validOrigin = true
					w.Header().Set("Access-Control-Allow-Origin", origin) // Mirror the origin
					break
				} else if ao == origin {
					validOrigin = true
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}

			// If there is no valid origin match, continue.
			if !validOrigin {
				next.ServeHTTP(w, r)
				return
			}

			if len(opt.AllowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(opt.AllowHeaders, ", "))
			} else if acrh := r.Header.Get("Access-Control-Request-Headers"); acrh != "" {
				// If no allow headers are set, mirror the request headers
				w.Header().Set("Access-Control-Allow-Headers", acrh)
			}

			if len(opt.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(opt.ExposeHeaders, ", "))
			}

			if opt.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if opt.MaxAge.Seconds() > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%.0f", opt.MaxAge.Seconds()))
			}

			// For OPTIONS requests, don't continue to next middleware
			if r.Method == kumi.OPTIONS {
				w.Header().Set("Access-Control-Allow-Methods", allowedMethods(checker, r))
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func allowedMethods(checker kumi.RouteChecker, req *http.Request) string {
	methods := make([]string, 0, len(kumi.HTTPMethods))
	for _, method := range kumi.HTTPMethods {
		if checker.HasRoute(method, req.URL.Path) {
			methods = append(methods, method)
		}
	}
	return strings.Join(methods, ", ")
}
