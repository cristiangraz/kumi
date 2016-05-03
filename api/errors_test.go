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
		InvalidJSONError:        {StatusCode: http.StatusBadRequest, Type: InvalidJSONError, Message: "Invalid or malformed JSON"},
		AccessDeniedError:       {StatusCode: http.StatusForbidden, Type: AccessDeniedError, Message: "Access denied"},
		NotFoundError:           {StatusCode: http.StatusNotFound, Type: NotFoundError, Message: "Not found"},
		MethodNotAllowedError:   {StatusCode: http.StatusMethodNotAllowed, Type: MethodNotAllowedError, Message: "Method not allowed"},
		AlreadyExistsError:      {StatusCode: http.StatusConflict, Type: AlreadyExistsError, Message: "Another resource has the same value as this field"},
		InvalidContentTypeError: {StatusCode: http.StatusUnsupportedMediaType, Type: InvalidContentTypeError, Message: "Invalid or missing content-type header"},
	}
}

func TestErrors(t *testing.T) {
	tests := []struct {
		in   string
		want Error
	}{
		{in: "invalid_json", want: Error{StatusCode: http.StatusBadRequest, Type: InvalidJSONError, Message: "Invalid or malformed JSON"}},
		{in: "access_denied", want: Error{StatusCode: http.StatusForbidden, Type: AccessDeniedError, Message: "Access denied"}},
		{in: "not_found", want: Error{StatusCode: http.StatusNotFound, Type: NotFoundError, Message: "Not found"}},
		{in: "method_not_allowed", want: Error{StatusCode: http.StatusMethodNotAllowed, Type: MethodNotAllowedError, Message: "Method not allowed"}},
		{in: "already_exists", want: Error{StatusCode: http.StatusConflict, Type: AlreadyExistsError, Message: "Another resource has the same value as this field"}},
		{in: "invalid_content_type", want: Error{StatusCode: http.StatusUnsupportedMediaType, Type: InvalidContentTypeError, Message: "Invalid or missing content-type header"}},
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
		ErrorResponse(tt.want.StatusCode, tt.want).Send(expected)

		if rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("TestErrors (%d): Wrong Content-Type. Want %q, given %q", i, "application/json", rec.Header().Get("Content-Type"))
		}

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}

		rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
		given.SendWith(SendInput{Field: fieldName}, rec)
		ErrorResponse(tt.want.StatusCode, Error{Type: tt.want.Type, Field: fieldName, Message: tt.want.Message}).Send(expected)

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body for SendWith using field. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}

		rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
		given.SendWith(SendInput{Message: msg}, rec)
		ErrorResponse(tt.want.StatusCode, Error{Type: tt.want.Type, Message: msg}).Send(expected)

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body for SendWith using message. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}
	}
}

func TestErrorsFormat(t *testing.T) {
	tests := []struct {
		in   string
		want Error
	}{
		{in: "invalid_json", want: Error{StatusCode: http.StatusBadRequest, Type: InvalidJSONError, Message: "Invalid or malformed JSON"}},
		{in: "access_denied", want: Error{StatusCode: http.StatusForbidden, Type: AccessDeniedError, Message: "Access denied"}},
		{in: "not_found", want: Error{StatusCode: http.StatusNotFound, Type: NotFoundError, Message: "Not found"}},
		{in: "method_not_allowed", want: Error{StatusCode: http.StatusMethodNotAllowed, Type: MethodNotAllowedError, Message: "Method not allowed"}},
		{in: "already_exists", want: Error{StatusCode: http.StatusConflict, Type: AlreadyExistsError, Message: "Another resource has the same value as this field"}},
		{in: "invalid_content_type", want: Error{StatusCode: http.StatusUnsupportedMediaType, Type: InvalidContentTypeError, Message: "Invalid or missing content-type header"}},
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
		ErrorResponse(tt.want.StatusCode, tt.want).SendFormat(expected, formatJSON)

		if rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("TestErrors (%d): Wrong Content-Type. Want %q, given %q", i, "application/json", rec.Header().Get("Content-Type"))
		}

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}

		rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
		given.With(SendInput{Field: fieldName}).SendFormat(rec, formatJSON)
		ErrorResponse(tt.want.StatusCode, Error{Type: tt.want.Type, Field: fieldName, Message: tt.want.Message}).SendFormat(expected, formatJSON)

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body for SendWithFormat using field. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}

		rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
		given.With(SendInput{Message: "bla bla bla"}).SendFormat(rec, formatJSON)
		ErrorResponse(tt.want.StatusCode, Error{Type: tt.want.Type, Message: msg}).SendFormat(expected, formatJSON)

		if !reflect.DeepEqual(rec, expected) {
			t.Errorf("TestErrors (%d): Wrong response body for SendWithFormat using message. Want %s, given %s", i, rec.Body.Bytes(), expected.Body.Bytes())
		}
	}
}
