package api

import (
	"encoding/xml"
	"net/http"
	"strings"
)


// Response is the response format for responding to all API requests
type Response struct {
	XMLName xml.Name `xml:"response" json:"-"`

	// Success indicates whether or not the response was successful
	Success bool `json:"success" xml:"success"`

	// Holds an exportable/visible status code. Errors only
	Status int `json:"status,omitempty" xml:"status,omitempty"`

	// Holds a text representation of the status code (i.e. not_found for 404)
	// Errors only
	Code string `json:"code,omitempty" xml:"code,omitempty"`

	// Holds errors.
	Errors []Error `json:"errors,omitempty" xml:"errors,omitempty"`

	// Data holds the data specific to the request
	Result interface{} `json:"result,omitempty" xml:"result,omitempty"`

	// Pagination info
	Pagination *Paging `json:"paging,omitempty" xml:"paging,omitempty"`
}

// Sender interface is used by kumi to send an API response to a
// http.ResponseWriter.
type Sender interface {
	func Send(http.ResponseWriter)
}

// Paging holds pagination information for the response
type Paging struct {
	XMLName xml.Name     `xml:"paging" json:"-"`
	Count   int          `json:"total_count" xml:"total_count"`
	Limit   int          `json:"limit" xml:"limit"`
	Offset  int          `json:"offset" xml:"offset"`
	Order   *PagingOrder `json:"order,omitempty" xml:"order,omitempty"`
}

// PagingOrder is the order of the pagination.
type PagingOrder struct {
	XMLName   xml.Name `xml:"order" json:"-"`
	Field     string   `json:"field,omitempty" xml:"field"`
	Direction string   `json:"direction,omitempty" xml:"direction"`
}

// FormatterFn is used to format responses.
type FormatterFn func(r *Response, w http.ResponseWriter) error

// Formatter holds the ResponseFormatter to use.
// You must set a Formatter once before calling Send.
// Otherwise use SendFormat.
var Formatter FormatterFn

// Success creates a new successful response.
func Success(result interface{}) *Response {
	return &Response{
		Success: true,
		Status:  http.StatusOK,
		Result:  result,
	}
}

// ErrorResponse returns an error API response.
// statusCode should be >= 400 and <= 599
func ErrorResponse(statusCode int, errors ...Error) *Response {
	code := strings.Replace(strings.ToLower(http.StatusText(statusCode)), " ", "_", -1)
	if statusCode == 422 {
		code = "unprocessable_entity"
	}

	return &Response{
		Success: false,
		Status:  statusCode,
		Code:    code,
		Errors:  errors,
	}
}

// Paging adds pagination data to the response.
func (r *Response) Paging(p Paging) *Response {
	r.Pagination = &p
	return r
}

// Send passes the response off to the formatter and writes it.
func (r *Response) Send(w http.ResponseWriter) error {
	return Formatter(r, w)
}

// SendFormat sends the response using a given formatter
func (r *Response) SendFormat(w http.ResponseWriter, f FormatterFn) error {
	return f(r, w)
}
