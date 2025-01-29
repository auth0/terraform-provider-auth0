package provider

import (
	/*
		"os"
		"sort"
		"strings"
	*/
	"testing"

	/*
		"github.com/hashicorp/terraform-plugin-framework/diag"
	*/
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

/*
Func TestProvider(t *testing.T) {
	if err := New().InternalValidate(); err != nil {
		t.Fatal(err)
	}
}

func TestProvider_debugDefaults(t *testing.T) {
	for value, expected := range map[string]bool{
		"1":     true,
		"true":  true,
		"on":    true,
		"0":     false,
		"off":   false,
		"false": false,
		"foo":   false,
		"":      false,
	} {
		_ = os.Unsetenv("AUTH0_DEBUG")
		if value != "" {
			_ = os.Setenv("AUTH0_DEBUG", value)
		}

		p := New()
		debug, err := p.Schema["debug"].DefaultValue()
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if debug.(bool) != expected {
			t.Fatalf("Expected debug to be %v, but got %v", expected, debug)
		}
	}
}

func TestProvider_configValidation(t *testing.T) {
	testCases := []struct {
		name           string
		resourceConfig map[string]interface{}
		expectedErrors diag.Diagnostics
	}{
		{
			name:           "missing client id",
			resourceConfig: map[string]interface{}{"domain": "test", "client_secret": "test"},
			expectedErrors: diag.Diagnostics{
				diag.Diagnostic{
					Summary: "RequiredWith",
					Detail:  "\"client_secret\": all of `client_id,client_secret` must be specified",
				},
			},
		},
		{
			name:           "missing client secret",
			resourceConfig: map[string]interface{}{"domain": "test", "client_id": "test"},
			expectedErrors: diag.Diagnostics{
				diag.Diagnostic{
					Summary: "RequiredWith",
					Detail:  "\"client_id\": all of `client_id,client_secret` must be specified",
				},
			},
		},
		{
			name:           "conflicting auth0 client and management token without domain",
			resourceConfig: map[string]interface{}{"client_id": "test", "client_secret": "test", "api_token": "test"},
			expectedErrors: diag.Diagnostics{
				diag.Diagnostic{
					Summary: "ConflictsWith",
					Detail:  "\"api_token\": conflicts with client_id",
				},
				diag.Diagnostic{
					Summary: "ConflictsWith",
					Detail:  "\"client_id\": conflicts with api_token",
				},
				diag.Diagnostic{
					Summary: "ConflictsWith",
					Detail:  "\"client_secret\": conflicts with api_token",
				},
			},
		},
		{
			name:           "valid auth0 client",
			resourceConfig: map[string]interface{}{"domain": "valid_domain", "client_id": "test", "client_secret": "test"},
			expectedErrors: nil,
		},
		{
			name:           "valid auth0 token",
			resourceConfig: map[string]interface{}{"domain": "valid_domain", "api_token": "test"},
			expectedErrors: nil,
		},
	}

	originalEnvironment := os.Environ()
	os.Clearenv()
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			config := terraform.NewResourceConfigRaw(test.resourceConfig)
			provider := New()

			errs := provider.Validate(config)
			assertErrorsSliceEqual(t, test.expectedErrors, errs)
		})
	}

	for _, e := range originalEnvironment {
		environmentPair := strings.Split(e, "=")
		_ = os.Setenv(environmentPair[0], environmentPair[1])
	}
}

func sortErrors(errs diag.Diagnostics) {
	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Detail < errs[j].Detail
	})
}

func assertErrorsSliceEqual(t *testing.T, expected, actual diag.Diagnostics) {
	if len(expected) != len(actual) {
		t.Fatalf(
			"actual did not match expected. len(expected) != len(actual). expected: %v, actual: %v",
			expected,
			actual,
		)
	}

	sortErrors(expected)
	sortErrors(actual)

	for i := range expected {
		if expected[i].Detail != actual[i].Detail {
			t.Fatalf(
				"actual did not match expected. expected[%d] != actual[%d]. expected: %+v, actual: %+v",
				i,
				i,
				expected,
				actual,
			)
		}
	}
}.
*/
