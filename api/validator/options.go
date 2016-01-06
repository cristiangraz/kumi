package validator

import "github.com/cristiangraz/kumi/api"

// Options defines validation rules for validating requests.
type Options struct {
	Formatter           api.FormatterFn
	RequestBodyRequired api.StatusError
	InvalidJSON         api.StatusError
	BadRequest          api.StatusError
	Rules               Rules

	// Limit is used to create an io.LimitReader when reading the request
	// body. Consider this the global maximum... each schema can contain
	// a specific limit that will override this value.
	Limit int64

	// ErrorStatus is the status code to use in the response for schema errors.
	// If left empty a 422 Unprocessable Entity code will be used.
	ErrorStatus int
}
