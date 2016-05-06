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

// ErrorCollection maps strings to Errors. This is a
// good place to put standardized API error definitions.
type ErrorCollection map[string]Error

// Errors holds a collection of standardized API error definitions for easy
// error responses.
var Errors ErrorCollection

// GetError is a convenience method to access the ErrorCollection.
func GetError(errType string) Error {
	return Errors.Get(errType)
}

// Error implements the error interface.
func (e Error) Error() string {
	if e.Field == "" {
		return e.Message
	}

	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Get returns the StatusError from the ErrorCollection.
// If none is found an empty StatusError is returned.
func (c ErrorCollection) Get(errType string) Error {
	if se, ok := c[errType]; ok {
		return se
	}

	return Error{}
}

// Send sends the Error with no field.
func (e Error) Send(w http.ResponseWriter) {
	statusCode := e.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	Failure(statusCode, Error{Type: e.Type, Message: e.Message}).Send(w)
}

// SendFormat sends the StatusError with no field.
func (e Error) SendFormat(w http.ResponseWriter, f FormatterFn) {
	statusCode := e.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	Failure(statusCode, Error{Type: e.Type, Message: e.Message}).SendFormat(w, f)
}

// With returns an api.Sender with the given fields.
func (e Error) With(input SendInput) *ErrorResponse {
	se := e
	if input.Field != "" {
		se.Field = input.Field
	}

	if input.Message != "" {
		se.Message = input.Message
	}

	statusCode := se.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusBadRequest
	}

	return Failure(statusCode, Error{
		Field:   se.Field,
		Type:    se.Type,
		Message: se.Message,
	})
}

// SendWith sends the StatusError with the input params providing overrides.
func (e Error) SendWith(input SendInput, w http.ResponseWriter) {
	e.With(input).Send(w)
}
