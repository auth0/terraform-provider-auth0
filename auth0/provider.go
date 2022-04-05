package auth0

import (
	"context"
	"fmt"
	"os"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	"github.com/auth0/terraform-provider-auth0/version"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AUTH0_DOMAIN", nil),
			},
			"client_id": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("AUTH0_CLIENT_ID", nil),
				RequiredWith:  []string{"client_secret"},
				ConflictsWith: []string{"api_token"},
			},
			"client_secret": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("AUTH0_CLIENT_SECRET", nil),
				RequiredWith:  []string{"client_id"},
				ConflictsWith: []string{"api_token"},
			},
			"api_token": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("AUTH0_API_TOKEN", nil),
				ConflictsWith: []string{"client_id", "client_secret"},
			},
			"debug": {
				Type:     schema.TypeBool,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					v := os.Getenv("AUTH0_DEBUG")
					if v == "" {
						return false, nil
					}
					return v == "1" || v == "true" || v == "on", nil
				},
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"auth0_client":                     newClient(),
			"auth0_global_client":              newGlobalClient(),
			"auth0_client_grant":               newClientGrant(),
			"auth0_connection":                 newConnection(),
			"auth0_custom_domain":              newCustomDomain(),
			"auth0_custom_domain_verification": newCustomDomainVerification(),
			"auth0_resource_server":            newResourceServer(),
			"auth0_rule":                       newRule(),
			"auth0_rule_config":                newRuleConfig(),
			"auth0_hook":                       newHook(),
			"auth0_prompt":                     newPrompt(),
			"auth0_prompt_custom_text":         newPromptCustomText(),
			"auth0_email":                      newEmail(),
			"auth0_email_template":             newEmailTemplate(),
			"auth0_user":                       newUser(),
			"auth0_tenant":                     newTenant(),
			"auth0_role":                       newRole(),
			"auth0_log_stream":                 newLogStream(),
			"auth0_branding":                   newBranding(),
			"auth0_guardian":                   newGuardian(),
			"auth0_organization":               newOrganization(),
			"auth0_action":                     newAction(),
			"auth0_trigger_binding":            newTriggerBinding(),
			"auth0_attack_protection":          newAttackProtection(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"auth0_client":        newDataClient(),
			"auth0_global_client": newDataGlobalClient(),
			"auth0_tenant":        newDataTenant(),
		},
	}

	provider.ConfigureContextFunc = ConfigureProvider(provider.TerraformVersion)

	return provider
}

// ConfigureProvider will configure the *schema.Provider so that *management.Management
// client is stored and passed into the subsequent resources as the meta parameter.
func ConfigureProvider(
	terraformVersion string,
) func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		providerVersion := version.ProviderVersion
		sdkVersion := auth0.Version
		terraformSDKVersion := meta.SDKVersionString()

		userAgent := fmt.Sprintf(
			"Terraform-Provider-Auth0/%s (Go-Auth0-SDK/%s; Terraform-SDK/%s; Terraform/%s)",
			providerVersion,
			sdkVersion,
			terraformSDKVersion,
			terraformVersion,
		)

		domain := data.Get("domain").(string)
		debug := data.Get("debug").(bool)
		clientID := data.Get("client_id").(string)
		clientSecret := data.Get("client_secret").(string)
		apiToken := data.Get("api_token").(string)

		authenticationOption := management.WithStaticToken(apiToken)
		// if api_token is not specified, authenticate with client ID and client secret.
		// This is safe because of the provider schema.
		if apiToken == "" {
			authenticationOption = management.WithClientCredentials(clientID, clientSecret)
		}

		apiClient, err := management.New(domain,
			authenticationOption,
			management.WithDebug(debug),
			management.WithUserAgent(userAgent),
		)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return apiClient, nil
	}
}
