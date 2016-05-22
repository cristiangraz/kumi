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
				RequestBodyRequired: RequestBodyRequiredError,
			},
			expect: errOptionsRequestBodyExceededHandlerRequired,
		},
		{
			options: &Options{
				RequestBodyRequired: RequestBodyRequiredError,
				RequestBodyExceeded: RequestBodyRequiredError,
			},
			expect: errOptionsInvalidJSONHandlerRequired,
		},
		{
			options: &Options{
				RequestBodyRequired: RequestBodyRequiredError,
				RequestBodyExceeded: RequestBodyRequiredError,
				InvalidJSON:         InvalidJSONError,
			},
			expect: errOptionsBadRequestHandlerRequired,
		},
		{
			options: &Options{
				RequestBodyRequired: RequestBodyRequiredError,
				RequestBodyExceeded: RequestBodyRequiredError,
				InvalidJSON:         InvalidJSONError,
				BadRequest:          BadRequestError,
			},
			expect: errOptionsRulesRequired,
		},
		{
			options: &Options{
				RequestBodyRequired: RequestBodyRequiredError,
				RequestBodyExceeded: RequestBodyRequiredError,
				InvalidJSON:         InvalidJSONError,
				BadRequest:          BadRequestError,
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
