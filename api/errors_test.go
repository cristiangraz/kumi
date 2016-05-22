package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrors(t *testing.T) {
	Formatter = formatJSON

	e := Error{StatusCode: http.StatusBadRequest, Type: "invalid_json", Message: "Invalid or malformed JSON"}

	rec, expected := httptest.NewRecorder(), httptest.NewRecorder()
	e.Send(rec)
	Failure(e.StatusCode, e).SendFormat(expected, formatJSON)

	if rec.Body.String() != expected.Body.String() {
		t.Fatalf("unexpected response: %s %s", rec.Body.Bytes(), expected.Body.Bytes())
	} else if rec.Code != e.StatusCode {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}

	// Set fields
	e2 := e
	e2.Field = "field_name"

	rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
	e2.SendWith(SendInput{Field: "field_name"}, rec)

	Failure(e.StatusCode, e2).Send(expected)
	if !bytes.Equal(rec.Body.Bytes(), expected.Body.Bytes()) {
		t.Fatalf("unexpected response: %s", rec.Body.Bytes())
	}

	Formatter = nil

	// Send Format
	rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
	e.SendFormat(rec, formatJSON)
	Failure(e.StatusCode, e).SendFormat(expected, formatJSON)

	if rec.Body.String() != expected.Body.String() {
		t.Fatalf("unexpected response: %s %s", rec.Body.Bytes(), expected.Body.Bytes())
	} else if rec.Code != e.StatusCode {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}

	e2 = e
	e2.Field = "field_name"

	rec, expected = httptest.NewRecorder(), httptest.NewRecorder()
	e2.SendFormat(rec, formatXML)
	Failure(e.StatusCode, e2).SendFormat(expected, formatXML)
	if !bytes.Equal(rec.Body.Bytes(), expected.Body.Bytes()) {
		t.Fatalf("unexpected response: %s %s", rec.Body.Bytes(), expected.Body.Bytes())
	}
}
