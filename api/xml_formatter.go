package api

import (
	"encoding/xml"
	"io"
	"net/http"
)

type (
	// XMLFormatter formats the response to XML.
	XMLFormatter struct {
		UseContextInfo bool
	}
)

// Send formats an API response and writes it as XML.
// Formats errors using context_info.
func (f XMLFormatter) Send(r Response, w io.Writer) error {
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "application/xml")
		rw.WriteHeader(r.Status)
	}

	// Remove StatusCode for successful responses
	if r.Success == true {
		r.Status = 0

		return xml.NewEncoder(w).Encode(r)
	}

	type alias Response
	if f.UseContextInfo == false {
		a := &struct {
			*alias
			Errors []Error `xml:"errors>error,omitempty"`
		}{
			Errors: r.Errors,
			alias:  (*alias)(&r),
		}
		a.alias.Errors = nil

		return xml.NewEncoder(w).Encode(a)
	}

	for i := range r.Errors {
		r.Errors[i].XMLName.Space = "error"
	}

	a := &struct {
		*alias
		Errors []Error `xml:"context_info>errors>error,omitempty"`
	}{
		Errors: r.Errors,
		alias:  (*alias)(&r),
	}
	a.alias.Errors = nil

	return xml.NewEncoder(w).Encode(a)
}
