package api

import (
	"encoding/xml"
	"fmt"
	"net/http"
)

// Error is the format for each individual API error
type Error struct {
	XMLName xml.Name `xml:"error" json:"-"`

	// StatusCode is optional status code to send with error.
	StatusCode int `json:"-" xml:"-"`

	// Field relates to if the error is parameter-specific. You can use
	// this to display a message near the correct form field, for example.
	Field string `json:"field,omitempty" xml:"field,attr"`

	// Code describes the kind of error that occurred.
	Type string `json:"type" xml:"type,attr"`

	// Message is a human-readable string giving more details about the error.
	Message string `json:"message,omitempty" xml:",innerxml"`
}

// SendInput provides a means to override Error fields
// when sending.
type SendInput struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Send sends the Error with no field. Implements the Sender interface.
func (e Error) Send(w http.ResponseWriter) {
	statusCode := e.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	Failure(statusCode, Error{Field: e.Field, Type: e.Type, Message: e.Message}).Send(w)
}

// SendFormat sends the StatusError with no field.
func (e Error) SendFormat(w http.ResponseWriter, f FormatterFn) {
	statusCode := e.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	Failure(statusCode, Error{Field: e.Field, Type: e.Type, Message: e.Message}).SendFormat(w, f)
}

// With returns a new Error with the given fields.
func (e Error) With(input SendInput) Error {
	if input.Field != "" {
		e.Field = input.Field
	}
	if input.Message != "" {
		e.Message = input.Message
	}

	return e
}

// WithField replaces the Field property of the error.
func (e Error) WithField(field string) Error {
	e.Field = field
	return e
}

// WithMessage replaces the Field property of the error.
func (e Error) WithMessage(msg string) Error {
	e.Message = msg
	return e
}

// SendWith sends the Error with the input params providing overrides.
func (e Error) SendWith(input SendInput, w http.ResponseWriter) {
	e = e.With(input)
	statusCode := e.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	Failure(statusCode, Error{Field: e.Field, Type: e.Type, Message: e.Message}).Send(w)
}
