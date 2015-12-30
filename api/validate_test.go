package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

type errorStruct struct {
	Type    string
	Message string
}

const (
	errRequired = iota
	errInvalidParameter
	errInvalidParameters
	errUnknownParameter
	errInvalidValue
)

var errorMap = map[int]errorStruct{
	errRequired:          errorStruct{"required", "Required field missing"},
	errInvalidParameter:  errorStruct{"invalid_parameter", "Field is invalid. See documentation for more details"},
	errInvalidParameters: errorStruct{"invalid_parameters", "One or more parameters is invalid."},
	errUnknownParameter:  errorStruct{"unknown_parameter", "Unknown parameter sent"},
	errInvalidValue:      errorStruct{"invalid_value", "The provided value is invalid"},
}

func TestValidation(t *testing.T) {
	rules := Rules{
		"*": []Mapping{
			makeRule("required", errRequired),
			makeRule("additional_property_not_allowed", errUnknownParameter),
			makeRule("enum", errInvalidValue),
			makeRule("number_one_of", errInvalidParameters),
			makeRule("number_any_of", errInvalidParameters),
			makeRule("number_all_of", errInvalidParameters),
			makeRule("*", errInvalidParameter),
		},
	}

	tests := []struct {
		document string
		schema   string
		expected []Error
	}{
		{
			document: `{"type":"user"}`,
			schema:   `{"type": "object", "properties": {"name": { "type": "string"}}, "required": ["name"]}`,
			expected: []Error{Error{Field: "name", Type: "required", Message: "Required field missing"}},
		},
		{
			document: `{"type":"user"}`,
			schema:   `{"type": "object", "properties": {"name": { "type": "string"}}, "required": ["name"], "additionalProperties": false}`,
			expected: []Error{
				Error{Field: "name", Type: "required", Message: "Required field missing"},
				Error{Field: "type", Type: "unknown_parameter", Message: "Unknown parameter sent"},
			},
		},
		{
			document: `{"type":"user"}`,
			schema:   `{"type": "object", "properties": {"type": { "type": "string", "enum": ["document", "object"]}}, "required": ["type"], "additionalProperties": false}`,
			expected: []Error{
				Error{Field: "type", Type: "invalid_value", Message: "The provided value is invalid"},
			},
		},
		{
			document: `{"type":"user"}`,
			schema:   `{"type": "object", "properties": {"type": { "type": "string", "pattern": "^(document|object)$"}}, "required": ["type"], "additionalProperties": false}`,
			expected: []Error{
				Error{Field: "type", Type: "invalid_parameter", Message: "Field is invalid. See documentation for more details"},
			},
		},
	}

	for i, tt := range tests {
		document := gojsonschema.NewStringLoader(tt.document)
		schema := gojsonschema.NewStringLoader(tt.schema)
		result, err := gojsonschema.Validate(schema, document)
		if err != nil {
			switch err.(type) {
			case *json.SyntaxError:
				t.Fatalf("TestValidation (%d): Syntax error with your json. Please fix. Error: %s", i, err)
			}

			t.Fatalf("TestValidation (%d): Error with your json inputs for test. Error: %s", i, err)
		}

		if result.Valid() {
			t.Errorf("TestValidation (%d): Expected error. None given.", i)
			continue
		}

		given := Validate(result.Errors(), rules)
		if !reflect.DeepEqual(tt.expected, given) {
			t.Errorf("TestValidation (%d): Expected %+v, given %+v", i, tt.expected, given)
		}
	}
}

func makeRule(field string, err int) Mapping {
	if es, ok := errorMap[err]; ok {
		return Mapping{
			Type:      field,
			ErrorType: es.Type,
			Message:   es.Message,
		}
	}

	panic(fmt.Sprintf("Unknown error '%d'", err))
}
