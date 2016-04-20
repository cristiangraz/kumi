package validator

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cristiangraz/kumi/api"
	"github.com/cristiangraz/kumi/api/formatter"
	"github.com/xeipuuv/gojsonschema"
)

// API Errors.
const (
	InvalidJSONError             = "invalid_json"
	RequestBodyRequiredError     = "request_body_required"
	RequestBodyExceededError     = "request_body_exceeded"
	NotFoundError                = "not_found"
	MethodNotAllowedError        = "method_not_allowed"
	AlreadyExistsError           = "already_exists"
	DomainAlreadyRegisteredError = "domain_already_registered"
	InvalidContentTypeError      = "invalid_content_type"
	RequiredError                = "required"
	InvalidTypeError             = "invalid_type"
	InvalidParameterError        = "invalid_parameter"
	InvalidParametersError       = "invalid_parameters"
	UnknownParameterError        = "unknown_parameter"
	InvalidValueError            = "invalid_value"
	DomainContactRequiredError   = "domain_contact_required"
	BadRequestError              = "bad_request"
	InternalServerError          = "server_error"
	ServiceUnavailableError      = "service_unavailable"
)

var errorCollection = api.ErrorCollection{
	InvalidJSONError:             {StatusCode: http.StatusBadRequest, Error: api.Error{Type: InvalidJSONError, Message: "Invalid or malformed JSON"}},
	RequestBodyRequiredError:     {StatusCode: http.StatusBadRequest, Error: api.Error{Type: RequestBodyRequiredError, Message: "Request sent with no body"}},
	RequestBodyExceededError:     {StatusCode: http.StatusBadRequest, Error: api.Error{Type: RequestBodyExceededError, Message: "Request body exceeded"}},
	NotFoundError:                {StatusCode: http.StatusNotFound, Error: api.Error{Type: NotFoundError, Message: "Not found"}},
	MethodNotAllowedError:        {StatusCode: http.StatusMethodNotAllowed, Error: api.Error{Type: MethodNotAllowedError, Message: "Method not allowed"}},
	AlreadyExistsError:           {StatusCode: http.StatusConflict, Error: api.Error{Type: AlreadyExistsError, Message: "Another resource has the same value as this field"}},
	DomainAlreadyRegisteredError: {StatusCode: http.StatusConflict, Error: api.Error{Type: DomainAlreadyRegisteredError, Message: "One or more domains is not available for registration"}},
	InvalidContentTypeError:      {StatusCode: http.StatusUnsupportedMediaType, Error: api.Error{Type: InvalidContentTypeError, Message: "Invalid or missing Content-Type header"}},
	RequiredError:                {StatusCode: 422, Error: api.Error{Type: RequiredError, Message: "Required field missing"}},
	InvalidTypeError:             {StatusCode: 422, Error: api.Error{Type: InvalidTypeError, Message: "Field is wrong type. See documentation for more details"}},
	InvalidParameterError:        {StatusCode: 422, Error: api.Error{Type: InvalidParameterError, Message: "Field is invalid. See documentation for more details"}},
	InvalidParametersError:       {StatusCode: 422, Error: api.Error{Type: InvalidParametersError, Message: "One or more parameters is invalid"}},
	UnknownParameterError:        {StatusCode: 422, Error: api.Error{Type: UnknownParameterError, Message: "Unknown parameter sent"}},
	InvalidValueError:            {StatusCode: 422, Error: api.Error{Type: InvalidValueError, Message: "The provided value is invalid"}},
	DomainContactRequiredError:   {StatusCode: 422, Error: api.Error{Type: DomainContactRequiredError, Message: "A domain contact is required for this action"}},
	BadRequestError:              {StatusCode: http.StatusBadRequest, Error: api.Error{Type: BadRequestError, Message: "Bad request."}},
	InternalServerError:          {StatusCode: http.StatusInternalServerError, Error: api.Error{Type: InternalServerError, Message: "Internal server error. The error has been logged and we are working on it"}},
	ServiceUnavailableError:      {StatusCode: http.StatusServiceUnavailable, Error: api.Error{Type: ServiceUnavailableError, Message: "Service unavailable. Please try again shortly"}},
}

var (
	validatorOpts = &Options{
		RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
		RequestBodyExceeded: errorCollection.Get(RequestBodyExceededError),
		InvalidJSON:         errorCollection.Get(InvalidJSONError),
		BadRequest:          errorCollection.Get(BadRequestError),
		Rules: Rules{
			"*": []Mapping{
				{Type: "required", ErrorType: RequiredError, Message: "Required field missing"},
				{Type: "additional_property_not_allowed", ErrorType: UnknownParameterError, Message: "Unknown parameter sent"},
				{Type: "enum", ErrorType: InvalidValueError, Message: "The provided value is invalid"},
				{Type: "number_one_of", ErrorType: InvalidParametersError, Message: "One or more parameters is invalid"},
				{Type: "number_any_of", ErrorType: InvalidParametersError, Message: "One or more parameters is invalid"},
				{Type: "number_all_of", ErrorType: InvalidParametersError, Message: "One or more parameters is invalid"},
				{Type: "invalid_type", ErrorType: InvalidTypeError, Message: errorCollection.Get(InvalidTypeError).Message},
				{Type: "*", ErrorType: InvalidParameterError, Message: "Field is invalid. See documentation for more details"},
			},
		},
		Limit:       int64(1<<20) + 1, // Limit request body at 1MB
		ErrorStatus: 422,
	}
)

func init() {
	api.Formatter = formatter.JSON
}

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
					Type:    errorCollection.Get(InvalidValueError).Type,
					Message: errorCollection.Get(InvalidValueError).Message,
				},
			},
		},
		{
			// Set tiny limit of 1 byte, exceed it
			schema:       schema,
			payload:      []byte(`{"name": "Lilly", "city": "foo"}`),
			limit:        1,
			expectStatus: http.StatusBadRequest,
			expect: []api.Error{
				api.Error{
					Type:    errorCollection.Get(RequestBodyExceededError).Type,
					Message: errorCollection.Get(RequestBodyExceededError).Message,
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
					Type:    errorCollection.Get(InvalidJSONError).Type,
					Message: errorCollection.Get(InvalidJSONError).Message,
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
					Type:    errorCollection.Get(RequestBodyRequiredError).Type,
					Message: errorCollection.Get(RequestBodyRequiredError).Message,
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
					Type:    errorCollection.Get(InvalidTypeError).Type,
					Message: errorCollection.Get(InvalidTypeError).Message,
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

		v := NewValidator(gojsonschema.NewStringLoader(tt.schema), validatorOpts, tt.limit)
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
		api.ErrorResponse(tt.expectStatus, tt.expect...).Send(expect)
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
					Type:    errorCollection.Get(RequiredError).Type,
					Message: errorCollection.Get(RequiredError).Message,
					Field:   "first_name",
				},
				api.Error{
					Type:    errorCollection.Get(RequiredError).Type,
					Message: errorCollection.Get(RequiredError).Message,
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
					Type:    errorCollection.Get(RequiredError).Type,
					Message: errorCollection.Get(RequiredError).Message,
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
					Type:    errorCollection.Get(RequiredError).Type,
					Message: errorCollection.Get(RequiredError).Message,
					Field:   "last_name",
				},
				api.Error{
					Type:    errorCollection.Get(InvalidValueError).Type,
					Message: errorCollection.Get(InvalidValueError).Message,
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
					Type:    errorCollection.Get(RequiredError).Type,
					Message: errorCollection.Get(RequiredError).Message,
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
			return nil, errorCollection.Get(RequiredError).With(api.SendInput{
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
			return nil, errorCollection.Get(InvalidValueError).With(api.SendInput{
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

		v := NewValidator(gojsonschema.NewStringLoader(tt.schema), validatorOpts, 0)
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
		api.ErrorResponse(tt.expectStatus, tt.expect...).Send(expect)
		sender.Send(given)

		if !reflect.DeepEqual(expect, given) {
			t.Errorf("TestSecondaryValidator [anyOf/oneOf/allOf] (%d): Expected %v, given %v", i, expect.Body.String(), given.Body.String())
		}
	}

	// Test secondary validator
	for i, tt := range tests {
		var dst dest

		v := NewSecondaryValidator(gojsonschema.NewStringLoader(tt.schema), validatorOpts, 0, secondary)
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
		api.ErrorResponse(tt.expectStatus, tt.expect...).Send(expect)
		sender.Send(given)

		if !reflect.DeepEqual(expect, given) {
			t.Errorf("TestSecondaryValidator [secondary] (%d): Expected %v, given %v", i, expect.Body.String(), given.Body.String())
		}
	}
}

// func TestDependency(t *testing.T) {
// 	v := NewValidator(gojsonschema.NewStringLoader(`{
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
