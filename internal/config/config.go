package config

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	"github.com/auth0/terraform-provider-auth0/internal/mutex"
)

const providerName = "Terraform-Provider-Auth0"

var version = "dev"

// Config is the type used for the
// *schema.Provider meta parameter.
type Config struct {
	api   *management.Management
	mutex *mutex.KeyValue
}

// New instantiates a new Config.
func New(apiClient *management.Management) *Config {
	return &Config{
		api:   apiClient,
		mutex: mutex.New(),
	}
}

// GetAPI fetches an instance of the *management.Management client.
func (c *Config) GetAPI() *management.Management {
	return c.api
}

// GetMutex fetches an instance of the *mutex.KeyValue.
func (c *Config) GetMutex() *mutex.KeyValue {
	return c.mutex
}

// ConfigureProvider will configure the *schema.Provider so that
// *management.Management client and *mutex.KeyValue is stored
// and passed into the subsequent resources as the meta parameter.
func ConfigureProvider(terraformVersion *string) schema.ConfigureContextFunc {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		domain := data.Get("domain").(string)
		clientID := data.Get("client_id").(string)
		clientSecret := data.Get("client_secret").(string)
		apiToken := data.Get("api_token").(string)
		audience := data.Get("audience").(string)
		debug := data.Get("debug").(bool)

		apiClient, err := management.New(domain,
			authenticationOption(ctx, clientID, clientSecret, apiToken, audience),
			management.WithDebug(debug),
			management.WithUserAgent(userAgent(terraformVersion)),
			management.WithAuth0ClientEnvEntry(providerName, version),
		)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return New(apiClient), nil
	}
}

// userAgent computes the desired User-Agent header for the *management.Management client.
func userAgent(terraformVersion *string) string {
	sdkVersion := auth0.Version
	terraformSDKVersion := meta.SDKVersionString()

	userAgent := fmt.Sprintf(
		"%s/%s (Go-Auth0-SDK/%s; Terraform-SDK/%s; Terraform/%s)",
		providerName,
		version,
		sdkVersion,
		terraformSDKVersion,
		*terraformVersion,
	)

	return userAgent
}

// authenticationOption computes the desired authentication option for the *management.Management client.
func authenticationOption(ctx context.Context, clientID, clientSecret, apiToken, audience string) management.Option {
	if apiToken != "" {
		return management.WithStaticToken(apiToken)
	}

	if audience != "" {
		return management.WithClientCredentialsAndAudience(
			ctx,
			clientID,
			clientSecret,
			audience,
		)
	}

	return management.WithClientCredentials(ctx, clientID, clientSecret)
}
