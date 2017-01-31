package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestErrors_With(t *testing.T) {
	Formatter = JSON

	e := Error{StatusCode: http.StatusBadRequest, Type: "TYPE", Message: "MSG"}
	e2 := e.With(SendInput{Field: "field_name", Message: "MSG2"})
	if !reflect.DeepEqual(e2, Error{StatusCode: http.StatusBadRequest, Type: "TYPE", Field: "field_name", Message: "MSG2"}) {
		t.Fatalf("unexpected error: %#v", e2)
	}
}

func TestErrors_WithField(t *testing.T) {
	e := Error{StatusCode: http.StatusBadRequest, Type: "TYPE", Message: "MSG"}
	e2 := e.WithField("NEWFIELD")
	if !reflect.DeepEqual(e2, Error{StatusCode: http.StatusBadRequest, Type: "TYPE", Field: "NEWFIELD", Message: "MSG"}) {
		t.Fatalf("unexpected error: %#v", e2)
	}
}

func TestErrors_WithMessage(t *testing.T) {
	e := Error{StatusCode: http.StatusBadRequest, Type: "TYPE", Message: "MSG"}
	e2 := e.WithMessage("NEWMSG")
	if !reflect.DeepEqual(e2, Error{StatusCode: http.StatusBadRequest, Type: "TYPE", Message: "NEWMSG"}) {
		t.Fatalf("unexpected error: %#v", e2)
	}
}

func TestErrors_Send(t *testing.T) {
	e := Error{StatusCode: http.StatusBadRequest, Field: "FIELD", Type: "TYPE", Message: "MSG"}

	rec, expected := httptest.NewRecorder(), httptest.NewRecorder()
	e.Send(rec)
	Failure(e.StatusCode, e).SendFormat(expected, JSON)

	if rec.Body.String() != `{"success":false,"status":400,"code":"bad_request","errors":[{"field":"FIELD","type":"TYPE","message":"MSG"}]}`+"\n" {
		t.Fatalf("unexpected response: %s %s", rec.Body.Bytes(), expected.Body.Bytes())
	} else if rec.Code != e.StatusCode {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
}

func TestErrors_SendWith(t *testing.T) {
	e := Error{StatusCode: http.StatusBadRequest, Type: "TYPE", Message: "MSG"}

	rec, expected := httptest.NewRecorder(), httptest.NewRecorder()

	// Override the field and send to generate response.
	e.SendWith(SendInput{Field: "NEW_FIELD"}, rec)

	// Build the Failure manually to get the expected response.
	Failure(e.StatusCode, e.With(SendInput{Field: "NEW_FIELD"})).SendFormat(expected, JSON)

	// Compare.
	if rec.Body.String() != expected.Body.String() {
		t.Fatalf("unexpected response: %s %s", rec.Body.Bytes(), expected.Body.Bytes())
	} else if rec.Code != e.StatusCode {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
}
