package tenant

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/go-auth0/management"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/commons"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalValidation "github.com/auth0/terraform-provider-auth0/internal/validation"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

const (
	idleSessionLifetimeDefault = 72.00
	sessionLifetimeDefault     = 168.00
)

// NewResource will return a new auth0_tenant resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createTenant,
		ReadContext:   readTenant,
		UpdateContext: updateTenant,
		DeleteContext: deleteTenant,
		CustomizeDiff: validateTenant,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage Auth0 tenants, including setting logos and support contact " +
			"information, setting error pages, and configuring default tenant behaviors.",
		Schema: map[string]*schema.Schema{
			"default_audience": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "API Audience to use by default for API Authorization flows. This setting is " +
					"equivalent to appending the audience to every authorization request made to the tenant " +
					"for every application.",
			},
			"default_directory": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Name of the connection to be used for Password Grant exchanges. " +
					"Options include `auth0-adldap`, `ad`, `auth0`, `email`, `sms`, `waad`, and `adfs`.",
			},
			"friendly_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Friendly name for the tenant.",
			},
			"picture_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "URL of logo to be shown for the tenant. Recommended size is 150px x 150px. " +
					"If no URL is provided, the Auth0 logo will be used.",
			},
			"support_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Support email address for authenticating users.",
			},
			"support_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Support URL for authenticating users.",
			},
			"allowed_logout_urls": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
				Description: "URLs that Auth0 may redirect to after logout.",
			},
			"sandbox_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Selected sandbox version for the extensibility environment, which allows you to " +
					"use custom scripts to extend parts of Auth0's functionality.",
			},
			"session_lifetime": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Default:      sessionLifetimeDefault,
				ValidateFunc: validation.FloatAtLeast(0.01),
				Description:  "Number of hours during which a session will stay valid.",
			},
			"idle_session_lifetime": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Default:      idleSessionLifetimeDefault,
				ValidateFunc: validation.FloatAtLeast(0.01),
				Description:  "Number of hours during which a session can be inactive before the user must log in again.",
			},
			"enabled_locales": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				Description: "Supported locales for the user interface. The first locale in the list will be " +
					"used to set the default locale.",
			},
			"flags": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration settings for tenant flags.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_client_connections": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether all current connections should be enabled when a new client is created.",
						},
						"enable_apis_section": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether the APIs section is enabled for the tenant.",
						},
						"enable_pipeline2": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether advanced API Authorization scenarios are enabled.",
						},
						"enable_dynamic_client_registration": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether the tenant allows dynamic client registration.",
						},
						"enable_custom_domain_in_emails": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							Description: "Indicates whether the tenant allows custom domains in emails. " +
								"Before enabling this flag, you must have a custom domain with status: `ready`.",
						},
						"enable_sso": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							Description: "Flag indicating whether users will not be prompted to confirm log in before SSO redirection. " +
								"This flag applies to existing tenants only; new tenants have it enforced as true.",
						},
						"enable_legacy_logs_search_v2": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether to use the older v2 legacy logs search.",
						},
						"disable_clickjack_protection_headers": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether classic Universal Login prompts include additional security headers to prevent clickjacking.",
						},
						"enable_public_signup_user_exists_error": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether the public sign up process shows a `user_exists` error if the user already exists.",
						},
						"use_scope_descriptions_for_consent": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether to use scope descriptions for consent.",
						},
						"allow_legacy_delegation_grant_types": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Whether the legacy delegation endpoint will be enabled for your account (true) or not available (false).",
						},
						"allow_legacy_ro_grant_types": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Whether the legacy `auth/ro` endpoint (used with resource owner password and passwordless features) will be enabled for your account (true) or not available (false).",
						},
						"allow_legacy_tokeninfo_endpoint": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "If enabled, customers can use Tokeninfo Endpoint, otherwise they can not use it.",
						},
						"enable_legacy_profile": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Whether ID tokens and the userinfo endpoint includes a complete user profile (true) or only OpenID Connect claims (false).",
						},
						"enable_idtoken_api2": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Whether ID tokens can be used to authorize some types of requests to API v2 (true) or not (false).",
						},
						"no_disclose_enterprise_connections": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Do not Publish Enterprise Connections Information with IdP domains on the lock configuration file.",
						},
						"disable_management_api_sms_obfuscation": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "If true, SMS phone numbers will not be obfuscated in Management API GET calls.",
						},
						"enable_adfs_waad_email_verification": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "If enabled, users will be presented with an email verification prompt during their first login when using Azure AD or ADFS connections.",
						},
						"revoke_refresh_token_grant": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Delete underlying grant when a refresh token is revoked via the Authentication API.",
						},
						"dashboard_log_streams_next": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Enables beta access to log streaming changes.",
						},
						"dashboard_insights_view": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Enables new insights activity page view.",
						},
						"disable_fields_map_fix": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Disables SAML fields map fix for bad mappings with repeated attributes.",
						},
						"mfa_show_factor_list_on_enrollment": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Used to allow users to pick which factor to enroll with from the list of available MFA factors.",
						},
						"require_pushed_authorization_requests": {
							Deprecated:  "This Flag is not supported by the Auth0 Management API and will be removed in the next major release.",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "This Flag is not supported by the Auth0 Management API and will be removed in the next major release.",
						},
						"remove_alg_from_jwks": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Remove `alg` from jwks(JSON Web Key Sets).",
						},
					},
				},
			},
			"default_redirection_uri": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
				Description:  "The default absolute redirection URI. Must be HTTPS or an empty string.",
			},
			"session_cookie": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Alters behavior of tenant's session cookie. Contains a single `mode` property.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"persistent",
								"non-persistent",
							}, false),
							Description: "Behavior of tenant session cookie. Accepts either \"persistent\" or \"non-persistent\".",
						},
					},
				},
			},
			"sessions": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Sessions related settings for the tenant.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"oidc_logout_prompt_enabled": {
							Type:     schema.TypeBool,
							Required: true,
							Description: "When active, users will be presented with a consent prompt to confirm the " +
								"logout request if the request is not trustworthy. Turn off the consent prompt to " +
								"bypass user confirmation.",
						},
					},
				},
			},
			"oidc_logout": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Settings related to OIDC RP-initiated Logout.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rp_logout_end_session_endpoint_discovery": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Enable the end_session_endpoint URL in the .well-known discovery configuration.",
						},
					},
				},
			},
			"allow_organization_name_in_authentication_api": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether to accept an organization name instead of an ID on auth endpoints.",
			},
			"customize_mfa_in_postlogin_action": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether to enable flexible factors for MFA in the PostLogin action.",
			},
			"acr_values_supported": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "List of supported ACR values.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"disable_acr_values_supported": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Disable list of supported ACR values.",
			},
			"pushed_authorization_requests_supported": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable pushed authorization requests.",
			},
			"mtls": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration for mTLS.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_endpoint_aliases": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable mTLS endpoint aliases.",
						},
						"disable": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Disable mTLS settings.",
						},
					},
				},
			},
			"error_page": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration for the error page",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"html": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Custom Error HTML (Liquid syntax is supported)",
						},
						"show_log_link": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether to show the link to log as part of the default error page (true, default) or not to show the link (false).",
						},
						"url": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "URL to redirect to when an error occurs instead of showing the default error page",
						},
					},
				},
			},
			"default_token_quota": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Token Quota configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"clients":       commons.TokenQuotaSchema(),
						"organizations": commons.TokenQuotaSchema(),
					},
				},
			},
			"skip_non_verifiable_callback_uri_confirmation_prompt": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Indicates whether to skip the confirmation prompt when using non-verifiable callback URIs. Accepts 'true', 'false', or 'null'.",
				ValidateFunc: validation.StringInSlice([]string{"true", "false", "null"}, false),
				DiffSuppressFunc: func(_, o, n string, _ *schema.ResourceData) bool {
					return (o == "null" && n == "") || o == n
				},
			},
		},
	}
}

func createTenant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())
	return updateTenant(ctx, data, meta)
}

func readTenant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	tenant, err := api.Tenant.Read(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenTenant(data, tenant))
}

func validateTenant(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	var result *multierror.Error
	disableACRValues := diff.GetRawConfig().GetAttr("disable_acr_values_supported")
	if !disableACRValues.IsNull() && disableACRValues.True() {
		acrValues := diff.GetRawConfig().GetAttr("acr_values_supported")
		if !acrValues.IsNull() && acrValues.LengthInt() > 0 {
			result = multierror.Append(
				result,
				fmt.Errorf("only one of disable_acr_values_supported and acr_values_supported should be set"),
			)
		}
	}

	mtlsConfig := diff.GetRawConfig().GetAttr("mtls")
	if !mtlsConfig.IsNull() {
		var disable, enableEndpointAliases *bool

		mtlsConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
			disable = value.Bool(cfg.GetAttr("disable"))
			enableEndpointAliases = value.Bool(cfg.GetAttr("enable_endpoint_aliases"))
			return stop
		})
		if disable != nil && *disable && enableEndpointAliases != nil && *enableEndpointAliases {
			result = multierror.Append(
				result,
				fmt.Errorf("only one of disable and enable_endpoint_aliases should be set in the mtls block"),
			)
		}
	}

	return result.ErrorOrNil()
}

func updateTenant(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	tenant := expandTenant(data)
	if err := api.Tenant.Update(ctx, tenant); err != nil {
		return diag.FromErr(err)
	}
	// These call should NOT be needed, but the tests fail sometimes if it they not there.
	time.Sleep(800 * time.Millisecond)

	nullFields := fetchNullableFields(data)
	if len(nullFields) != 0 {
		body, _ := json.Marshal(nullFields)
		if err := api.Tenant.Update(context.Background(), nil, management.Body(body)); err != nil {
			return diag.FromErr(err)
		}
	}

	return readTenant(ctx, data, meta)
}

func deleteTenant(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
