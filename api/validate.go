package api

import (
	"github.com/xeipuuv/gojsonschema"
)

type (
	// Rules defines the mapping to convert error types to APIError structs.
	// The string is the name of the field.
	Rules map[string][]Mapping

	// Mapping is the individual mapping for a specific error type within a field.
	Mapping struct {
		Type      string
		ErrorType string
		Message   string
	}
)

// Validate takes schema errors and mapping rules and returns an array of APIError structs
func Validate(errors []gojsonschema.ResultError, rules Rules) []Error {
	var e []Error
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

		for _, m := range r {
			if m.Type == errType || m.Type == "*" {
				e = append(e, Error{
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
