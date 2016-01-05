package api

import (
	"encoding/xml"
	"net/http"
)

type (
	// Error is the format for each individual API error
	Error struct {
		XMLName xml.Name `xml:"error" json:"-"`
		// Field relates to if the error is parameter-specific. You can use
		// this to display a message near the correct form field, for example.
		Field string `json:"field,omitempty" xml:"field,attr"`

		// Code describes the kind of error that occurred.
		Type string `json:"type" xml:"type,attr"`

		// Message is a human-readable string giving more details about the error.
		Message string `json:"message,omitempty" xml:",innerxml"`
	}

	// StatusError is an Error with an associated status code.
	StatusError struct {
		Error
		StatusCode int `json:"-" xml:"-"`
	}

	// ErrorCollection maps strings to StatusError errors. This is a
	// good place to put standardized API error definitions.
	ErrorCollection map[string]StatusError
)

// Errors holds a collection of standardized API error definitions for easy
// error responses.
var Errors ErrorCollection

// GetError is a convenience method to access the ErrorCollection.
func GetError(errType string) StatusError {
	return Errors.Get(errType)
}

// Get returns the StatusError from the ErrorCollection.
// If none is found an empty StatusError is returned.
func (c ErrorCollection) Get(errType string) StatusError {
	if se, ok := c[errType]; ok {
		return se
	}

	return StatusError{}
}

// Send sends the StatusError with no field.
func (e StatusError) Send(w http.ResponseWriter) {
	ErrorResponse(e.StatusCode, Error{Type: e.Type, Message: e.Message}).Send(w)
}

// SendFormat sends the StatusError with no field.
func (e StatusError) SendFormat(w http.ResponseWriter, f FormatterFn) {
	ErrorResponse(e.StatusCode, Error{Type: e.Type, Message: e.Message}).SendFormat(w, f)
}

// SendField sends the StatusError with a specific field.
func (e StatusError) SendField(field string, w http.ResponseWriter) {
	ErrorResponse(e.StatusCode, Error{
		Field:   field,
		Type:    e.Type,
		Message: e.Message,
	}).Send(w)
}

// SendFieldFormat sends the StatusError with a specific field.
func (e StatusError) SendFieldFormat(field string, w http.ResponseWriter, f FormatterFn) {
	ErrorResponse(e.StatusCode, Error{
		Field:   field,
		Type:    e.Type,
		Message: e.Message,
	}).SendFormat(w, f)
}
