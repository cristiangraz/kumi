package validator

import (
	"errors"

	"github.com/cristiangraz/kumi/api"
)

// Options defines validation rules for validating requests.
type Options struct {
	RequestBodyRequired api.Error
	RequestBodyExceeded api.Error
	InvalidJSON         api.Error
	BadRequest          api.Error
	Rules               Rules

	// Limit is used to create an io.LimitReader when reading the request
	// body. Consider this the global maximum... each validator can contain
	// a specific limit that will override this value.
	Limit int64

	// ErrorStatus is the status code to use in the response for schema errors.
	// If left empty a 400 Bad Request code will be used.
	ErrorStatus int

	// Swapper swaps json schema errors for api errors. If none is provided,
	// the Swap function in this package will be used.
	Swapper Swapper
}

var (
	errOptionsFormatterRequired                  = errors.New("options: Formatter is required")
	errOptionsRequestBodyHandlerRequired         = errors.New("options: RequestBodyRequired handler is nil")
	errOptionsRequestBodyExceededHandlerRequired = errors.New("options: RequestBodyExceeded handler is nil")
	errOptionsInvalidJSONHandlerRequired         = errors.New("options: InvalidJSON handler is nil")
	errOptionsBadRequestHandlerRequired          = errors.New("options: BadRequest handler is nil")
	errOptionsRulesRequired                      = errors.New("options: At least one rule is required")
)

// Valid ensures the options are valid.
func (o Options) Valid() error {
	if o.RequestBodyRequired.StatusCode == 0 {
		return errOptionsRequestBodyHandlerRequired
	} else if o.RequestBodyExceeded.StatusCode == 0 {
		return errOptionsRequestBodyExceededHandlerRequired
	} else if o.InvalidJSON.StatusCode == 0 {
		return errOptionsInvalidJSONHandlerRequired
	} else if o.BadRequest.StatusCode == 0 {
		return errOptionsBadRequestHandlerRequired
	} else if len(o.Rules) == 0 {
		return errOptionsRulesRequired
	}
	return nil
}
