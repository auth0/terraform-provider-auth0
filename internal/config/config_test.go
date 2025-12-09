package config_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/provider"

	"github.com/auth0/terraform-provider-auth0/internal/mutex"
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
			name: "it can configure a provider with client credentials and private key JWT",
			givenTerraformConfig: map[string]interface{}{
				"domain":                       "example.auth0.com",
				"client_id":                    "1234567",
				"client_assertion_private_key": "private-key",
				"client_assertion_signing_alg": "RS256",
			},
			expectedDiagnostics: nil,
		},
		{
			name: "it can configure a provider with client credentials, audience, and private key JWT",
			givenTerraformConfig: map[string]interface{}{
				"domain":                       "example.auth0.com",
				"client_id":                    "1234567",
				"client_assertion_private_key": "private-key",
				"client_assertion_signing_alg": "RS256",
				"audience":                     "myaudience",
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
						"or AUTH0_CLIENT_ID and AUTH0_CLIENT_SECRET, " +
						"or AUTH0_CLIENT_ID, AUTH0_CLIENT_ASSERTION_PRIVATE_KEY, and AUTH0_CLIENT_ASSERTION_SIGNING_ALG, " +
						"or enable CLI login with AUTH0_CLI_LOGIN=true.",
				},
			},
		},
		{
			name: "it can configure a provider with custom_domain_header",
			givenTerraformConfig: map[string]interface{}{
				"domain":               "example.auth0.com",
				"client_id":            "1234567",
				"client_secret":        "secret",
				"custom_domain_header": "demo-sdk.acmetest.org",
			},
			expectedDiagnostics: nil,
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

func TestParseResourceConfigData(t *testing.T) {
	jwtToken := createJwtToken(t)
	f := createCliConfigFile(t, jwtToken)
	config.SetCliConfigPath(f.Name())
	defer func(name string) {
		_ = os.Remove(name)
	}(f.Name())

	var testCases = []struct {
		name                 string
		givenTerraformConfig map[string]interface{}
		keyringAccessToken   string
		configAccessToken    string
		expectedDiagnostics  diag.Diagnostics
		expectedConfig       config.ProviderConfig
	}{
		{
			name: "it parses all auth configuration options",
			givenTerraformConfig: map[string]interface{}{
				"domain":                       "example.auth0.com",
				"debug":                        true,
				"client_id":                    "1234567",
				"client_secret":                "secret",
				"api_token":                    "api-token",
				"audience":                     "the-audience",
				"client_assertion_private_key": "private-key",
				"client_assertion_signing_alg": "signing-alg",
				"custom_domain_header":         "custom-domain",
			},
			expectedDiagnostics: nil,
			expectedConfig: config.ProviderConfig{
				Domain:                    "example.auth0.com",
				Debug:                     true,
				ClientID:                  "1234567",
				ClientSecret:              "secret",
				APIToken:                  "api-token",
				Audience:                  "the-audience",
				ClientAssertionPrivateKey: "private-key",
				ClientAssertionSigningAlg: "signing-alg",
				CustomDomainHeader:        "custom-domain",
			},
		},
		{
			name: "it returns an error when dynamic_credentials is set and domain is empty",
			givenTerraformConfig: map[string]interface{}{
				"domain":              "",
				"dynamic_credentials": true,
			},
			expectedDiagnostics: diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Missing required configuration",
				Detail:   "The 'AUTH0_DOMAIN' must be specified along with 'AUTH0_DYNAMIC_CREDENTIALS'.",
			}},
			expectedConfig: config.ProviderConfig{},
		},
		{
			name: "it returns an error when dynamic_credentials is set and domain is empty",
			givenTerraformConfig: map[string]interface{}{
				"domain":    "",
				"api_token": true,
			},
			expectedDiagnostics: diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Missing required configuration",
				Detail:   "The 'AUTH0_DOMAIN' must be specified along with 'AUTH0_API_TOKEN'.",
			}},
			expectedConfig: config.ProviderConfig{},
		},
		{
			name: "it returns an error when client_id with client_secret is set and domain is empty ",
			givenTerraformConfig: map[string]interface{}{
				"domain":        "",
				"client_id":     "client-id",
				"client_secret": "secret",
			},
			expectedDiagnostics: diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Missing required configuration",
				Detail:   "The 'AUTH0_DOMAIN' must be specified along with 'AUTH0_CLIENT_ID' and 'AUTH0_CLIENT_SECRET'.",
			}},
			expectedConfig: config.ProviderConfig{},
		},
		{
			name: "it returns an error when client_id, client_assertion_private_key and client_assertion_signing_alg is set and domain is empty ",
			givenTerraformConfig: map[string]interface{}{
				"domain":                       "",
				"client_id":                    "client-id",
				"client_assertion_private_key": "key",
				"client_assertion_signing_alg": "alg",
			},
			expectedDiagnostics: diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Missing required configuration",
				Detail:   "The 'AUTH0_DOMAIN' must be specified along with 'AUTH0_CLIENT_ID', 'AUTH0_CLIENT_ASSERTION_PRIVATE_KEY', and 'AUTH0_CLIENT_ASSERTION_SIGNING_ALG'.",
			}},
			expectedConfig: config.ProviderConfig{},
		},
		{
			name: "it returns an error when cli_login is set and domain is empty",
			givenTerraformConfig: map[string]interface{}{
				"domain":    "",
				"cli_login": true,
			},
			expectedDiagnostics: diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Missing required configuration",
				Detail:   "The 'AUTH0_DOMAIN' must be specified along with 'AUTH0_CLI_LOGIN'.",
			}},
			expectedConfig: config.ProviderConfig{},
		},
		{
			name: "it loads access token from keyring when cli_login is set",
			givenTerraformConfig: map[string]interface{}{
				"domain":    "example-token.auth0.com",
				"cli_login": true,
			},
			keyringAccessToken:  jwtToken,
			expectedDiagnostics: nil,
			expectedConfig: config.ProviderConfig{
				Domain:   "example-token.auth0.com",
				APIToken: jwtToken,
			},
		},
		{
			name: "it loads access token from config file when cli_login is set and no access token is found in keyring",
			givenTerraformConfig: map[string]interface{}{
				"domain":    "example-token.auth0.com",
				"cli_login": true,
			},
			keyringAccessToken:  "",
			configAccessToken:   jwtToken,
			expectedDiagnostics: nil,
			expectedConfig: config.ProviderConfig{
				Domain:   "example-token.auth0.com",
				APIToken: jwtToken,
			},
		},
		{
			name: "it return an error when cli_login is set and no access token is found in keyring and config file",
			givenTerraformConfig: map[string]interface{}{
				"domain":    "example.auth0.com",
				"cli_login": true,
			},
			expectedDiagnostics: diag.Diagnostics{{
				Severity: diag.Error,
				Summary:  "Authentication required",
				Detail:   "No CLI token found. Please log in using 'auth0 login' via auth0-cli or disable 'AUTH0_CLI_LOGIN'.",
			}},
			expectedConfig: config.ProviderConfig{},
		},
		{
			name:                 "it returns an error when it can't initialize the api client",
			givenTerraformConfig: map[string]interface{}{},
			expectedDiagnostics: diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Missing environment variables",
					Detail: "AUTH0_DOMAIN is required. Then, configure either AUTH0_API_TOKEN, " +
						"or AUTH0_CLIENT_ID and AUTH0_CLIENT_SECRET, " +
						"or AUTH0_CLIENT_ID, AUTH0_CLIENT_ASSERTION_PRIVATE_KEY, and AUTH0_CLIENT_ASSERTION_SIGNING_ALG, " +
						"or enable CLI login with AUTH0_CLI_LOGIN=true.",
				},
			},
			expectedConfig: config.ProviderConfig{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			keyring.MockInit()
			if testCase.keyringAccessToken != "" {
				_ = keyring.Set("Auth0 CLI Access Token 0", "example-token.auth0.com", testCase.keyringAccessToken)
			}

			resourceData := schema.TestResourceDataRaw(t, provider.New().Schema, testCase.givenTerraformConfig)
			cfg, diags := config.ParseResourceConfigData(resourceData)

			if testCase.expectedDiagnostics != nil {
				assert.Equal(t, diags, testCase.expectedDiagnostics)
				return
			}

			assert.Nil(t, diags)
			assert.Equal(t, testCase.expectedConfig, cfg)
		})
	}
}

func createJwtToken(t *testing.T) string {
	token, err := jwt.NewBuilder().
		Expiration(time.Now().Add(5 * time.Minute)).
		Build()
	assert.NoError(t, err)

	signed, err := jwt.Sign(token, jwt.WithInsecureNoSignature())
	assert.NoError(t, err)

	jwtToken := string(signed)
	return jwtToken
}

func createCliConfigFile(t *testing.T, token string) *os.File {
	f, err := os.CreateTemp("", "auth0-terraform-provider-test")
	assert.NoError(t, err)

	json := `
{
  "tenants": {
    "example-token.auth0.com": {
      "name": "example-token",
      "domain": "example-token.auth0.com",
      "scopes": [],
      "expires_at": "2025-10-09T19:06:30.45293+02:00",
      "access_token": "` + token + `"
    }
  }
}
`

	_, err = f.Write([]byte(json))
	assert.NoError(t, err)

	return f
}
