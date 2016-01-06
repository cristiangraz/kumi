package validator

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/cristiangraz/kumi/api"
	"github.com/xeipuuv/gojsonschema"
)

// Schema is a JSON schema and validator. It holds a json schema,
// pointer to a Validator, and optional limit for an io.LimitReader.
type Schema struct {
	schema  gojsonschema.JSONLoader
	options *Options
	limit   int64
}

// NewSchema returns a new Schema. If limit > 0, the limit overwrites
// the limit set in the Validator.
func NewSchema(schema gojsonschema.JSONLoader, options *Options, limit int64) Schema {
	return Schema{
		schema:  schema,
		options: options,
		limit:   limit,
	}
}

// Valid validates a schema against a validator and handles error responses.
// If the response is successful the dst struct will be populated.
func (s Schema) Valid(dst interface{}, w http.ResponseWriter, r *http.Request) bool {
	var reader io.Reader = r.Body
	if s.limit > 0 {
		reader = io.LimitReader(reader, s.limit)
	} else if s.options.Limit > 0 {
		reader = io.LimitReader(reader, s.options.Limit)
	}
	err := json.NewDecoder(reader).Decode(&dst)
	defer r.Body.Close()
	if err != nil {
		switch err {
		case io.EOF:
			// Empty body
			s.options.RequestBodyRequired.SendFormat(w, s.options.Formatter)
			return false
		default:
			s.options.InvalidJSON.SendFormat(w, s.options.Formatter)
			return false
		}
	}

	document := gojsonschema.NewGoLoader(dst)
	result, err := gojsonschema.Validate(s.schema, document)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			s.options.InvalidJSON.SendFormat(w, s.options.Formatter)
			return false
		default:
			s.options.BadRequest.SendFormat(w, s.options.Formatter)
			return false
		}
	}

	if result.Valid() {
		return true
	}

	e := Swap(result.Errors(), s.options.Rules)

	statusCode := 422
	if s.options.ErrorStatus > 0 {
		statusCode = s.options.ErrorStatus
	}
	api.ErrorResponse(statusCode, e...).SendFormat(w, s.options.Formatter)

	return false
}
