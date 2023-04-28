package acctest

import (
	"context"
	"os"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/auth0/terraform-provider-auth0/internal/provider"
)

// Test checks to see if http recordings are enabled and runs the tests
// in parallel if they are, otherwise it simply wraps resource.Test.
func Test(t *testing.T, testCase resource.TestCase) {
	if httpRecordingsAreEnabled() {
		httpRecorder := newHTTPRecorder(t)
		testCase.ProviderFactories = testFactoriesWithHTTPRecordings(httpRecorder)
		resource.ParallelTest(t, testCase)

		return
	}

	testCase.ProviderFactories = TestFactories()
	resource.Test(t, testCase)
}

func httpRecordingsAreEnabled() bool {
	httpRecordings := os.Getenv("AUTH0_HTTP_RECORDINGS")
	return httpRecordings == "true" || httpRecordings == "1" || httpRecordings == "on"
}

// TestFactories returns the configured auth0 provider used in testing.
func TestFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"auth0": func() (*schema.Provider, error) {
			return provider.New(), nil
		},
	}
}

func testFactoriesWithHTTPRecordings(httpRecorder *recorder.Recorder) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"auth0": func() (*schema.Provider, error) {
			auth0Provider := provider.New()

			auth0Provider.ConfigureContextFunc = configureTestProviderWithHTTPRecordings(httpRecorder)

			return auth0Provider, nil
		},
	}
}

func configureTestProviderWithHTTPRecordings(
	httpRecorder *recorder.Recorder,
) func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		domain := data.Get("domain").(string)
		debug := data.Get("debug").(bool)

		clientOptions := []management.Option{
			management.WithStaticToken("insecure"),
			management.WithClient(httpRecorder.GetDefaultClient()),
			management.WithDebug(debug),
		}

		if domain != RecordingsDomain {
			clientID := data.Get("client_id").(string)
			clientSecret := data.Get("client_secret").(string)
			apiToken := data.Get("api_token").(string)
			audience := data.Get("audience").(string)

			authenticationOption := management.WithStaticToken(apiToken)
			if apiToken == "" {
				authenticationOption = management.WithClientCredentials(clientID, clientSecret)
				if audience != "" {
					authenticationOption = management.WithClientCredentialsAndAudience(clientID, clientSecret, audience)
				}
			}

			clientOptions = append(clientOptions, authenticationOption)
		}

		apiClient, err := management.New(domain, clientOptions...)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return apiClient, nil
	}
}
