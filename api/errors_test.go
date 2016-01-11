package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	InvalidJSONError        = "invalid_json"
	AccessDeniedError       = "access_denied"
	NotFoundError           = "not_found"
	MethodNotAllowedError   = "method_not_allowed"
	AlreadyExistsError      = "already_exists"
	InvalidContentTypeError = "invalid_content_type"
)

func init() {
	Errors = ErrorCollection{
		InvalidJSONError:        {StatusCode: http.StatusBadRequest, Error: Error{Type: InvalidJSONError, Message: "Invalid or malformed JSON"}},
		AccessDeniedError:       {StatusCode: http.StatusForbidden, Error: Error{Type: AccessDeniedError, Message: "Access denied"}},
		NotFoundError:           {StatusCode: http.StatusNotFound, Error: Error{Type: NotFoundError, Message: "Not found"}},
		MethodNotAllowedError:   {StatusCode: http.StatusMethodNotAllowed, Error: Error{Type: MethodNotAllowedError, Message: "Method not allowed"}},
		AlreadyExistsError:      {StatusCode: http.StatusConflict, Error: Error{Type: AlreadyExistsError, Message: "Another resource has the same value as this field"}},
		InvalidContentTypeError: {StatusCode: http.StatusUnsupportedMediaType, Error: Error{Type: InvalidContentTypeError, Message: "Invalid or missing content-type header"}},
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		in   string
		want StatusError
	}{
		{in: "invalid_json", want: StatusError{StatusCode: http.StatusBadRequest, Error: Error{Type: InvalidJSONError, Message: "Invalid or malformed JSON"}}},
		{in: "access_denied", want: StatusError{StatusCode: http.StatusForbidden, Error: Error{Type: AccessDeniedError, Message: "Access denied"}}},
		{in: "not_found", want: StatusError{StatusCode: http.StatusNotFound, Error: Error{Type: NotFoundError, Message: "Not found"}}},
		{in: "method_not_allowed", want: StatusError{StatusCode: http.StatusMethodNotAllowed, Error: Error{Type: MethodNotAllowedError, Message: "Method not allowed"}}},
		{in: "already_exists", want: StatusError{StatusCode: http.StatusConflict, Error: Error{Type: AlreadyExistsError, Message: "Another resource has the same value as this field"}}},
		{in: "invalid_content_type", want: StatusError{StatusCode: http.StatusUnsupportedMediaType, Error: Error{Type: InvalidContentTypeError, Message: "Invalid or missing content-type header"}}},
	}

	// Set Formatter
	Formatter = formatJSON

	fieldName := "field_name"
	msg := "bla bla bla"
	for i, tt := range tests {
		given := GetError(tt.in)
		if !reflect.DeepEqual(tt.want, given) {
			t.Errorf("TestGet (%d): Want %+v, given %+v", i, tt.want, given)
		}

		rec, expected := httptest.NewRecorder(), httptest.NewRecorder()
		given.Send(rec)
		ErrorResponse(tt.want.StatusCode, tt.want.Error).Send(expected)

		if rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("TestErrors (%d): Wrong Content-Type. Want %q, given %q", i, "application/json", rec.Header().Get("Content-Type"))
		}

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}

		rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
		given.SendWith(SendInput{Field: fieldName}, rec)
		ErrorResponse(tt.want.StatusCode, Error{Type: tt.want.Error.Type, Field: fieldName, Message: tt.want.Error.Message}).Send(expected)

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body for SendWith using field. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}

		rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
		given.SendWith(SendInput{Message: msg}, rec)
		ErrorResponse(tt.want.StatusCode, Error{Type: tt.want.Error.Type, Message: msg}).Send(expected)

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body for SendWith using message. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}
	}
}

func TestErrorsFormat(t *testing.T) {
	tests := []struct {
		in   string
		want StatusError
	}{
		{in: "invalid_json", want: StatusError{StatusCode: http.StatusBadRequest, Error: Error{Type: InvalidJSONError, Message: "Invalid or malformed JSON"}}},
		{in: "access_denied", want: StatusError{StatusCode: http.StatusForbidden, Error: Error{Type: AccessDeniedError, Message: "Access denied"}}},
		{in: "not_found", want: StatusError{StatusCode: http.StatusNotFound, Error: Error{Type: NotFoundError, Message: "Not found"}}},
		{in: "method_not_allowed", want: StatusError{StatusCode: http.StatusMethodNotAllowed, Error: Error{Type: MethodNotAllowedError, Message: "Method not allowed"}}},
		{in: "already_exists", want: StatusError{StatusCode: http.StatusConflict, Error: Error{Type: AlreadyExistsError, Message: "Another resource has the same value as this field"}}},
		{in: "invalid_content_type", want: StatusError{StatusCode: http.StatusUnsupportedMediaType, Error: Error{Type: InvalidContentTypeError, Message: "Invalid or missing content-type header"}}},
	}

	fieldName := "field_name"
	msg := "bla bla bla"
	for i, tt := range tests {
		given := GetError(tt.in)
		if !reflect.DeepEqual(tt.want, given) {
			t.Errorf("TestGet (%d): Want %+v, given %+v", i, tt.want, given)
		}

		rec, expected := httptest.NewRecorder(), httptest.NewRecorder()
		given.SendFormat(rec, formatJSON)
		ErrorResponse(tt.want.StatusCode, tt.want.Error).SendFormat(expected, formatJSON)

		if rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("TestErrors (%d): Wrong Content-Type. Want %q, given %q", i, "application/json", rec.Header().Get("Content-Type"))
		}

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}

		rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
		given.SendWithFormat(SendInput{Field: fieldName}, rec, formatJSON)
		ErrorResponse(tt.want.StatusCode, Error{Type: tt.want.Error.Type, Field: fieldName, Message: tt.want.Error.Message}).SendFormat(expected, formatJSON)

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body for SendWithFormat using field. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}

		rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
		given.SendWithFormat(SendInput{Message: "bla bla bla"}, rec, formatJSON)
		ErrorResponse(tt.want.StatusCode, Error{Type: tt.want.Error.Type, Message: msg}).SendFormat(expected, formatJSON)

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body for SendWithFormat using message. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}
	}
}
