package validator

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cristiangraz/kumi/api"
	"github.com/xeipuuv/gojsonschema"
)

func TestSwap(t *testing.T) {
	rules := Rules{
		"*": []Mapping{
			{Type: "required", ErrorType: "required", Message: "Required field missing"},
			{Type: "additional_property_not_allowed", ErrorType: "unknown_parameter", Message: "Unknown parameter sent"},
			{Type: "enum", ErrorType: "invalid_value", Message: "The provided value is invalid"},
			{Type: "number_one_of", ErrorType: "invalid_parameters", Message: "One or more parameters is invalid."},
			{Type: "number_any_of", ErrorType: "invalid_parameters", Message: "One or more parameters is invalid."},
			{Type: "number_all_of", ErrorType: "invalid_parameters", Message: "One or more parameters is invalid."},
			{Type: "*", ErrorType: "invalid_parameter", Message: "Field is invalid. See documentation for more details"},
		},
	}

	tests := []struct {
		document string
		schema   string
		expected []api.Error
	}{
		{
			document: `{"type":"user"}`,
			schema:   `{"type": "object", "properties": {"name": { "type": "string"}}, "required": ["name"]}`,
			expected: []api.Error{api.Error{Field: "name", Type: "required", Message: "Required field missing"}},
		},
		{
			document: `{"type":"user"}`,
			schema:   `{"type": "object", "properties": {"name": { "type": "string"}}, "required": ["name"], "additionalProperties": false}`,
			expected: []api.Error{
				api.Error{Field: "name", Type: "required", Message: "Required field missing"},
				api.Error{Field: "type", Type: "unknown_parameter", Message: "Unknown parameter sent"},
			},
		},
		{
			document: `{"type":"user"}`,
			schema:   `{"type": "object", "properties": {"type": { "type": "string", "enum": ["document", "object"]}}, "required": ["type"], "additionalProperties": false}`,
			expected: []api.Error{
				api.Error{Field: "type", Type: "invalid_value", Message: "The provided value is invalid"},
			},
		},
		{
			document: `{"type":"user"}`,
			schema:   `{"type": "object", "properties": {"type": { "type": "string", "pattern": "^(document|object)$"}}, "required": ["type"], "additionalProperties": false}`,
			expected: []api.Error{
				api.Error{Field: "type", Type: "invalid_parameter", Message: "Field is invalid. See documentation for more details"},
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

		given := Swap(result.Errors(), rules)
		if !reflect.DeepEqual(tt.expected, given) {
			t.Errorf("TestValidation (%d): Expected %+v, given %+v", i, tt.expected, given)
		}
	}
}
