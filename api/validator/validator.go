package validator

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/cristiangraz/kumi/api"
	"github.com/xeipuuv/gojsonschema"
)

// Swapper swaps json schema errors for api errors.
type Swapper func(errors []gojsonschema.ResultError, rules Rules) []api.Error

// Validator is a JSON schema and validator. It holds a json schema,
// pointer to a Validator, and optional limit for an io.LimitReader.
type Validator struct {
	Schema    gojsonschema.JSONLoader
	Options   *Options
	Limit     int64
	secondary SecondaryValidator
}

// SecondaryValidator allows for custom validation logic if the document
// is invalid. See NewSecondaryValidator function for more details.
type SecondaryValidator func(dst interface{}, body string, r *http.Request) (result *gojsonschema.Result, sender api.Sender)

// Match json.UnmarshalTypeError
var rxUnmarshalTypeError = regexp.MustCompile(`^json: cannot unmarshal (?:array|object|number|string|bool) into Go value of type`)

// NewValidator returns a new Validator. If limit > 0, the limit overwrites
// the limit set in the Validator.
func NewValidator(schema gojsonschema.JSONLoader, options *Options, limit int64) *Validator {
	if options == nil {
		log.Fatal("NewValidator: Options cannot be nil")
	}

	if err := options.Valid(); err != nil {
		log.Fatalf("NewValidation. Invalid Options: %s", err)
	}

	if options.Swapper == nil {
		options.Swapper = Swap
	}

	return &Validator{
		Schema:  schema,
		Options: options,
		Limit:   limit,
	}
}

// NewSecondaryValidator returns a new Validator with a fallback validation function.
// The fallback validation function is useful for returning specific error messages.
// Example: You have a schema validation with oneOf, and if the validation fails
// you have a fallback validator that can provide specific errors based on the
// specific "type" that was submitted.
func NewSecondaryValidator(schema gojsonschema.JSONLoader, options *Options, limit int64, secondary SecondaryValidator) *Validator {
	v := NewValidator(schema, options, limit)
	v.secondary = secondary

	return v
}

// Valid validates a request against a json schema and handles error responses.
// If the response is successful, dst will be populated.
func (v *Validator) Valid(dst interface{}, r *http.Request) api.Sender {
	defer r.Body.Close()

	var reader io.Reader = r.Body
	if v.Limit > 0 {
		reader = io.LimitReader(reader, v.Limit)
	} else if v.Options.Limit > 0 {
		reader = io.LimitReader(reader, v.Options.Limit)
	}

	buf := new(bytes.Buffer)
	tee := io.TeeReader(reader, buf)
	if err := json.NewDecoder(tee).Decode(&dst); err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			return v.Options.InvalidJSON
		case *json.UnmarshalTypeError:
			// Do nothing. Let the validator catch it below so that the API caller
			// receives specific feedback on the error.
		default:
			switch err {
			case io.ErrUnexpectedEOF, io.EOF:
				// If there are no bytes left to read on io.LimitedReader,
				// then we hit a RequestBodyExceeded error.
				if lr, ok := reader.(*io.LimitedReader); ok && lr.N == 0 {
					return v.Options.RequestBodyExceeded
				}

				// Empty body
				if len(buf.Bytes()) == 0 {
					return v.Options.RequestBodyRequired
				}

				return v.Options.InvalidJSON
			default:
				// If not a json.UnmarshalTypeError, send InvalidJSON error.
				// TODO: Why is Go returning an *errors.errorString rather than
				// *json.UnmarshalTypeError that we could catch above?
				if !rxUnmarshalTypeError.MatchString(err.Error()) {
					return v.Options.InvalidJSON
				}
			}
		}
	}

	body := buf.String()

	document := gojsonschema.NewStringLoader(body)
	result, err := gojsonschema.Validate(v.Schema, document)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			return v.Options.InvalidJSON
		default:
			// Most likely an error in the schema.
			return v.Options.BadRequest
		}
	}

	if result.Valid() {
		return nil
	}

	if v.secondary != nil {
		secondaryResult, sender := v.secondary(dst, body, r)
		if sender != nil {
			return sender
		}

		if secondaryResult != nil {
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
