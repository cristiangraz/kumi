package validator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/cristiangraz/kumi/api"
	"github.com/xeipuuv/gojsonschema"
)

// Swapper swaps json schema errors for api errors.
type Swapper func(errors []gojsonschema.ResultError, rules Rules) []api.Error

// Validator is a JSON schema validator. It holds a JSON schema,
// pointer to a Validator, and optional limit for an io.LimitReader.
type Validator struct {
	Schema    gojsonschema.JSONLoader
	Options   *Options
	Limit     int64
	secondary SecondaryValidator
}

// SecondaryValidator allows for custom validation logic if the document
// is invalid. See NewSecondaryValidator function for more details.
type SecondaryValidator func(dst interface{}, document *JSONLoader) (result *gojsonschema.Result, sender api.Sender)

// New returns a new Validator. If limit > 0, the limit overwrites
// the limit set in the Validator.
func New(schema gojsonschema.JSONLoader, options *Options, limit int64) *Validator {
	if options == nil {
		panic("validator: options cannot be nil")
	} else if err := options.Valid(); err != nil {
		panic(fmt.Sprintf("validator: invalid options: %s", err))
	} else if options.Swapper == nil {
		options.Swapper = Swap
	}
	return &Validator{
		Schema:  schema,
		Options: options,
		Limit:   limit,
	}
}

// NewSecondary returns a new Validator with a fallback validation function.
// The fallback validation function is useful for returning specific error messages.
// Example: You have a schema validation with oneOf, and if the validation fails
// you have a fallback validator that can provide specific errors based on the
// specific "type" that was submitted.
func NewSecondary(schema gojsonschema.JSONLoader, options *Options, limit int64, secondary SecondaryValidator) *Validator {
	v := New(schema, options, limit)
	v.secondary = secondary

	return v
}

// Valid validates a request against a json schema and handles error responses.
// If the response is successful, dst will be populated.
func (v *Validator) Valid(dst interface{}, r *http.Request) api.Sender {
	if dst == nil {
		panic("dst required")
	}
	defer r.Body.Close()

	limit := v.Options.Limit
	if v.Limit > 0 {
		limit = v.Limit
	}

	limitReader := limitReaderPool.Get().(*io.LimitedReader)
	limitReader.R = r.Body
	limitReader.N = limit + 1 // extend by 1 byte, if N bytes are left to read we've hit max
	defer limitReaderPool.Put(limitReader)

	decoder := json.NewDecoder(limitReader)
	decoder.UseNumber()
	if err := decoder.Decode(&dst); err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			return v.Options.InvalidJSON
		case *json.UnmarshalTypeError:
			// Do nothing. Let the validator catch it below so that the API caller
			// receives specific feedback on the error.
		default:
			switch err {
			case io.ErrUnexpectedEOF, io.EOF:
				if limitReader.N == 0 { // Nothing left to read on io.LimitedReader, body exceeded
					return v.Options.RequestBodyExceeded
				} else if limitReader.N == limit+1 { // Empty body
					return v.Options.RequestBodyRequired
				}
				return v.Options.InvalidJSON
			default:
				return v.Options.InvalidJSON
			}
		}
	}

	document := loaderPool.Get().(*JSONLoader)
	document.dst = dst
	document.i = nil
	defer loaderPool.Put(document)

	result, err := gojsonschema.Validate(v.Schema, document)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			return v.Options.InvalidJSON
		default:
			return v.Options.BadRequest // An error with the schema
		}
	} else if result.Valid() {
		return nil
	}

	// Run through secondary validator
	if v.secondary != nil {
		secondaryResult, sender := v.secondary(dst, document)
		if sender != nil {
			return sender
		} else if secondaryResult != nil {
			result = secondaryResult
		}
	}

	e := Swap(result.Errors(), v.Options.Rules)
	statusCode := http.StatusBadRequest
	if v.Options.ErrorStatus > 0 {
		statusCode = v.Options.ErrorStatus
	}

	return api.Failure(statusCode, e...)
}

var limitReaderPool = &sync.Pool{
	New: func() interface{} {
		return &io.LimitedReader{}
	},
}
