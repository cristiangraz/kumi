package validator

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cristiangraz/kumi/api"
	"github.com/xeipuuv/gojsonschema"
)

var (
	InvalidJSONError         = api.Error{StatusCode: http.StatusBadRequest, Type: "invalid_json", Message: "Invalid or malformed JSON"}
	RequestBodyRequiredError = api.Error{StatusCode: http.StatusBadRequest, Type: "request_body_required", Message: "Request sent with no body"}
	RequestBodyExceededError = api.Error{StatusCode: http.StatusBadRequest, Type: "request_body_exceeded", Message: "Request body exceeded"}
	NotFoundError            = api.Error{StatusCode: http.StatusNotFound, Type: "not_found", Message: "Not found"}
	MethodNotAllowedError    = api.Error{StatusCode: http.StatusMethodNotAllowed, Type: "method_not_allowed", Message: "Method not allowed"}
	AlreadyExistsError       = api.Error{StatusCode: http.StatusConflict, Type: "already_exists", Message: "Another resource has the same value as this field"}
	InvalidContentTypeError  = api.Error{StatusCode: http.StatusUnsupportedMediaType, Type: "invalid_content_type", Message: "Invalid or missing Content-Type header"}
	RequiredError            = api.Error{StatusCode: 422, Type: "required", Message: "Required field missing"}
	InvalidTypeError         = api.Error{StatusCode: 422, Type: "invalid_type", Message: "Field is wrong type. See documentation for more details"}
	InvalidParameterError    = api.Error{StatusCode: 422, Type: "invalid_parameter", Message: "Field is invalid. See documentation for more details"}
	InvalidParametersError   = api.Error{StatusCode: 422, Type: "invalid_parameters", Message: "One or more parameters is invalid"}
	UnknownParameterError    = api.Error{StatusCode: 422, Type: "unknown_parameter", Message: "Unknown parameter sent"}
	InvalidValueError        = api.Error{StatusCode: 422, Type: "invalid_value", Message: "The provided value is invalid"}
	BadRequestError          = api.Error{StatusCode: http.StatusBadRequest, Type: "bad_request", Message: "Bad request."}
	InternalServerError      = api.Error{StatusCode: http.StatusInternalServerError, Type: "server_error", Message: "Internal server error. The error has been logged and we are working on it"}
	ServiceUnavailableError  = api.Error{StatusCode: http.StatusServiceUnavailable, Type: "service_unavailable", Message: "Service unavailable. Please try again shortly"}
)

var (
	validatorOpts = &Options{
		RequestBodyRequired: RequestBodyRequiredError,
		RequestBodyExceeded: RequestBodyExceededError,
		InvalidJSON:         InvalidJSONError,
		BadRequest:          BadRequestError,
		Rules: Rules{
			"*": []Mapping{
				{Type: "required", ErrorType: RequiredError.Type, Message: "Required field missing"},
				{Type: "additional_property_not_allowed", ErrorType: UnknownParameterError.Type, Message: "Unknown parameter sent"},
				{Type: "enum", ErrorType: InvalidValueError.Type, Message: "The provided value is invalid"},
				{Type: "number_one_of", ErrorType: InvalidParametersError.Type, Message: "One or more parameters is invalid"},
				{Type: "number_any_of", ErrorType: InvalidParametersError.Type, Message: "One or more parameters is invalid"},
				{Type: "number_all_of", ErrorType: InvalidParametersError.Type, Message: "One or more parameters is invalid"},
				{Type: "invalid_type", ErrorType: InvalidTypeError.Type, Message: InvalidTypeError.Message},
				{Type: "*", ErrorType: InvalidParameterError.Type, Message: "Field is invalid. See documentation for more details"},
			},
		},
		Limit:       int64(1<<20) + 1, // Limit request body at 1MB
		ErrorStatus: 422,
	}
)

func TestValidator(t *testing.T) {
	schema := `{
        "type": "object",
        "properties": {
            "name": {
                "type": "string"
            },
            "city": {
                "type": "string",
                "enum": ["foo", "bar"]
            }
        },
        "required": ["name"],
        "additionalProperties": false
    }`

	type schemaDest struct {
		Name string `json:"name"`
		City string `json:"string"`
	}

	tests := []struct {
		schema       string
		payload      []byte
		limit        int64
		expectStatus int
		expect       []api.Error
		dst          interface{}
	}{
		{
			schema:  schema,
			payload: []byte(`{"name": "Lilly", "city": "foo"}`),
		},
		{
			schema:       schema,
			payload:      []byte(`{"name": "Lilly", "city": "baz"}`),
			expectStatus: 422,
			expect: []api.Error{
				api.Error{
					Field:   "city",
					Type:    InvalidValueError.Type,
					Message: InvalidValueError.Message,
				},
			},
		},
		{
			// Set limit of 1 byte, exceed it
			schema:       schema,
			payload:      []byte(`{"name": "Lilly", "city": "foo"}`),
			limit:        1,
			expectStatus: http.StatusBadRequest,
			expect: []api.Error{
				api.Error{
					Type:    RequestBodyExceededError.Type,
					Message: RequestBodyExceededError.Message,
				},
			},
		},
		{
			// Set limit of 1 byte, exceed it before JSON is validated
			schema:       schema,
			payload:      []byte(`{ `),
			limit:        1,
			expectStatus: http.StatusBadRequest,
			expect: []api.Error{
				api.Error{
					Type:    RequestBodyExceededError.Type,
					Message: RequestBodyExceededError.Message,
				},
			},
		},
		{
			// Set limit of 1 byte, match the limit exactly and expect a JSON error
			schema:       schema,
			payload:      []byte(`{`),
			limit:        1,
			expectStatus: http.StatusBadRequest,
			expect: []api.Error{
				api.Error{
					Type:    InvalidJSONError.Type,
					Message: InvalidJSONError.Message,
				},
			},
		},
		{
			// Invalid JSON
			schema:       schema,
			payload:      []byte(`{"na`),
			expectStatus: http.StatusBadRequest,
			expect: []api.Error{
				api.Error{
					Type:    InvalidJSONError.Type,
					Message: InvalidJSONError.Message,
				},
			},
		},
		{
			// No body.
			schema:       schema,
			payload:      []byte(``),
			expectStatus: http.StatusBadRequest,
			expect: []api.Error{
				api.Error{
					Type:    RequestBodyRequiredError.Type,
					Message: RequestBodyRequiredError.Message,
				},
			},
		},
		{
			// UnmarshalTypeError
			schema:       schema,
			payload:      []byte(`{"name": {"invalid": "type", "should_be": "string"}}`),
			expectStatus: 422,
			expect: []api.Error{
				api.Error{
					Type:    InvalidTypeError.Type,
					Message: InvalidTypeError.Message,
					Field:   "name",
				},
			},
			dst: schemaDest{},
		},
	}

	for i, tt := range tests {
		if tt.dst == nil {
			tt.dst = ioutil.Discard
		}

		v := New(gojsonschema.NewStringLoader(tt.schema), validatorOpts, tt.limit)
		r, err := http.NewRequest("POST", "/", bytes.NewBuffer(tt.payload))
		if err != nil {
			t.Errorf("TestValidator (%d): Error creating request", i)
		}
		r.Header.Set("Content-Type", "application/json")

		sender := v.Valid(&tt.dst, r)
		if sender != nil && len(tt.expect) == 0 {
			t.Errorf("TestValidator (%d): Expected no errors, one or more given", i)
		}

		if sender == nil && len(tt.expect) > 0 {
			t.Errorf("TestValidator (%d): Expected errors. None given", i)
		}

		if tt.expectStatus > 0 {
			w := httptest.NewRecorder()
			sender.Send(w)
			if w.Code != tt.expectStatus {
				t.Errorf("TestValidator (%d): Expected status code of %d, given %d", i, tt.expectStatus, w.Code)
			}
		}

		if len(tt.expect) == 0 {
			continue
		}

		expect := httptest.NewRecorder()
		given := httptest.NewRecorder()
		api.Failure(tt.expectStatus, tt.expect...).Send(expect)
		sender.Send(given)

		if !reflect.DeepEqual(expect, given) {
			t.Errorf("TestValidator (%d): Expected %v, given %v", i, expect, given)
		}
	}
}

// Tests to make sure more specific validators are used to provide better/more detailed
// error message, and that anyOf/oneOf/allOf methods are handled properly.
func TestSecondaryValidator(t *testing.T) {
	schema := `{
        "type": "object",
        "properties": {
            "type": {
                "type": "string",
				"enum": ["Person", "Company"]
            },
            "name": {
                "type": "string"
            },
			"first_name": {
                "type": "string"
            },
			"last_name": {
                "type": "string"
            }
        },
        "required": ["type"],
        "additionalProperties": false,

		"oneOf": [{
			"properties": {
				"type": {
					"type": "string",
					"enum": ["Person"]
				},
				"first_name": {
					"type": "string",
					"enum": ["Jon", "Sally", "Sarah"]
				},
				"last_name": {
					"type": "string"
				}
			},
			"required": ["first_name", "last_name"]
		}, {
			"properties": {
				"type": {
					"type": "string",
					"enum": ["Company"]
				},
				"name": {
					"type": "string"
				}
			},
			"required": ["name"]
		}]
    }`

	personSchema := gojsonschema.NewStringLoader(`{
		"properties": {
			"type": {
				"type": "string",
				"enum": ["Person"]
			},
			"first_name": {
				"type": "string",
				"enum": ["Jon", "Sally", "Sarah"]
			},
			"last_name": {
				"type": "string"
			}
		},
		"required": ["type", "first_name", "last_name"]
	}`)

	companySchema := gojsonschema.NewStringLoader(`{
		"properties": {
			"type": {
				"type": "string",
				"enum": ["Company"]
			},
			"name": {
				"type": "string"
			}
		},
		"required": ["type", "name"]
	}`)

	type dest struct {
		Type      string `json:"type"`
		Name      string `json:"name,omitempty"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
	}

	tests := []struct {
		schema       string
		payload      []byte
		expectStatus int
		expect       []api.Error
	}{
		{
			schema:       schema,
			payload:      []byte(`{"type": "Person"}`),
			expectStatus: 422,
			expect: []api.Error{
				api.Error{
					Type:    RequiredError.Type,
					Message: RequiredError.Message,
					Field:   "first_name",
				},
				api.Error{
					Type:    RequiredError.Type,
					Message: RequiredError.Message,
					Field:   "last_name",
				},
			},
		},
		{
			schema:       schema,
			payload:      []byte(`{"type": "Company"}`),
			expectStatus: 422,
			expect: []api.Error{
				api.Error{
					Type:    RequiredError.Type,
					Message: RequiredError.Message,
					Field:   "name",
				},
			},
		},
		{
			schema:       schema,
			payload:      []byte(`{"type": "Person", "first_name": "bob"}`),
			expectStatus: 422,
			expect: []api.Error{
				api.Error{
					Type:    RequiredError.Type,
					Message: RequiredError.Message,
					Field:   "last_name",
				},
				api.Error{
					Type:    InvalidValueError.Type,
					Message: InvalidValueError.Message,
					Field:   "first_name",
				},
			},
		},
		{
			schema:       schema,
			payload:      []byte(`{"type": "Person", "first_name": "Sally"}`),
			expectStatus: 422,
			expect: []api.Error{
				api.Error{
					Type:    RequiredError.Type,
					Message: RequiredError.Message,
					Field:   "last_name",
				},
			},
		},
	}

	// secondary validator
	secondary := func(dst interface{}, body string, r *http.Request) (result *gojsonschema.Result, sender api.Sender) {
		data, ok := dst.(*dest)
		if !ok {
			return nil, nil
		}

		if data.Type == "" {
			return nil, RequiredError.With(api.SendInput{
				Field: "type",
			})
		}

		document := gojsonschema.NewStringLoader(body)

		var err error
		switch data.Type {
		case "Person":
			result, err = gojsonschema.Validate(personSchema, document)
		case "Company":
			result, err = gojsonschema.Validate(companySchema, document)
		default:
			return nil, InvalidValueError.With(api.SendInput{
				Field: "type",
			})
		}

		if err != nil {
			return nil, nil
		}

		return result, nil
	}

	for i, tt := range tests {
		var dst dest

		v := New(gojsonschema.NewStringLoader(tt.schema), validatorOpts, 0)
		r, err := http.NewRequest("POST", "/", bytes.NewBuffer(tt.payload))
		if err != nil {
			t.Errorf("TestSecondaryValidator [anyOf/oneOf/allOf] (%d): Error creating request", i)
		}
		r.Header.Set("Content-Type", "application/json")

		sender := v.Valid(&dst, r)
		if sender != nil && len(tt.expect) == 0 {
			t.Errorf("TestSecondaryValidator [anyOf/oneOf/allOf] (%d): Expected no errors, one given", i)
		}

		if sender == nil && len(tt.expect) > 0 {
			t.Errorf("TestSecondaryValidator [anyOf/oneOf/allOf] (%d): Expected errors. None given", i)
		}

		if tt.expectStatus > 0 {
			w := httptest.NewRecorder()
			sender.Send(w)

			if w.Code != tt.expectStatus {
				t.Errorf("TestSecondaryValidator [anyOf/oneOf/allOf] (%d): Expected status code of %d, given %d", i, tt.expectStatus, w.Code)
			}
		}

		if len(tt.expect) == 0 {
			continue
		}

		expect, given := httptest.NewRecorder(), httptest.NewRecorder()
		api.Failure(tt.expectStatus, tt.expect...).Send(expect)
		sender.Send(given)

		if !reflect.DeepEqual(expect, given) {
			t.Errorf("TestSecondaryValidator [anyOf/oneOf/allOf] (%d): Expected %v, given %v", i, expect.Body.String(), given.Body.String())
		}
	}

	// Test secondary validator
	for i, tt := range tests {
		var dst dest

		v := NewSecondary(gojsonschema.NewStringLoader(tt.schema), validatorOpts, 0, secondary)
		r, err := http.NewRequest("POST", "/", bytes.NewBuffer(tt.payload))
		if err != nil {
			t.Errorf("TestSecondaryValidator [secondary] (%d): Error creating request", i)
		}
		r.Header.Set("Content-Type", "application/json")

		sender := v.Valid(&dst, r)
		if sender != nil && len(tt.expect) == 0 {
			t.Errorf("TestSecondaryValidator [secondary] (%d): Expected no errors, one given", i)
		}

		if sender == nil && len(tt.expect) > 0 {
			t.Errorf("TestSecondaryValidator [secondary] (%d): Expected errors. None given", i)
		}

		if tt.expectStatus > 0 {
			w := httptest.NewRecorder()
			sender.Send(w)

			if w.Code != tt.expectStatus {
				t.Errorf("TestSecondaryValidator [secondary] (%d): Expected status code of %d, given %d", i, tt.expectStatus, w.Code)
			}
		}

		if len(tt.expect) == 0 {
			continue
		}

		expect, given := httptest.NewRecorder(), httptest.NewRecorder()
		api.Failure(tt.expectStatus, tt.expect...).Send(expect)
		sender.Send(given)

		if !reflect.DeepEqual(expect, given) {
			t.Errorf("TestSecondaryValidator [secondary] (%d): Expected %v, given %v", i, expect.Body.String(), given.Body.String())
		}
	}
}

// func TestDependency(t *testing.T) {
// 	v := New(gojsonschema.NewStringLoader(`{
//                 "type":"number",
//                 "minimum": 0,
//                 "exclusiveMinimum": true
//             }`), validatorOpts, 0)
// 	w := httptest.NewRecorder()
// 	r, err := http.NewRequest("POST", "/", bytes.NewBufferString(`0`))
// 	if err != nil {
// 		t.Error("TestDependency: Error creating request")
// 	}
// 	r.Header.Set("Content-Type", "application/json")
//
// 	v.Valid(ioutil.Discard, w, r)
// 	log.Println(w.Body.String())
//
// }
