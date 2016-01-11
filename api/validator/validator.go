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
	Schema  gojsonschema.JSONLoader
	Options *Options
	Limit   int64
}

// NewValidator returns a new Validator. If limit > 0, the limit overwrites
// the limit set in the Validator.
func NewValidator(schema gojsonschema.JSONLoader, options *Options, limit int64) Validator {
	return Validator{
		Schema:  schema,
		Options: options,
		Limit:   limit,
	}
}

// Valid validates a request against a json schema and handles error responses.
// If the response is successful the dst struct will be populated.
func (v Validator) Valid(dst interface{}, w http.ResponseWriter, r *http.Request) bool {
	var reader io.Reader = r.Body
	if v.Limit > 0 {
		reader = io.LimitReader(reader, v.Limit)
	} else if v.Options.Limit > 0 {
		reader = io.LimitReader(reader, v.Options.Limit)
	}

	b := new(bytes.Buffer)
	tee := io.TeeReader(reader, b)
	err := json.NewDecoder(tee).Decode(&dst)
	defer r.Body.Close()
	if err != nil {
		switch err {
		case io.ErrUnexpectedEOF:
			// Request body exceeded
			v.Options.RequestBodyExceeded.SendFormat(w, v.Options.Formatter)
			return false
		case io.EOF:
			// Empty body
			v.Options.RequestBodyRequired.SendFormat(w, v.Options.Formatter)
			return false
		default:
			v.Options.InvalidJSON.SendFormat(w, v.Options.Formatter)
			return false
		}
	}

	body := string(b.Bytes())
	if body == "{}" || body == "[]" {
		// Empty objects and empty arrays are considered empty request
		// bodies for the purposes of the API.
		v.Options.RequestBodyRequired.SendFormat(w, v.Options.Formatter)
		return false
	}

	document := gojsonschema.NewStringLoader(body)
	result, err := gojsonschema.Validate(v.Schema, document)
	if err != nil {
		switch err.(type) {
		case *json.SyntaxError:
			v.Options.InvalidJSON.SendFormat(w, v.Options.Formatter)
			return false
		default:
			// This could mean there is an error processing the schema.
			v.Options.BadRequest.SendFormat(w, v.Options.Formatter)
			return false
		}
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
