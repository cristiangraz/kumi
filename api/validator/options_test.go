package validator

import "testing"

func TestValidatorOptionsValid(t *testing.T) {
	tests := []struct {
		options *Options
		expect  error
	}{
		{
			options: &Options{},
			expect:  errOptionsRequestBodyHandlerRequired,
		},
		{
			options: &Options{
				RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
			},
			expect: errOptionsRequestBodyExceededHandlerRequired,
		},
		{
			options: &Options{
				RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
				RequestBodyExceeded: errorCollection.Get(RequestBodyRequiredError),
			},
			expect: errOptionsInvalidJSONHandlerRequired,
		},
		{
			options: &Options{
				RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
				RequestBodyExceeded: errorCollection.Get(RequestBodyRequiredError),
				InvalidJSON:         errorCollection.Get(InvalidJSONError),
			},
			expect: errOptionsBadRequestHandlerRequired,
		},
		{
			options: &Options{
				RequestBodyRequired: errorCollection.Get(RequestBodyRequiredError),
				RequestBodyExceeded: errorCollection.Get(RequestBodyRequiredError),
				InvalidJSON:         errorCollection.Get(InvalidJSONError),
				BadRequest:          errorCollection.Get(BadRequestError),
			},
			expect: errOptionsRulesRequired,
		},
		{
			options: &Options{
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
