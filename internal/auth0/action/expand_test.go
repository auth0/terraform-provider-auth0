package action

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
)

func TestCheckForUntrackedActionSecrets(t *testing.T) {
	var testCases = []struct {
		name                 string
		givenSecretsInConfig []interface{}
		givenActionSecrets   []management.ActionSecret
		expectedDiagnostics  diag.Diagnostics
	}{
		{
			name:                 "action has no secrets",
			givenSecretsInConfig: []interface{}{},
			givenActionSecrets:   []management.ActionSecret{},
			expectedDiagnostics:  diag.Diagnostics(nil),
		},
		{
			name: "action has no untracked secrets",
			givenSecretsInConfig: []interface{}{
				map[string]interface{}{
					"name": "secretName",
				},
			},
			givenActionSecrets: []management.ActionSecret{
				{
					Name: auth0.String("secretName"),
				},
			},
			expectedDiagnostics: diag.Diagnostics(nil),
		},
		{
			name: "action has untracked secrets",
			givenSecretsInConfig: []interface{}{
				map[string]interface{}{
					"name": "secretName",
				},
			},
			givenActionSecrets: []management.ActionSecret{
				{
					Name: auth0.String("secretName"),
				},
				{
					Name: auth0.String("anotherSecretName"),
				},
			},
			expectedDiagnostics: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Unmanaged Action Secret",
					Detail: "Detected an action secret not managed though Terraform: anotherSecretName. " +
						"If you proceed, this secret will get deleted. It is required to add this secret to " +
						"your action configuration to prevent unintentionally destructive results.",
					AttributePath: cty.Path{cty.GetAttrStep{Name: "secrets"}},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualDiagnostics := checkForUnmanagedActionSecrets(
				testCase.givenSecretsInConfig,
				testCase.givenActionSecrets,
			)

			assert.Equal(t, testCase.expectedDiagnostics, actualDiagnostics)
		})
	}
}

// TestCheckForUntrackedActionSecretsWithSecretsWO verifies that the guard accepts
// secrets_wo entries (which are shaped identically to secrets entries as map[string]interface{})
// and does not flag API-side secrets whose names appear in the secrets_wo list.
func TestCheckForUntrackedActionSecretsWithSecretsWO(t *testing.T) {
	var testCases = []struct {
		name                 string
		givenSecretsInConfig []interface{}
		givenActionSecrets   []management.ActionSecret
		expectedDiagnostics  diag.Diagnostics
	}{
		{
			name: "secrets_wo entry covers API secret with same name",
			givenSecretsInConfig: []interface{}{
				map[string]interface{}{
					"name":  "apiKey",
					"value": "s3cret",
				},
			},
			givenActionSecrets: []management.ActionSecret{
				{Name: auth0.String("apiKey")},
			},
			expectedDiagnostics: diag.Diagnostics(nil),
		},
		{
			name: "API has secret not present in secrets_wo",
			givenSecretsInConfig: []interface{}{
				map[string]interface{}{
					"name":  "apiKey",
					"value": "s3cret",
				},
			},
			givenActionSecrets: []management.ActionSecret{
				{Name: auth0.String("apiKey")},
				{Name: auth0.String("unmanagedKey")},
			},
			expectedDiagnostics: diag.Diagnostics{
				{
					Severity: diag.Error,
					Summary:  "Unmanaged Action Secret",
					Detail: "Detected an action secret not managed though Terraform: unmanagedKey. " +
						"If you proceed, this secret will get deleted. It is required to add this secret to " +
						"your action configuration to prevent unintentionally destructive results.",
					AttributePath: cty.Path{cty.GetAttrStep{Name: "secrets"}},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualDiagnostics := checkForUnmanagedActionSecrets(
				testCase.givenSecretsInConfig,
				testCase.givenActionSecrets,
			)

			assert.Equal(t, testCase.expectedDiagnostics, actualDiagnostics)
		})
	}
}
