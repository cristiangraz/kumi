package api

import (
	"encoding/xml"
	"io"
	"net/http"
	"strings"
)

type (
	// Response is the response format for responding to all API requests
	Response struct {
		XMLName xml.Name `xml:"response" json:"-"`

		// Success indicates whether or not the response was successful
		Success bool `json:"success" xml:"success"`

		// Holds an exportable/visible status code. Errors only
		Status int `json:"status,omitempty" xml:"status,omitempty"`

		// Holds a text representation of the status code (i.e. not_found for 404)
		// Errors only
		Code string `json:"code,omitempty" xml:"code,omitempty"`

		// Holds contextual information about the response
		// Errors only
		// Note for XML: If errors are present, they'll be rendered in the XMLFormatter
		// with an alias type. This is needed for inconsistencies between with and
		// without context info.
		Errors []Error `json:"errors,omitempty" xml:"_"`

		// Data holds the data specific to the request
		Result interface{} `json:"result,omitempty" xml:"result,omitempty"`

		// Pagination info
		Pagination *Paging `json:"paging,omitempty" xml:"paging,omitempty"`
	}

	// ResponseFormatter allows for formatting how the Response is sent.
	ResponseFormatter interface {
		Send(Response, io.Writer) error
	}

	// Paging holds pagination information for the response
	Paging struct {
		XMLName xml.Name `xml:"paging" json:"-"`
		Count   int      `json:"total_count" xml:"total_count"`
		Limit   int      `json:"limit" xml:"limit"`
		Offset  int      `json:"offset" xml:"offset"`
	}

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
)

var (
	// Formatter holds the ResponseFormatter to use.
	// The JSONFormatter is used by default and configured to display context_info.
	Formatter ResponseFormatter = JSONFormatter{
		UseContextInfo: true,
	}
)

// Success creates a new successful response.
func Success(result interface{}) Response {
	return Response{
		Success: true,
		Status:  http.StatusOK,
		Result:  result,
	}
}

// ErrorResponse returns an error API response.
// statusCode should be >= 400 and <= 599
func ErrorResponse(statusCode int, errors ...Error) Response {
	code := strings.Replace(strings.ToLower(http.StatusText(statusCode)), " ", "_", -1)
	if statusCode == 422 {
		code = "unprocessable_entity"
	}

	return Response{
		Success: false,
		Status:  statusCode,
		Code:    code,
		Errors:  errors,
	}
}

// Paging adds pagination data to the response.
func (r Response) Paging(p Paging) Response {
	r.Pagination = &p
	return r
}

// Send passes the response off to the formatter and writes it.
func (r Response) Send(w io.Writer) error {
	return Formatter.Send(r, w)
}
