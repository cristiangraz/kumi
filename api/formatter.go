package api

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

// FormatterFn is used to format responses.
type FormatterFn func(r *Response, w http.ResponseWriter) error

// Formatter holds the ResponseFormatter to use.
// You must set a Formatter once before calling Send.
// Otherwise use SendFormat.
var Formatter FormatterFn = JSON

// JSON formats an API response and writes it as JSON.
func JSON(r *Response, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)

	// hide status code for successful responses
	if r.Success {
		r.Status = 0
	}
	return json.NewEncoder(w).Encode(r)
}

// XML formats an API response and writes it as XML.
func XML(r *Response, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(r.Status)

	// hide status code for successful responses
	if r.Success {
		r.Status = 0
	}
	if r.Success || len(r.Errors) == 0 {
		return xml.NewEncoder(w).Encode(r)
	}

	type alias Response
	a := struct {
		*alias
		Errors []Error `xml:"errors>error,omitempty"`
	}{
		Errors: r.Errors,
		alias:  (*alias)(r),
	}
	return xml.NewEncoder(w).Encode(a)
}
