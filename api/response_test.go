package api

import (
	"bytes"
	"net/http/httptest"
	"reflect"
	"testing"
)

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
			errors: []Error{
				Error{Field: "email", Type: "already_exists", Message: "A user with that email address already exists"},
			},
			want: []byte(`{"success":false,"status":409,"code":"conflict","errors":[{"field":"email","type":"already_exists","message":"A user with that email address already exists"}]}`),
		},
		{
			formatter:  XML,
			statusCode: 409,
			errors: []Error{
				Error{Field: "email", Type: "already_exists", Message: "A user with that email address already exists"},
			},
			want: []byte(`<response><success>false</success><status>409</status><code>conflict</code><errors><error field="email" type="already_exists">A user with that email address already exists</error></errors></response>`),
		},
		{
			formatter: JSON,
			paging:    Paging{Count: 1, Offset: 0, Limit: 20},
			want:      []byte(`{"success":true,"result":{"first_name":"Jon","last_name":"Doe","age":30},"paging":{"total_count":1,"limit":20,"offset":0}}`),
		},
		{
			formatter: JSON,
			paging:    Paging{Count: 1, Offset: 0, Limit: 20, Order: &PagingOrder{Field: "id", Direction: "asc"}},
			want:      []byte(`{"success":true,"result":{"first_name":"Jon","last_name":"Doe","age":30},"paging":{"total_count":1,"limit":20,"offset":0,"order":{"field":"id","direction":"asc"}}}`),
		},
		{
			formatter: XML,
			paging:    Paging{Count: 1, Offset: 0, Limit: 20},
			want:      []byte(`<response><success>true</success><result><first_name>Jon</first_name><last_name>Doe</last_name><age>30</age></result><paging><total_count>1</total_count><limit>20</limit><offset>0</offset></paging></response>`),
		},
		{
			formatter: XML,
			paging:    Paging{Count: 1, Offset: 0, Limit: 20, Order: &PagingOrder{Field: "id", Direction: "asc"}},
			want:      []byte(`<response><success>true</success><result><first_name>Jon</first_name><last_name>Doe</last_name><age>30</age></result><paging><total_count>1</total_count><limit>20</limit><offset>0</offset><order><field>id</field><direction>asc</direction></order></paging></response>`),
		},
		{
			formatter:  JSON,
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

		if len(tt.errors) == 0 {
			response := Success(result)
			if tt.paging.Count > 0 || tt.paging.Limit > 0 || tt.paging.Offset > 0 {
				response.Paging(tt.paging)
			}

			response.Send(given)
		} else {
			Failure(tt.statusCode, tt.errors...).Send(given)
		}

		if !reflect.DeepEqual(tt.want, bytes.TrimSpace(given.Body.Bytes())) {
			t.Errorf("TestResponse (%d): Want %s, given %s", i, tt.want, given.Body)
		}
	}

	// Test SendFormat
	for i, tt := range tests {
		given := httptest.NewRecorder()

		if len(tt.errors) == 0 {
			response := Success(result)
			if tt.paging.Count > 0 || tt.paging.Limit > 0 || tt.paging.Offset > 0 {
				response.Paging(tt.paging)
			}

			response.SendFormat(given, tt.formatter)
		} else {
			Failure(tt.statusCode, tt.errors...).SendFormat(given, tt.formatter)
		}

		if !reflect.DeepEqual(tt.want, bytes.TrimSpace(given.Body.Bytes())) {
			t.Errorf("TestResponse (%d): Invalid format. Want %s, given %s", i, tt.want, given.Body)
		}
	}
}
