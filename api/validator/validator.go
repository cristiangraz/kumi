package validator

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/cristiangraz/kumi/api"
	"github.com/xeipuuv/gojsonschema"
)

// Validator is a JSON schema and validator. It holds a json schema,
// pointer to a Validator, and optional limit for an io.LimitReader.
type Validator struct {
	schema  gojsonschema.JSONLoader
	options *Options
	limit   int64
}

// NewValidator returns a new Validator. If limit > 0, the limit overwrites
// the limit set in the Validator.
func NewValidator(schema gojsonschema.JSONLoader, options *Options, limit int64) Validator {
	return Validator{
		schema:  schema,
		options: options,
		limit:   limit,
	}
}

// Valid validates a request against a json schema and handles error responses.
// If the response is successful the dst struct will be populated.
func (v Validator) Valid(dst interface{}, w http.ResponseWriter, r *http.Request) bool {
	var reader io.Reader = r.Body
	if v.limit > 0 {
		reader = io.LimitReader(reader, v.limit)
	} else if v.options.Limit > 0 {
		reader = io.LimitReader(reader, v.options.Limit)
	}

	b := new(bytes.Buffer)
	tee := io.TeeReader(reader, b)
	err := json.NewDecoder(tee).Decode(&dst)
	defer r.Body.Close()
	if err != nil {
		switch err {
		case io.EOF:
			// Empty body
			v.options.RequestBodyRequired.SendFormat(w, v.options.Formatter)
			return false
		default:
			v.options.InvalidJSON.SendFormat(w, v.options.Formatter)
			return false
		}
	}

	document := gojsonschema.NewStringLoader(string(b.Bytes()))
	result, err := gojsonschema.Validate(v.schema, document)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			v.options.InvalidJSON.SendFormat(w, v.options.Formatter)
			return false
		default:
			// This could mean there is an error processing your schema.
			v.options.BadRequest.SendFormat(w, v.options.Formatter)
			return false
		}
	}

	if result.Valid() {
		return true
	}

	e := Swap(result.Errors(), v.options.Rules)

	statusCode := 422
	if v.options.ErrorStatus > 0 {
		statusCode = v.options.ErrorStatus
	}
	api.ErrorResponse(statusCode, e...).SendFormat(w, v.options.Formatter)

	return false
}
