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
