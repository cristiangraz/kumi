package formatter

import (
	"bytes"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/cristiangraz/kumi/api"
)

func TestFormatters(t *testing.T) {
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
		formatter  api.FormatterFn
		statusCode int
		paging     api.Paging
		errors     []api.Error
		want       []byte
	}{
		{
			formatter: JSON,
			want:      []byte(`{"success":true,"result":{"first_name":"Jon","last_name":"Doe","age":30}}`),
		},
		{
			formatter: XML,
			want:      []byte(`<response><success>true</success><result><first_name>Jon</first_name><last_name>Doe</last_name><age>30</age></result></response>`),
		},
		{
			formatter:  JSON,
			statusCode: 409,
			errors: []api.Error{
				api.Error{Field: "email", Type: "already_exists", Message: "A user with that email address already exists"},
			},
			want: []byte(`{"success":false,"status":409,"code":"conflict","errors":[{"field":"email","type":"already_exists","message":"A user with that email address already exists"}]}`),
		},
		{
			formatter:  XML,
			statusCode: 409,
			errors: []api.Error{
				api.Error{Field: "email", Type: "already_exists", Message: "A user with that email address already exists"},
			},
			want: []byte(`<response><success>false</success><status>409</status><code>conflict</code><errors><error field="email" type="already_exists">A user with that email address already exists</error></errors></response>`),
		},
		{
			formatter: JSON,
			paging:    api.Paging{Count: 1, Offset: 0, Limit: 20},
			want:      []byte(`{"success":true,"result":{"first_name":"Jon","last_name":"Doe","age":30},"paging":{"total_count":1,"limit":20,"offset":0}}`),
		},
		{
			formatter: XML,
			paging:    api.Paging{Count: 1, Offset: 0, Limit: 20},
			want:      []byte(`<response><success>true</success><result><first_name>Jon</first_name><last_name>Doe</last_name><age>30</age></result><paging><total_count>1</total_count><limit>20</limit><offset>0</offset></paging></response>`),
		},
		{
			formatter:  JSON,
			statusCode: 422,
			errors: []api.Error{
				api.Error{Field: "email", Type: "required", Message: "Required field missing"},
			},
			want: []byte(`{"success":false,"status":422,"code":"unprocessable_entity","errors":[{"field":"email","type":"required","message":"Required field missing"}]}`),
		},
		{
			formatter:  JSONContext,
			statusCode: 422,
			errors: []api.Error{
				api.Error{Field: "email", Type: "required", Message: "Required field missing"},
			},
			want: []byte(`{"success":false,"status":422,"code":"unprocessable_entity","context_info":{"errors":[{"field":"email","type":"required","message":"Required field missing"}]}}`),
		},
		{
			formatter:  XMLContext,
			statusCode: 409,
			errors: []api.Error{
				api.Error{Field: "email", Type: "already_exists", Message: "A user with that email address already exists"},
			},
			want: []byte(`<response><success>false</success><status>409</status><code>conflict</code><context_info><errors><error field="email" type="already_exists">A user with that email address already exists</error></errors></context_info></response>`),
		},
	}

	for i, tt := range tests {
		api.Formatter = tt.formatter
		given := httptest.NewRecorder()

		var response api.Response
		if len(tt.errors) == 0 {
			response = api.Success(result)
		} else {
			response = api.ErrorResponse(tt.statusCode, tt.errors...)
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

		var response api.Response
		if len(tt.errors) == 0 {
			response = api.Success(result)
		} else {
			response = api.ErrorResponse(tt.statusCode, tt.errors...)
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
