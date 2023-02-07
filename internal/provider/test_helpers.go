package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
)

// TestFactories returns the configured Auth0 provider to be used within tests.
func TestFactories(httpRecorder *recorder.Recorder) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"auth0": func() (*schema.Provider, error) {
			auth0Provider := New()

			auth0Provider.ConfigureContextFunc = configureTestProvider(httpRecorder)

			return auth0Provider, nil
		},
	}
}

func configureTestProvider(
	httpRecorder *recorder.Recorder,
) func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		domain := data.Get("domain").(string)
		debug := data.Get("debug").(bool)

		testClient := http.DefaultClient
		if httpRecorder != nil {
			testClient = httpRecorder.GetDefaultClient()
		}

		apiClient, err := management.New(
			domain,
			management.WithStaticToken("insecure"),
			management.WithClient(testClient),
			management.WithDebug(debug),
		)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		if domain != recorder.RecordingsDomain {
			clientID := data.Get("client_id").(string)
			clientSecret := data.Get("client_secret").(string)
			apiToken := data.Get("api_token").(string)
			audience := data.Get("audience").(string)

			authenticationOption := management.WithStaticToken(apiToken)
			if apiToken == "" {
				authenticationOption = management.WithClientCredentials(clientID, clientSecret)

				if audience != "" {
					authenticationOption = management.WithClientCredentialsAndAudience(
						clientID,
						clientSecret,
						audience,
					)
				}
			}

			apiClient, err = management.New(
				domain,
				authenticationOption,
				management.WithClient(testClient),
				management.WithDebug(debug),
			)
			if err != nil {
				return nil, diag.FromErr(err)
			}
		}

		return apiClient, nil
	}
}
