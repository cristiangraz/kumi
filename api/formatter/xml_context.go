package formatter

import (
	"encoding/xml"
	"net/http"

	"github.com/cristiangraz/kumi/api"
)

// XMLContext formats an API response and writes it as XML.
// The errors are stored in a context_info tag.
func XMLContext(r *api.Response, w http.ResponseWriter) error {
	if r.Success || r.Errors == nil {
		return XML(r, w)
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(r.Status)

	type alias api.Response
	a := &struct {
		*alias
		Errors []api.Error `xml:"context_info>errors>error,omitempty"`
	}{
		Errors: r.Errors,
		alias:  (*alias)(r),
	}
	a.alias.Errors = nil

	return xml.NewEncoder(w).Encode(a)
}
