package validator

import (
	"github.com/cristiangraz/kumi/api"
	"github.com/xeipuuv/gojsonschema"
)

type (
	// Rules defines the mapping to convert error types to Error structs.
	// The string is the name of the field.
	Rules map[string][]Mapping

	// Mapping is the individual mapping for a specific error type within a field.
	Mapping struct {
		Type      string
		ErrorType string
		Message   string
	}
)

// Swap takes json schema errors and swaps them for an array of
// api errors based on mapping rules.
func Swap(errors []gojsonschema.ResultError, rules Rules) []api.Error {
	var e []api.Error
	count := len(errors)
	for _, err := range errors {
		field, errType := err.Field(), err.Type()
		r, ok := rules[field]
		if !ok {
			// check for "global" error field
			r, ok = rules["*"]
			if !ok {
				continue
			}

			if field == "(root)" {
				field = ""
			}
		}

		// The validation failed against oneOf/anyOf/allOf validation, but more errors are returned.
		// Skip returning this error in favor of the other more specific errors.
		if count > 1 && (errType == "number_one_of" || errType == "number_any_of" || errType == "number_all_of") {
			continue
		}

		for _, m := range r {
			if m.Type == errType || m.Type == "*" {
				e = append(e, api.Error{
					Field:   field,
					Type:    m.ErrorType,
					Message: m.Message,
				})

				break
			}
		}
	}

	return e
}
