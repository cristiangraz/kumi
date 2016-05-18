package kumi

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/imdario/mergo"
)

// CorsOptions provides settings for CORS.
// The recommended approach is to set this globally, then
// provide route-specific overrides on a per-route or per-group
// basis.
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

	// The list of allowed methods for the route. HEAD is automatically
	// added for all GET requests. OPTIONS is automatically added
	// for all requests. Both are handled by Kumi automatically.
	AllowMethods []string

	// Configures the Access-Control-Allow-Headers header.
	// If this is empty, deafults to reflecting the headers specified
	// in the request's Access-Control-Request-Headers.
	AllowHeaders []string
}

const (
	corsOrigin           = "Origin"
	corsAllowOrigin      = "Access-Control-Allow-Origin"
	corsAllowHeaders     = "Access-Control-Allow-Headers"
	corsExposeHeaders    = "Access-Control-Expose-Headers"
	corsAllowCredentials = "Access-Control-Allow-Credentials"
	corsMaxAge           = "Access-Control-Max-Age"
	corsAllowMethods     = "Access-Control-Allow-Methods"
	corsRequestHeaders   = "Access-Control-Request-Headers"
)

// SetGlobalCors sets global CORS settings.
// This is the basis for all CORS responses.
func (e *Engine) SetGlobalCors(cors *CorsOptions) {
	e.cors = cors
}

// CorsOptions handles CORS requests by setting the appropriate
// response headers.
func (e *Engine) CorsOptions(ac *CorsOptions) HandlerFunc {
	if ac == nil {
		if e.cors == nil {
			log.Fatal("Cannot use CorsHandler without CorsOptions settings")
		}

		ac = e.cors
	} else if e.cors != nil {
		if err := mergo.Merge(ac, e.cors); err != nil {
			log.Fatalf("Error merging CorsOptions. Error: %s", err)
		}
	}

	var hasGet, hasHead, hasOptions bool
	for _, m := range ac.AllowMethods {
		switch m {
		case GET:
			hasGet = true
		case HEAD:
			hasHead = true
		case OPTIONS:
			hasOptions = true
		}
	}

	// Create a local copy to modify
	allowMethods := ac.AllowMethods

	// Add HEAD to list of allowed methods when GET is allowed.
	if hasGet && !hasHead {
		allowMethods = append(allowMethods, HEAD)
	}

	// Add OPTIONS to list of allowed methods
	if !hasOptions {
		allowMethods = append(allowMethods, OPTIONS)
	}

	return func(c *Context) {
		if c.Request.Method == OPTIONS {
			// All OPTIONS requests should set the Allow header.
			c.Header().Set("Allow", strings.Join(allowMethods, ", "))
		}

		reqOrigin := c.Request.Header.Get(corsOrigin)
		if reqOrigin == "" {
			// This is not a CORS requests
			if c.Request.Method == OPTIONS {
				c.WriteHeader(http.StatusNoContent)
				return
			}

			c.Next()
			return
		}

		allow := false
		for _, ao := range ac.AllowOrigin {
			if ao == "*" {
				c.Header().Set(corsAllowOrigin, "*")
				allow = true
				break
			} else if ao == reqOrigin {
				c.Header().Set("Vary", "Origin")
				c.Header().Set(corsAllowOrigin, ao)
				allow = true
				break
			}
		}

		if !allow {
			return
		}

		if len(ac.AllowHeaders) > 0 {
			c.Header().Set(corsAllowHeaders, strings.Join(ac.AllowHeaders, ", "))
		} else if acrh := c.Request.Header.Get(corsRequestHeaders); acrh != "" {
			// If no allow headers are set, mirror the request headers
			c.Header().Set(corsAllowHeaders, acrh)
		}

		if len(ac.ExposeHeaders) > 0 {
			c.Header().Set(corsExposeHeaders, strings.Join(ac.ExposeHeaders, ", "))
		}

		if ac.AllowCredentials {
			c.Header().Set(corsAllowCredentials, "true")
		}

		if ac.MaxAge.Seconds() > 0 {
			c.Header().Set(corsMaxAge, fmt.Sprintf("%.0f", ac.MaxAge.Seconds()))
		}

		// For OPTIONS requests, don't continue to next middleware
		if c.Request.Method == OPTIONS {
			c.Header().Set(corsAllowMethods, strings.Join(allowMethods, ", "))
			c.WriteHeader(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Cors is a convenience middleware that defines HTTP method overrides
// to the global cors settings. This assists with readability when
// all you need is to specify route-specific methods to
// override the global CORS settings.
func (e *Engine) Cors(methods ...string) HandlerFunc {
	return e.CorsOptions(&CorsOptions{
		AllowMethods: methods,
	})
}
