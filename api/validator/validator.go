package validator

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/cristiangraz/kumi/api"
	"github.com/xeipuuv/gojsonschema"
)

// Validator is a JSON schema and validator. It holds a json schema,
// pointer to a Validator, and optional limit for an io.LimitReader.
type Validator struct {
	Schema  gojsonschema.JSONLoader
	Options *Options
	Limit   int64
}

// NewValidator returns a new Validator. If limit > 0, the limit overwrites
// the limit set in the Validator.
func NewValidator(schema gojsonschema.JSONLoader, options *Options, limit int64) *Validator {
	if options == nil {
		log.Fatal("NewValidator: Options cannot be nil")
	}

	if err := options.Valid(); err != nil {
		log.Fatalf("NewValidation. Invalid Options: %s", err)
	}

	return &Validator{
		Schema:  schema,
		Options: options,
		Limit:   limit,
	}
}

// Valid validates a request against a json schema and handles error responses.
// If the response is successful, dst will be populated.
func (v *Validator) Valid(dst interface{}, w http.ResponseWriter, r *http.Request) (valid bool) {
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
			v.Options.InvalidJSON.SendFormat(w, v.Options.Formatter)
			return false
		case *json.UnmarshalTypeError:
			// Do nothing. Let the validator catch it below so that the API caller
			// receives specific feedback on the error.
		default:
			switch err {
			case io.ErrUnexpectedEOF, io.EOF:
				// If there are no bytes left to read on io.LimitedReader,
				// then we hit a RequestBodyExceeded error.
				if lr, ok := reader.(*io.LimitedReader); ok && lr.N == 0 {
					v.Options.RequestBodyExceeded.SendFormat(w, v.Options.Formatter)
					return false
				}

				// Empty body
				if len(buf.Bytes()) == 0 {
					v.Options.RequestBodyRequired.SendFormat(w, v.Options.Formatter)
					return false
				}

				v.Options.InvalidJSON.SendFormat(w, v.Options.Formatter)
				return false
			default:
				v.Options.InvalidJSON.SendFormat(w, v.Options.Formatter)
				return false
			}
		}
	}

	body := buf.String()

	document := gojsonschema.NewStringLoader(body)
	result, err := gojsonschema.Validate(v.Schema, document)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			v.Options.InvalidJSON.SendFormat(w, v.Options.Formatter)
		default:
			// Most likely an error in the schema.
			v.Options.BadRequest.SendFormat(w, v.Options.Formatter)
		}

		return false
	}

	if result.Valid() {
		return true
	}

	e := Swap(result.Errors(), v.Options.Rules)

	statusCode := http.StatusBadRequest
	if v.Options.ErrorStatus > 0 {
		statusCode = v.Options.ErrorStatus
	}
	api.ErrorResponse(statusCode, e...).SendFormat(w, v.Options.Formatter)

	return false
}
