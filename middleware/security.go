package middleware

import "net/http"

type (
	// SecurityCheck takes contextual data and validates it
	SecurityCheck func(r *http.Request) bool
)

// TODO: Replace this with a struct that holds the handler.
var (
	// AccessDeniedHandler is called when the user is not allowed to access
	// a resource or perform some action.
	// The AccessDeniedHandler is expected to return a response.
	AccessDeniedHandler http.HandlerFunc
)

// Assert is used to ensure all of the expressions are true.
// Assertions occur after authorization, so any SecurityCheck
// that returns false will be handed off to the
// AccessDeniedHandler.
func Assert(expressions ...SecurityCheck) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if ok := securityCheck(r, false, expressions); ok {
				next.ServeHTTP(w, r)
				return
			}
			AccessDeniedHandler(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// AssertNot is used to ensure all of the expressions are false.
// Assertions occur after authorization, so any SecurityCheck
// that returns true will be handed off to the
// AccessDeniedHandler.
func AssertNot(expressions ...SecurityCheck) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if ok := securityCheck(r, true, expressions); ok {
				next.ServeHTTP(w, r)
				return
			}
			AccessDeniedHandler(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// securityCheck is the internal function that validates the expressions and returns a boolean
func securityCheck(r *http.Request, negate bool, expressions []SecurityCheck) bool {
	for _, fn := range expressions {
		if ok := fn(r); negate == ok {
			return false
		}
	}
	return true
}
