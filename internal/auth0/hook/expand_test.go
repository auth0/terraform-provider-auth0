package hook

import (
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
)

func TestHookNameRegexp(t *testing.T) {
	for givenHookName, expectedError := range map[string]bool{
		"my-hook-1":                 false,
		"hook 2 name with spaces":   false,
		" hook with a space prefix": true,
		"hook with a space suffix ": true,
		" ":                         true,
		"   ":                       true,
	} {
		validationResult := validateHookName()(givenHookName, cty.Path{cty.GetAttrStep{Name: "name"}})
		assert.Equal(t, expectedError, validationResult.HasError())
	}
}

func TestCheckForUntrackedHookSecrets(t *testing.T) {
	var testCases = []struct {
		name                string
		givenConfigSecrets  map[string]interface{}
		givenHookSecrets    management.HookSecrets
		expectedDiagnostics diag.Diagnostics
	}{
		{
			name:                "hook has no secrets",
			givenConfigSecrets:  map[string]interface{}{},
			givenHookSecrets:    management.HookSecrets{},
			expectedDiagnostics: diag.Diagnostics(nil),
		},
		{
			name:                "hook has no untracked secrets",
			givenConfigSecrets:  map[string]interface{}{"key": "value"},
			givenHookSecrets:    management.HookSecrets{"key": "value"},
			expectedDiagnostics: diag.Diagnostics(nil),
		},
		{
			name:               "hook has untracked secrets",
			givenConfigSecrets: map[string]interface{}{"key1": "value1"},
			givenHookSecrets:   management.HookSecrets{"key1": "value1", "key2": "value2"},
			expectedDiagnostics: diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Unexpected Hook Secrets",
					Detail: "Found unexpected hook secrets with key: key2. To prevent issues, manage them through " +
						"terraform. If you've just imported this resource (and your secrets match), to make this " +
						"warning disappear, run a terraform apply.",
					AttributePath: cty.Path{cty.GetAttrStep{Name: "secrets"}},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualDiagnostics := checkForUntrackedHookSecrets(
				testCase.givenHookSecrets,
				testCase.givenConfigSecrets,
			)

			assert.Equal(t, testCase.expectedDiagnostics, actualDiagnostics)
		})
	}
}
