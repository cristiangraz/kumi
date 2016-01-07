package middleware

import (
	"github.com/cristiangraz/kumi"
)

type (
	// SecurityCheck takes contextual data and validates it
	SecurityCheck func(*kumi.Context) bool
)

var (
	// AccessDeniedHandler is called when the user is not allowed to access
	// a resource or perform some action.
	// The AccessDeniedHandler is expected to return a response.
	AccessDeniedHandler kumi.HandlerFunc
)

// Assert is used to ensure all of the expressions are true.
// Assertions occur after authorization, so any SecurityCheck
// that returns false will be handed off to the
// AccessDeniedHandler.
func Assert(expressions ...SecurityCheck) kumi.HandlerFunc {
	return func(c *kumi.Context) {
		if ok := securityCheck(c, false, expressions); ok {
			c.Next()
			return
		}

		if AccessDeniedHandler != nil {
			AccessDeniedHandler(c)
		}
	}
}

// AssertNot is used to ensure all of the expressions are false.
// Assertions occur after authorization, so any SecurityCheck
// that returns true will be handed off to the
// AccessDeniedHandler.
func AssertNot(expressions ...SecurityCheck) kumi.HandlerFunc {
	return func(c *kumi.Context) {
		if ok := securityCheck(c, true, expressions); ok {
			c.Next()
			return
		}

		if AccessDeniedHandler != nil {
			AccessDeniedHandler(c)
		}
	}
}

// securityCheck is the internal function that validates the expressions and returns a boolean
func securityCheck(c *kumi.Context, negate bool, expressions []SecurityCheck) bool {
	for _, fn := range expressions {
		if ok := fn(c); negate == ok {
			return false
		}
	}

	return true
}
