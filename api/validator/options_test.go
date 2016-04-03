package validator

import (
	"testing"

	"github.com/cristiangraz/kumi/api/formatter"
)

func TestValidatorOptionsValid(t *testing.T) {
	tests := []struct {
		options *Options
		expect  error
	}{
		{
			options: &Options{},
			expect:  errOptionsFormatterRequired,
		},
		{
			options: &Options{
				Formatter: formatter.JSON,
			},
			expect: errOptionsRequestBodyHandlerRequired,
		},
		{
			options: &Options{
				Formatter:           formatter.JSON,
				RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
			},
			expect: errOptionsRequestBodyExceededHandlerRequired,
		},
		{
			options: &Options{
				Formatter:           formatter.JSON,
				RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
				RequestBodyExceeded: errorCollection.Get(RequestBodyRequiredError),
			},
			expect: errOptionsInvalidJSONHandlerRequired,
		},
		{
			options: &Options{
				Formatter:           formatter.JSON,
				RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
				RequestBodyExceeded: errorCollection.Get(RequestBodyRequiredError),
				InvalidJSON:         errorCollection.Get(InvalidJSONError),
			},
			expect: errOptionsBadRequestHandlerRequired,
		},
		{
			options: &Options{
				Formatter:           formatter.JSON,
				RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
				RequestBodyExceeded: errorCollection.Get(RequestBodyRequiredError),
				InvalidJSON:         errorCollection.Get(InvalidJSONError),
				BadRequest:          errorCollection.Get(BadRequestError),
			},
			expect: errOptionsRulesRequired,
		},
		{
			options: &Options{
				Formatter:           formatter.JSON,
				RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
				RequestBodyExceeded: errorCollection.Get(RequestBodyRequiredError),
				InvalidJSON:         errorCollection.Get(InvalidJSONError),
				BadRequest:          errorCollection.Get(BadRequestError),
				Rules:               Rules{"*": []Mapping{}},
			},
			expect: nil,
		},
	}

	for i, tt := range tests {
		if err := tt.options.Valid(); err != tt.expect {
			t.Errorf("TestValidatorOptionsValid (%d): Expected %v, given %v", i, tt.expect, err)
		}
	}
}
