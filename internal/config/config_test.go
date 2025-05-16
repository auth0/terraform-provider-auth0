package config_test

import (
	"context"
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/mutex"
	"github.com/auth0/terraform-provider-auth0/internal/provider"
)

func TestConfigureProvider(t *testing.T) {
	var testCases = []struct {
		name                 string
		givenTerraformConfig map[string]interface{}
		expectedDiagnostics  diag.Diagnostics
	}{
		{
			name: "it can configure a provider with client credentials",
			givenTerraformConfig: map[string]interface{}{
				"domain":        "example.auth0.com",
				"client_id":     "1234567",
				"client_secret": "secret",
			},
			expectedDiagnostics: nil,
		},
		{
			name: "it can configure a provider with client credentials and audience",
			givenTerraformConfig: map[string]interface{}{
				"domain":        "example.auth0.com",
				"client_id":     "1234567",
				"client_secret": "secret",
				"audience":      "myaudience",
			},
			expectedDiagnostics: nil,
		},
		{
			name: "it can configure a provider with an api token",
			givenTerraformConfig: map[string]interface{}{
				"domain":    "example.auth0.com",
				"api_token": "123456",
			},
			expectedDiagnostics: nil,
		},
		{
			name:                 "it returns an error when it can't initialize the api client",
			givenTerraformConfig: map[string]interface{}{},
			expectedDiagnostics: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Missing environment variables",
					Detail: "AUTH0_DOMAIN is required. Then, configure either AUTH0_API_TOKEN, " +
						"or both AUTH0_CLIENT_ID and AUTH0_CLIENT_SECRET. Or enable CLI login with AUTH0_CLI_LOGIN=true",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			resourceData := schema.TestResourceDataRaw(t, provider.New().Schema, testCase.givenTerraformConfig)

			configureFunc := config.ConfigureProvider(auth0.String("v1.14.0"))
			cfg, diags := configureFunc(context.Background(), resourceData)

			if testCase.expectedDiagnostics != nil {
				assert.Equal(t, diags, testCase.expectedDiagnostics)
				return
			}

			assert.Nil(t, diags)
			assert.IsType(t, &config.Config{}, cfg)
			assert.NotNil(t, cfg.(*config.Config).GetAPI())
			assert.IsType(t, &management.Management{}, cfg.(*config.Config).GetAPI())
			assert.NotNil(t, cfg.(*config.Config).GetMutex())
			assert.IsType(t, &mutex.KeyValue{}, cfg.(*config.Config).GetMutex())
		})
	}
}
