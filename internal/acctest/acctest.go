package acctest

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	frameworkError "github.com/auth0/terraform-provider-auth0/internal/framework/error"
	frameworkProvider "github.com/auth0/terraform-provider-auth0/internal/framework/provider"
	internalProvider "github.com/auth0/terraform-provider-auth0/internal/provider"
)

// Test checks to see if http recordings are enabled and runs the tests
// in parallel if they are, otherwise it simply wraps resource.Test.
func Test(t *testing.T, testCase resource.TestCase) {
	if httpRecordingsAreEnabled() {
		httpRecorder := newHTTPRecorder(t)
		testCase.ProtoV6ProviderFactories = testProviderFactoriesWithHTTPRecordings(httpRecorder)
		resource.ParallelTest(t, testCase)

		return
	}

	testCase.ProtoV6ProviderFactories = TestProviderFactories()
	resource.Test(t, testCase)
}

func httpRecordingsAreEnabled() bool {
	httpRecordings := os.Getenv("AUTH0_HTTP_RECORDINGS")
	return httpRecordings == "true" || httpRecordings == "1" || httpRecordings == "on"
}

// TestProviderFactories returns the configured auth0 provider used in testing for the framework.
func TestProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	// Set descriptions to support Markdown syntax for SDK resources,
	// this will be used in document generation.
	schema.DescriptionKind = schema.StringMarkdown
	return map[string]func() (tfprotov6.ProviderServer, error){
		"auth0": func() (tfprotov6.ProviderServer, error) {
			return frameworkProvider.MuxServer(internalProvider.New(), frameworkProvider.New())
		},
	}
}

func testProviderFactoriesWithHTTPRecordings(httpRecorder *recorder.Recorder) map[string]func() (tfprotov6.ProviderServer, error) {
	// Set descriptions to support Markdown syntax for SDK resources,
	// this will be used in document generation.
	schema.DescriptionKind = schema.StringMarkdown
	return map[string]func() (tfprotov6.ProviderServer, error){
		"auth0": func() (tfprotov6.ProviderServer, error) {
			sdkProvider := internalProvider.New()
			sdkProvider.ConfigureContextFunc = configureTestProviderWithHTTPRecordings(httpRecorder)
			fwkProvider := frameworkProvider.New()
			fwkProvider.SetConfigureFunc(configureTestFrameworkProviderWithHTTPRecordings(httpRecorder))
			return frameworkProvider.MuxServer(sdkProvider, fwkProvider)
		},
	}
}

func configureTestProviderWithHTTPRecordings(httpRecorder *recorder.Recorder) schema.ConfigureContextFunc {
	return func(_ context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		domain := data.Get("domain").(string)
		debug := data.Get("debug").(bool)

		clientOptions := []management.Option{
			management.WithStaticToken("insecure"),
			management.WithClient(httpRecorder.GetDefaultClient()),
			management.WithDebug(debug),
			management.WithRetries(3, []int{http.StatusTooManyRequests, http.StatusInternalServerError}),
		}

		if domain != RecordingsDomain {
			clientID := data.Get("client_id").(string)
			clientSecret := data.Get("client_secret").(string)
			apiToken := data.Get("api_token").(string)
			audience := data.Get("audience").(string)

			authenticationOption := management.WithStaticToken(apiToken)
			if apiToken == "" {
				ctx := context.Background()

				authenticationOption = management.WithClientCredentials(ctx, clientID, clientSecret)
				if audience != "" {
					authenticationOption = management.WithClientCredentialsAndAudience(ctx, clientID, clientSecret, audience)
				}
			}

			clientOptions = append(clientOptions, authenticationOption)
		}

		apiClient, err := management.New(domain, clientOptions...)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return config.New(apiClient), nil
	}
}

func configureTestFrameworkProviderWithHTTPRecordings(httpRecorder *recorder.Recorder) func(context.Context, provider.ConfigureRequest, *provider.ConfigureResponse) {
	return func(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
		domain := os.Getenv("AUTH0_DOMAIN")
		debugStr := os.Getenv("AUTH0_DEBUG")
		debug := (debugStr == "1" || debugStr == "true" || debugStr == "TRUE" || debugStr == "on" || debugStr == "ON")

		var data config.FrameworkProviderModel
		response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

		if data.Domain.ValueString() != "" {
			domain = data.Domain.ValueString()
		}
		if !data.Debug.IsNull() && !data.Debug.IsUnknown() {
			debug = data.Debug.ValueBool()
		}

		clientOptions := []management.Option{
			management.WithStaticToken("insecure"),
			management.WithClient(httpRecorder.GetDefaultClient()),
			management.WithDebug(debug),
			management.WithRetries(3, []int{http.StatusTooManyRequests, http.StatusInternalServerError}),
		}

		if domain != RecordingsDomain {
			clientID := os.Getenv("AUTH0_CLIENT_ID")
			clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
			apiToken := os.Getenv("AUTH0_API_TOKEN")
			audience := os.Getenv("AUTH0_AUDIENCE")

			if data.ClientID.ValueString() != "" {
				clientID = data.ClientID.ValueString()
			}
			if data.ClientSecret.ValueString() != "" {
				clientSecret = data.ClientSecret.ValueString()
			}
			if data.APIToken.ValueString() != "" {
				apiToken = data.APIToken.ValueString()
			}
			if data.Audience.ValueString() != "" {
				audience = data.Audience.ValueString()
			}

			authenticationOption := management.WithStaticToken(apiToken)
			if apiToken == "" {
				ctx := context.Background()

				authenticationOption = management.WithClientCredentials(ctx, clientID, clientSecret)
				if audience != "" {
					authenticationOption = management.WithClientCredentialsAndAudience(ctx, clientID, clientSecret, audience)
				}
			}

			clientOptions = append(clientOptions, authenticationOption)
		}

		apiClient, err := management.New(domain, clientOptions...)
		if err != nil {
			response.Diagnostics.Append(frameworkError.Diagnostics(err)...)
		}

		if !response.Diagnostics.HasError() {
			config := config.New(apiClient)
			response.ResourceData = config
			response.DataSourceData = config
		}
	}
}
