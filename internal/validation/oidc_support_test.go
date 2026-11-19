package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOIDCSupport(t *testing.T) {
	yes, no := true, false

	testCases := []struct {
		name           string
		isAllowed      *bool
		allowAllScopes *bool
		scopes         []string
		expectedError  string
	}{
		{name: "scopes only", isAllowed: &yes, scopes: []string{"profile", "email"}},
		{name: "all scopes", isAllowed: &yes, allowAllScopes: &yes},
		{name: "disabled", isAllowed: &no},
		{name: "disabled with allow_all_scopes false", isAllowed: &no, allowAllScopes: &no},
		{name: "unknown is_allowed", scopes: []string{"phone"}},
		{
			name:          "disabled with scopes",
			isAllowed:     &no,
			scopes:        []string{"profile"},
			expectedError: "cannot be set when `is_allowed` is false",
		},
		{
			name:           "disabled with allow_all_scopes",
			isAllowed:      &no,
			allowAllScopes: &yes,
			expectedError:  "cannot be set when `is_allowed` is false",
		},
		{
			name:           "both scope fields",
			isAllowed:      &yes,
			allowAllScopes: &yes,
			scopes:         []string{"profile"},
			expectedError:  "mutually exclusive",
		},
		{
			name:           "allow_all_scopes false without scopes",
			isAllowed:      &yes,
			allowAllScopes: &no,
			expectedError:  "is not set",
		},
		{name: "neither", isAllowed: &yes, expectedError: "requires either"},
		{name: "openid", isAllowed: &yes, scopes: []string{"openid"}, expectedError: "implied"},
		{name: "unknown scope", isAllowed: &yes, scopes: []string{"foo"}, expectedError: "expected one of"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := ValidateOIDCSupport("oidc_support.0", testCase.isAllowed, testCase.allowAllScopes, testCase.scopes)
			if testCase.expectedError == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, testCase.expectedError)
		})
	}
}
