package api

import (
	"encoding/json"
	"net/http"
)

// JSONFormatter formats an API response and writes it as JSON.
func JSONFormatter(r Response, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)

	// hide status code for successful responses
	if r.Success == true {
		r.Status = 0
	}

	return json.NewEncoder(w).Encode(r)
}
