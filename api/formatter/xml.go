package formatter

import (
	"encoding/xml"
	"net/http"

	"github.com/cristiangraz/kumi/api"
)

// XML formats an API response and writes it as XML.
func XML(r api.Response, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(r.Status)

	// hide status code for successful responses
	if r.Success {
		r.Status = 0
	}

	if r.Success || len(r.Errors) == 0 {
		return xml.NewEncoder(w).Encode(r)
	}

	type alias api.Response
	a := struct {
		*alias
		Errors []api.Error `json:"errors,omitempty" xml:"errors>error,omitempty"`
	}{
		Errors: r.Errors,
		alias:  (*alias)(&r),
	}

	return xml.NewEncoder(w).Encode(a)
}
