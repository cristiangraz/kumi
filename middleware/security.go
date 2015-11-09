package middleware

import (
	"github.com/cristiangraz/kumi"
)

type (
	// SecurityCheck takes contextual data and validates it
	SecurityCheck func(*kumi.Context) bool
)

var (
	// UnauthorizedHandler is called when the user is not allowed to access the resource.
	// It should return a response.
	// If you need custom responses, you can leave this nil and respond in your
	// SecurityCheck before returning bool false
	UnauthorizedHandler kumi.HandlerFunc
)

// Assert is used to ensure all of the expressions are true.
func Assert(expressions ...SecurityCheck) kumi.HandlerFunc {
	return func(c *kumi.Context) {
		if ok := securityCheck(c, false, expressions); ok {
			c.Next()
			return
		}

		if UnauthorizedHandler != nil {
			UnauthorizedHandler(c)
		}
	}
}

// AssertNot is used to ensure all of the expressions are false
func AssertNot(expressions ...SecurityCheck) kumi.HandlerFunc {
	return func(c *kumi.Context) {
		if ok := securityCheck(c, true, expressions); ok {
			c.Next()
			return
		}

		if UnauthorizedHandler != nil {
			UnauthorizedHandler(c)
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
