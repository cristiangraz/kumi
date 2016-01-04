package formatter

import (
	"encoding/json"
	"net/http"

	"github.com/cristiangraz/kumi/api"
)

// JSONContext formats an API response and writes it as JSON.
// The errors are stored in a context_info object.
func JSONContext(r api.Response, w http.ResponseWriter) error {
	if r.Success || r.Errors == nil {
		return JSON(r, w)
	}

	type alias api.Response
	a := &struct {
		*alias
		Context map[string][]api.Error `json:"context_info"`
	}{
		Context: map[string][]api.Error{"errors": r.Errors},
		alias:   (*alias)(&r),
	}
	a.alias.Errors = nil

	return json.NewEncoder(w).Encode(a)
}
