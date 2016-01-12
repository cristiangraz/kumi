package formatter

import (
	"encoding/json"
	"net/http"

	"github.com/cristiangraz/kumi/api"
)

// JSON formats an API response and writes it as JSON.
func JSON(r *api.Response, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)

	// hide status code for successful responses
	if r.Success {
		r.Status = 0
	}

	return json.NewEncoder(w).Encode(r)
}
