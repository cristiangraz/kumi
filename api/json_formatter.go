package api

import (
	"encoding/json"
	"io"
	"net/http"
)

type (
	// JSONFormatter is the default formatter. It formats the response to JSON.
	JSONFormatter struct {
		UseContextInfo bool
	}
)

// Send formats an API response and writes it as JSON.
// Formats errors using context_info.
func (f JSONFormatter) Send(r Response, w io.Writer) error {
	if rw, ok := w.(http.ResponseWriter); ok {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(r.Status)
	}

	// Hide StatusCode for successful responses
	if r.Success == true {
		r.Status = 0

		return json.NewEncoder(w).Encode(r)
	}

	if f.UseContextInfo == false {
		return json.NewEncoder(w).Encode(r)
	}

	type alias Response
	a := &struct {
		*alias
		Context map[string][]Error `json:"context_info,omitempty"`
	}{
		Context: map[string][]Error{"errors": r.Errors},
		alias:   (*alias)(&r),
	}
	a.alias.Errors = nil

	return json.NewEncoder(w).Encode(a)
}
