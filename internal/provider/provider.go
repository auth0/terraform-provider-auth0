package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/action"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/attackprotection"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/branding"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/client"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/connection"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/customdomain"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/guardian"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/hook"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/logstream"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/organization"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/prompt"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/resourceserver"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/role"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/rule"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/tenant"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/user"
)

var version = "dev"

// New returns a *schema.Provider.
func New() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("AUTH0_DOMAIN", nil),
				Description: "Your Auth0 domain name. " +
					"It can also be sourced from the `AUTH0_DOMAIN` environment variable.",
			},
			"audience": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AUTH0_AUDIENCE", nil),
				Description: "Your Auth0 audience when using a custom domain. " +
					"It can also be sourced from the `AUTH0_AUDIENCE` environment variable.",
			},
			"client_id": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("AUTH0_CLIENT_ID", nil),
				RequiredWith:  []string{"client_secret"},
				ConflictsWith: []string{"api_token"},
				Description: "Your Auth0 client ID. " +
					"It can also be sourced from the `AUTH0_CLIENT_ID` environment variable.",
			},
			"client_secret": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("AUTH0_CLIENT_SECRET", nil),
				RequiredWith:  []string{"client_id"},
				ConflictsWith: []string{"api_token"},
				Description: "Your Auth0 client secret. " +
					"It can also be sourced from the `AUTH0_CLIENT_SECRET` environment variable.",
			},
			"api_token": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("AUTH0_API_TOKEN", nil),
				ConflictsWith: []string{"client_id", "client_secret"},
				Description: "Your Auth0 [management api access token]" +
					"(https://auth0.com/docs/security/tokens/access-tokens/management-api-access-tokens). " +
					"It can also be sourced from the `AUTH0_API_TOKEN` environment variable. " +
					"It can be used instead of `client_id` + `client_secret`. " +
					"If both are specified, `api_token` will be used over `client_id` + `client_secret` fields.",
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
				Description: "Indicates whether to turn on debug mode.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"auth0_client":                     client.NewResource(),
			"auth0_global_client":              client.NewGlobalResource(),
			"auth0_client_grant":               client.NewGrantResource(),
			"auth0_connection":                 connection.NewResource(),
			"auth0_connection_client":          connection.NewClientResource(),
			"auth0_custom_domain":              customdomain.NewResource(),
			"auth0_custom_domain_verification": customdomain.NewVerificationResource(),
			"auth0_resource_server":            resourceserver.NewResource(),
			"auth0_rule":                       rule.NewResource(),
			"auth0_rule_config":                rule.NewConfigResource(),
			"auth0_hook":                       hook.NewResource(),
			"auth0_prompt":                     prompt.NewResource(),
			"auth0_prompt_custom_text":         prompt.NewCustomTextResource(),
			"auth0_email":                      newEmail(),
			"auth0_email_template":             newEmailTemplate(),
			"auth0_user":                       user.NewResource(),
			"auth0_tenant":                     tenant.NewResource(),
			"auth0_role":                       role.NewResource(),
			"auth0_log_stream":                 logstream.NewResource(),
			"auth0_branding":                   branding.NewResource(),
			"auth0_guardian":                   guardian.NewResource(),
			"auth0_organization":               organization.NewResource(),
			"auth0_organization_connection":    organization.NewConnectionResource(),
			"auth0_organization_member":        organization.NewMemberResource(),
			"auth0_action":                     action.NewResource(),
			"auth0_trigger_binding":            action.NewTriggerBindingResource(),
			"auth0_attack_protection":          attackprotection.NewResource(),
			"auth0_branding_theme":             branding.NewThemeResource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"auth0_client":            client.NewDataSource(),
			"auth0_global_client":     client.NewGlobalDataSource(),
			"auth0_connection":        connection.NewDataSource(),
			"auth0_organization":      organization.NewDataSource(),
			"auth0_resource_server":   resourceserver.NewDataSource(),
			"auth0_tenant":            tenant.NewDataSource(),
			"auth0_user":              user.NewDataSource(),
			"auth0_role":              role.NewDataSource(),
			"auth0_attack_protection": attackprotection.NewDataSource(),
		},
	}

	provider.ConfigureContextFunc = configureProvider(&provider.TerraformVersion)

	return provider
}

// ConfigureProvider will configure the *schema.Provider so that *management.Management
// client is stored and passed into the subsequent resources as the meta parameter.
func configureProvider(
	terraformVersion *string,
) func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		sdkVersion := auth0.Version
		terraformSDKVersion := meta.SDKVersionString()

		userAgent := fmt.Sprintf(
			"Terraform-Provider-Auth0/%s (Go-Auth0-SDK/%s; Terraform-SDK/%s; Terraform/%s)",
			version,
			sdkVersion,
			terraformSDKVersion,
			*terraformVersion,
		)

		domain := data.Get("domain").(string)
		audience := data.Get("audience").(string)
		debug := data.Get("debug").(bool)
		clientID := data.Get("client_id").(string)
		clientSecret := data.Get("client_secret").(string)
		apiToken := data.Get("api_token").(string)

		authenticationOption := management.WithStaticToken(apiToken)
		// If api_token is not specified, authenticate with client ID and client secret.
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
