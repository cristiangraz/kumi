package api

import (
	"encoding/xml"
	"net/http"
)

// XMLFormatter formats an API response and writes it as XML.
func XMLFormatter(r Response, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(r.Status)

	// hide status code for successful responses
	if r.Success {
		r.Status = 0
	}

	return xml.NewEncoder(w).Encode(r)
}
