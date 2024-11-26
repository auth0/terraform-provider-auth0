package provider

import (
	"os"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/flow"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/form"

	selfserviceprofile "github.com/auth0/terraform-provider-auth0/internal/auth0/selfserviceprofile"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/action"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/attackprotection"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/branding"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/client"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/connection"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/customdomain"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/email"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/encryptionkeymanager"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/guardian"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/hook"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/logstream"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/organization"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/page"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/prompt"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/resourceserver"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/role"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/rule"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/signingkey"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/tenant"
	"github.com/auth0/terraform-provider-auth0/internal/auth0/user"
	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// New returns a *schema.Provider.
func New() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AUTH0_DOMAIN", config.MissingEnvVar),
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
				DefaultFunc:   schema.EnvDefaultFunc("AUTH0_CLIENT_ID", config.MissingEnvVar),
				RequiredWith:  []string{"client_secret"},
				ConflictsWith: []string{"api_token"},
				Description: "Your Auth0 client ID. " +
					"It can also be sourced from the `AUTH0_CLIENT_ID` environment variable.",
			},
			"client_secret": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("AUTH0_CLIENT_SECRET", config.MissingEnvVar),
				RequiredWith:  []string{"client_id"},
				ConflictsWith: []string{"api_token"},
				Description: "Your Auth0 client secret. " +
					"It can also be sourced from the `AUTH0_CLIENT_SECRET` environment variable.",
			},
			"api_token": {
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("AUTH0_API_TOKEN", config.MissingEnvVar),
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
			"auth0_action":                           action.NewResource(),
			"auth0_trigger_actions":                  action.NewTriggerActionsResource(),
			"auth0_trigger_action":                   action.NewTriggerActionResource(),
			"auth0_attack_protection":                attackprotection.NewResource(),
			"auth0_branding":                         branding.NewResource(),
			"auth0_branding_theme":                   branding.NewThemeResource(),
			"auth0_client":                           client.NewResource(),
			"auth0_client_credentials":               client.NewCredentialsResource(),
			"auth0_client_grant":                     client.NewGrantResource(),
			"auth0_connection":                       connection.NewResource(),
			"auth0_connection_client":                connection.NewClientResource(),
			"auth0_connection_clients":               connection.NewClientsResource(),
			"auth0_connection_scim_configuration":    connection.NewSCIMConfigurationResource(),
			"auth0_custom_domain":                    customdomain.NewResource(),
			"auth0_custom_domain_verification":       customdomain.NewVerificationResource(),
			"auth0_email_provider":                   email.NewResource(),
			"auth0_email_template":                   email.NewTemplateResource(),
			"auth0_encryption_key_manager":           encryptionkeymanager.NewEncryptionKeyManagerResource(),
			"auth0_flow":                             flow.NewResource(),
			"auth0_flow_vault_connection":            flow.NewVaultConnectionResource(),
			"auth0_form":                             form.NewResource(),
			"auth0_guardian":                         guardian.NewResource(),
			"auth0_hook":                             hook.NewResource(),
			"auth0_log_stream":                       logstream.NewResource(),
			"auth0_organization":                     organization.NewResource(),
			"auth0_organization_client_grant":        organization.NewOrganizationClientGrantResource(),
			"auth0_organization_connection":          organization.NewConnectionResource(),
			"auth0_organization_connections":         organization.NewConnectionsResource(),
			"auth0_organization_member":              organization.NewMemberResource(),
			"auth0_organization_member_role":         organization.NewMemberRoleResource(),
			"auth0_organization_member_roles":        organization.NewMemberRolesResource(),
			"auth0_organization_members":             organization.NewMembersResource(),
			"auth0_pages":                            page.NewResource(),
			"auth0_prompt":                           prompt.NewResource(),
			"auth0_prompt_custom_text":               prompt.NewCustomTextResource(),
			"auth0_prompt_partials":                  prompt.NewPartialsResource(),
			"auth0_prompt_screen_partial":            prompt.NewScreenPartialResource(),
			"auth0_prompt_screen_partials":           prompt.NewScreenPartialsResource(),
			"auth0_resource_server":                  resourceserver.NewResource(),
			"auth0_resource_server_scope":            resourceserver.NewScopeResource(),
			"auth0_resource_server_scopes":           resourceserver.NewScopesResource(),
			"auth0_role":                             role.NewResource(),
			"auth0_role_permission":                  role.NewPermissionResource(),
			"auth0_role_permissions":                 role.NewPermissionsResource(),
			"auth0_rule":                             rule.NewResource(),
			"auth0_rule_config":                      rule.NewConfigResource(),
			"auth0_self_service_profile":             selfserviceprofile.NewResource(),
			"auth0_self_service_profile_custom_text": selfserviceprofile.NewCustomTextResource(),
			"auth0_tenant":                           tenant.NewResource(),
			"auth0_user":                             user.NewResource(),
			"auth0_user_permission":                  user.NewPermissionResource(),
			"auth0_user_permissions":                 user.NewPermissionsResource(),
			"auth0_user_role":                        user.NewRoleResource(),
			"auth0_user_roles":                       user.NewRolesResource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"auth0_attack_protection":             attackprotection.NewDataSource(),
			"auth0_branding":                      branding.NewDataSource(),
			"auth0_branding_theme":                branding.NewThemeDataSource(),
			"auth0_client":                        client.NewDataSource(),
			"auth0_connection":                    connection.NewDataSource(),
			"auth0_connection_scim_configuration": connection.NewSCIMConfigurationDataSource(),
			"auth0_custom_domain":                 customdomain.NewDataSource(),
			"auth0_flow":                          flow.NewDataSource(),
			"auth0_flow_vault_connection":         flow.NewVaultConnectionDataSource(),
			"auth0_form":                          form.NewDataSource(),
			"auth0_organization":                  organization.NewDataSource(),
			"auth0_pages":                         page.NewDataSource(),
			"auth0_prompt_screen_partials":        prompt.NewPromptScreenPartialsDataSource(),
			"auth0_resource_server":               resourceserver.NewDataSource(),
			"auth0_role":                          role.NewDataSource(),
			"auth0_self_service_profile":          selfserviceprofile.NewDataSource(),
			"auth0_signing_keys":                  signingkey.NewDataSource(),
			"auth0_tenant":                        tenant.NewDataSource(),
			"auth0_user":                          user.NewDataSource(),
		},
	}

	provider.ConfigureContextFunc = config.ConfigureProvider(&provider.TerraformVersion)

	return provider
}
