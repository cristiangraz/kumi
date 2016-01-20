package validator

import (
	"errors"

	"github.com/cristiangraz/kumi/api"
)

// Options defines validation rules for validating requests.
type Options struct {
	Formatter           api.FormatterFn
	RequestBodyRequired api.StatusError
	RequestBodyExceeded api.StatusError
	InvalidJSON         api.StatusError
	BadRequest          api.StatusError
	Rules               Rules

	// Limit is used to create an io.LimitReader when reading the request
	// body. Consider this the global maximum... each schema can contain
	// a specific limit that will override this value.
	Limit int64

	// ErrorStatus is the status code to use in the response for schema errors.
	// If left empty a 400 Bad Request code will be used.
	ErrorStatus int
}

// Valid ensures the options are valid.
func (o Options) Valid() error {
	if o.Formatter == nil {
		return errors.New("Options: Formatter is required")
	}

	if o.RequestBodyRequired.StatusCode == 0 {
		return errors.New("Options: RequestBodyRequired handler is nil")
	}

	if o.RequestBodyExceeded.StatusCode == 0 {
		return errors.New("Options: RequestBodyExceeded handler is nil")
	}

	if o.InvalidJSON.StatusCode == 0 {
		return errors.New("Options: InvalidJSON handler is nil")
	}

	if o.BadRequest.StatusCode == 0 {
		return errors.New("Options: BadRequest handler is nil")
	}

	if len(o.Rules) == 0 {
		return errors.New("Options: At least one rule is required")
	}

	return nil
}
