package auth0

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/wiremock"
)

const wireMockURL = "localhost:8080"

var (
	testProviderFactories = map[string]func() (*schema.Provider, error){
		"auth0": func() (*schema.Provider, error) {
			return Provider(), nil
		},
	}

	testProviderFactoriesWithMockedAPI = map[string]func() (*schema.Provider, error){
		"auth0": func() (*schema.Provider, error) {
			return providerWithMockedAPI(), nil
		},
	}

	wireMockReset sync.Once
)

func providerWithMockedAPI() *schema.Provider {
	provider := Provider()
	provider.ConfigureContextFunc = func(ctx context.Context, _ *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var err error

		wireMockClient := wiremock.NewClient(wireMockURL)
		wireMockReset.Do(func() {
			err = wireMockClient.Reset(ctx)
		})
		if err != nil {
			return nil, diag.FromErr(err)
		}

		apiClient, err := wireMockClient.NewManagementAPIClient()
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return apiClient, nil
	}
	return provider
}

// Auth0 returns an instance of the Management
// API Client used within test sweepers.
func Auth0() (*management.Management, error) {
	domain := os.Getenv("AUTH0_DOMAIN")
	if domain == "" {
		return nil, fmt.Errorf("failed to instantiate api client: AUTH0_DOMAIN is empty")
	}

	apiToken := os.Getenv("AUTH0_API_TOKEN")
	authenticationOption := management.WithStaticToken(apiToken)
	if apiToken == "" {
		authenticationOption = management.WithClientCredentials(
			os.Getenv("AUTH0_CLIENT_ID"),
			os.Getenv("AUTH0_CLIENT_SECRET"),
		)
	}

	return management.New(domain, authenticationOption)
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
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

		p := Provider()
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
					Summary: "Missing required argument",
					Detail:  "The argument \"domain\" is required, but no definition was found.",
				},
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
			provider := Provider()

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
}
