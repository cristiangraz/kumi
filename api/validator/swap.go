package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cristiangraz/kumi/api"
	"github.com/xeipuuv/gojsonschema"
)

// Rules defines the mapping to convert error types to Error structs.
// The string is the name of the field.
type Rules map[string][]Mapping

// Mapping is the individual mapping for a specific error type within a field.
type Mapping struct {
	Type      string
	ErrorType string
	Message   string
}

// regex to find nested fields i.e. names.0, names.1, etc
var rxNestedFields = regexp.MustCompile(`\.[0-9]+$`)

// Swap takes json schema errors and swaps them for an array of
// api errors based on mapping rules.
func Swap(errors []gojsonschema.ResultError, rules Rules) (e []api.Error) {
	count := len(errors)
	used := map[string]bool{}
	for _, err := range errors {
		details := err.Details()
		errType := err.Type()

		// Look for field in either "property" or "field" entries in the details map
		var field string
		if f, ok := details["property"]; ok {
			if f, ok := f.(string); ok {
				field = f
			}
		}
		if field == "" {
			if f, ok := details["field"]; ok {
				if f, ok := f.(string); ok {
					field = f
				}
			}
		}
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

			// Prevent duplicate errors for nested types
			// TODO: tests
			if strings.Contains(field, ".") && rxNestedFields.MatchString(field) {
				field = rxNestedFields.ReplaceAllString(field, "$1")
				key := fmt.Sprintf("%s_%s", field, errType)
				if _, ok := used[key]; ok {
					continue
				}
			}
		}

		// The validation failed against oneOf/anyOf/allOf validation, but more errors are returned.
		// Skip returning this error in favor of the other more specific errors.
		if count > 1 && (errType == "number_one_of" || errType == "number_any_of" || errType == "number_all_of") {
			continue
		}

		for _, m := range r {
			if m.Type == errType || m.Type == "*" {
				// Avoid double-writing errors
				key := fmt.Sprintf("%s_%s", field, m.ErrorType)
				if _, ok := used[key]; ok {
					continue
				}

				used[key] = true
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
