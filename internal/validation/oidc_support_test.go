package validation

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
)

func TestValidateOIDCSupport(t *testing.T) {
	nullBool := cty.NullVal(cty.Bool)
	nullScopes := cty.NullVal(cty.Set(cty.String))
	noScopes := cty.SetValEmpty(cty.String)
	profile := cty.SetVal([]cty.Value{cty.StringVal("profile")})

	block := func(isAllowed, allowAllScopes, scopes cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"is_allowed":       isAllowed,
			"allow_all_scopes": allowAllScopes,
			"allowed_scopes":   scopes,
		})
	}

	testCases := []struct {
		name          string
		block         cty.Value
		expectedError string
	}{
		{name: "absent block", block: cty.NullVal(cty.DynamicPseudoType)},
		{name: "unknown block", block: cty.UnknownVal(cty.DynamicPseudoType)},
		{name: "scopes only", block: block(cty.True, nullBool, profile)},
		{name: "all scopes", block: block(cty.True, cty.True, nullScopes)},
		{name: "all scopes with empty scopes", block: block(cty.True, cty.True, noScopes)},
		{name: "scopes with allow_all_scopes false", block: block(cty.True, cty.False, profile)},
		{name: "openid only", block: block(cty.True, nullBool, nullScopes)},
		{name: "openid only with empty scopes", block: block(cty.True, nullBool, noScopes)},
		{name: "openid only with allow_all_scopes false", block: block(cty.True, cty.False, noScopes)},
		{name: "disabled", block: block(cty.False, nullBool, nullScopes)},
		{
			name:          "disabled with empty scopes",
			block:         block(cty.False, nullBool, noScopes),
			expectedError: "cannot be set when `is_allowed` is false",
		},
		{
			name:          "disabled with allow_all_scopes false",
			block:         block(cty.False, cty.False, nullScopes),
			expectedError: "cannot be set when `is_allowed` is false",
		},
		{
			name:          "disabled with empty values",
			block:         block(cty.False, cty.False, noScopes),
			expectedError: "cannot be set when `is_allowed` is false",
		},
		{name: "unknown is_allowed", block: block(cty.UnknownVal(cty.Bool), cty.True, profile)},
		{name: "unknown allow_all_scopes", block: block(cty.True, cty.UnknownVal(cty.Bool), profile)},
		{name: "unknown scopes", block: block(cty.False, nullBool, cty.UnknownVal(cty.Set(cty.String)))},
		{
			name:  "unknown scope element",
			block: block(cty.True, cty.True, cty.SetVal([]cty.Value{cty.UnknownVal(cty.String)})),
		},
		{
			name:          "disabled with scopes",
			block:         block(cty.False, nullBool, profile),
			expectedError: "cannot be set when `is_allowed` is false",
		},
		{
			name:          "disabled with allow_all_scopes",
			block:         block(cty.False, cty.True, nullScopes),
			expectedError: "cannot be set when `is_allowed` is false",
		},
		{
			name:          "both scope fields",
			block:         block(cty.True, cty.True, profile),
			expectedError: "mutually exclusive",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := ValidateOIDCSupport("oidc_support", testCase.block)
			if testCase.expectedError == "" {
				assert.NoError(t, err)
				return
			}
			assert.ErrorContains(t, err, testCase.expectedError)
		})
	}
}
