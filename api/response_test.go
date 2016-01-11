package api

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// formatJSON formats an API response and writes it as JSON.
func formatJSON(r Response, w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(r.Status)

	// hide status code for successful responses
	if r.Success {
		r.Status = 0
	}

	return json.NewEncoder(w).Encode(r)
}

// formatXML formats an API response and writes it as XML.
func formatXML(r Response, w http.ResponseWriter) error {
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
		Errors []Error `json:"errors,omitempty" xml:"errors>error,omitempty"`
	}{
		Errors: r.Errors,
		alias:  (*alias)(&r),
	}

	return xml.NewEncoder(w).Encode(a)
}

func TestResponse(t *testing.T) {
	result := struct {
		FirstName string `json:"first_name,omitempty" xml:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty" xml:"last_name,omitempty"`
		Age       int    `json:"age,omitempty" xml:"age,omitempty"`
	}{
		FirstName: "Jon",
		LastName:  "Doe",
		Age:       30,
	}

	tests := []struct {
		formatter  FormatterFn
		statusCode int
		paging     Paging
		errors     []Error
		want       []byte
	}{
		{
			formatter: formatJSON,
			want:      []byte(`{"success":true,"result":{"first_name":"Jon","last_name":"Doe","age":30}}`),
		},
		{
			formatter: formatXML,
			want:      []byte(`<response><success>true</success><result><first_name>Jon</first_name><last_name>Doe</last_name><age>30</age></result></response>`),
		},
		{
			formatter:  formatJSON,
			statusCode: 409,
			errors: []Error{
				Error{Field: "email", Type: "already_exists", Message: "A user with that email address already exists"},
			},
			want: []byte(`{"success":false,"status":409,"code":"conflict","errors":[{"field":"email","type":"already_exists","message":"A user with that email address already exists"}]}`),
		},
		{
			formatter:  formatXML,
			statusCode: 409,
			errors: []Error{
				Error{Field: "email", Type: "already_exists", Message: "A user with that email address already exists"},
			},
			want: []byte(`<response><success>false</success><status>409</status><code>conflict</code><errors><error field="email" type="already_exists">A user with that email address already exists</error></errors></response>`),
		},
		{
			formatter: formatJSON,
			paging:    Paging{Count: 1, Offset: 0, Limit: 20},
			want:      []byte(`{"success":true,"result":{"first_name":"Jon","last_name":"Doe","age":30},"paging":{"total_count":1,"limit":20,"offset":0}}`),
		},
		{
			formatter: formatJSON,
			paging:    Paging{Count: 1, Offset: 0, Limit: 20, Order: &PagingOrder{Field: "id", Direction: "asc"}},
			want:      []byte(`{"success":true,"result":{"first_name":"Jon","last_name":"Doe","age":30},"paging":{"total_count":1,"limit":20,"offset":0,"order":{"field":"id","direction":"asc"}}}`),
		},
		{
			formatter: formatXML,
			paging:    Paging{Count: 1, Offset: 0, Limit: 20},
			want:      []byte(`<response><success>true</success><result><first_name>Jon</first_name><last_name>Doe</last_name><age>30</age></result><paging><total_count>1</total_count><limit>20</limit><offset>0</offset></paging></response>`),
		},
		{
			formatter: formatXML,
			paging:    Paging{Count: 1, Offset: 0, Limit: 20, Order: &PagingOrder{Field: "id", Direction: "asc"}},
			want:      []byte(`<response><success>true</success><result><first_name>Jon</first_name><last_name>Doe</last_name><age>30</age></result><paging><total_count>1</total_count><limit>20</limit><offset>0</offset><order><field>id</field><direction>asc</direction></order></paging></response>`),
		},
		{
			formatter:  formatJSON,
			statusCode: 422,
			errors: []Error{
				Error{Field: "email", Type: "required", Message: "Required field missing"},
			},
			want: []byte(`{"success":false,"status":422,"code":"unprocessable_entity","errors":[{"field":"email","type":"required","message":"Required field missing"}]}`),
		},
	}

	for i, tt := range tests {
		Formatter = tt.formatter
		given := httptest.NewRecorder()

		var response Response
		if len(tt.errors) == 0 {
			response = Success(result)
		} else {
			response = ErrorResponse(tt.statusCode, tt.errors...)
		}

		if tt.paging.Count > 0 || tt.paging.Limit > 0 || tt.paging.Offset > 0 {
			response = response.Paging(tt.paging)
		}

		response.Send(given)

		if !reflect.DeepEqual(tt.want, bytes.TrimSpace(given.Body.Bytes())) {
			t.Errorf("TestResponse (%d): Want %s, given %s", i, tt.want, given.Body)
		}
	}

	// Test SendFormat
	for i, tt := range tests {
		given := httptest.NewRecorder()

		var response Response
		if len(tt.errors) == 0 {
			response = Success(result)
		} else {
			response = ErrorResponse(tt.statusCode, tt.errors...)
		}

		if tt.paging.Count > 0 || tt.paging.Limit > 0 || tt.paging.Offset > 0 {
			response = response.Paging(tt.paging)
		}

		response.SendFormat(given, tt.formatter)

		if !reflect.DeepEqual(tt.want, bytes.TrimSpace(given.Body.Bytes())) {
			t.Errorf("TestResponse (%d): Invalid format. Want %s, given %s", i, tt.want, given.Body)
		}
	}
}
