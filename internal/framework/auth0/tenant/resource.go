package tenant

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/go-auth0/management"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/framework/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/framework/schema"
	internalValidation "github.com/auth0/terraform-provider-auth0/internal/frameworkframework/validation"
)

const (
	idleSessionLifetimeDefault = 72.00
	sessionLifetimeDefault     = 168.00
)

type resourceType struct {
	cfg *config.Config
}

type sessionCookieModel struct {
	Mode types.String `tfsdk:"mode"`
}

var sessionCookieTypeMap = map[string]attr.Type {
	"mode": types.StringType,
}

type sessionModel struct {
	OidcLogoutTypeEnabled types.Bool `tfsdk:"oidc_logout_type_enabled"`
}

var sessionTypeMap = map[string]attr.Type {
	"oidc_logout_type_enabled": types.BoolType,
}

type mtlsModel struct {
	EnableEndpointAliases types.Bool `tfsdk:"enable_endpoint_aliases"`
	Disable               types.Bool `tfsdk:"disable"`
}

var mtlsTypeMap = map[string]attr.Type {
	"enable_endpoint_aliases": types.BoolType,
	"disable":                 types.BoolType,
}

type resourceModel struct {
	Id                        types.String  `tfsdk:"id"`
	DefaultAudience           types.String  `tfsdk:"default_audience"`
	DefaultDirectory          types.String  `tfsdk:"default_directory"`
	FriendlyName              types.String  `tfsdk:"friendly_name"`
	PictureUrl                types.String  `tfsdk:"picture_url"`
	SupportEmail              types.String  `tfsdk:"support_email"`
	SupportUrl                types.String  `tfsdk:"support_url"`
	AllowedLogoutUrls         types.List    `tfsdk:"allowed_logout_urls"`
	SandboxVersion            types.String  `tfsdk:"sandbox_version"`
	SessionLifetime           types.Float64 `tfsdk:"session_lifetime"`
	IdleSessionLifetime       types.Float64 `tfsdk:"idle_session_lifetime"`
	EnabledLocales            types.List    `tfsdk:"enabled_locales"`
	Flags                     types.Object  `tfsdk:"flags"`
	DefaultRedirectionUri     types.String  `tfsdk:"default_redirection_uri"`
	SessionCookie             types.Object  `tfsdk:"session_cookie"`
	Sessions                  types.Object  `tfsdk:"sessions"`
	AllowOrgName              types.Bool    `tfsdk:"allow_organization_name_in_authentication_api"`
	CustomizeMfa              types.Bool    `tfsdk:"customize_mfa_in_postlogin_action"`
	AcrValuesSupported        types.Set     `tfsdk:"acr_values_supported"`
	DisableAcrValuesSupported types.Bool    `tfsdk:"disable_acr_values_supported"`
	ParSupported              types.Bool    `tfsdk:"pushed_authorization_requests_supported"`
	Mtls                      types.Object  `tfsdk:"mtls"`
}

// NewResource will return a new auth0_tenant resource.
func NewResource() resource.Resource {
	return &resourceType{}
}

// Configure will be called by the framework to configure the auth0_tenant resource.
func (r *resourceType) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData != nil {
		r.cfg = request.ProviderData.(*config.Config)
	}
}

// Metadata will be called by the framework to get the type name for the auth0_tenant resource.
func (r *resourceType) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	//response.TypeName = "auth0_tenant"
	response.TypeName = "auth0_framework_tenant"
}

// Schema will be called by the framework to get the schema for the auth0_tenant resource.
func (r *resourceType) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	if response != nil {
		response.Schema = schema.Schema{
			Description: "With this resource, you can manage Auth0 tenants, including setting logos and support contact " +
				"information, setting error pages, and configuring default tenant behaviors.",
			Attributes: map[string]schema.Attribute{
				"default_audience": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Description: "API Audience to use by default for API Authorization flows. This setting is " +
						"equivalent to appending the audience to every authorization request made to the tenant " +
						"for every application.",
				},
				"default_directory": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Description: "Name of the connection to be used for Password Grant exchanges. " +
						"Options include auth0-adldap, ad, auth0, email, sms, waad, and adfs.",
					MarkdownDescription: "Name of the connection to be used for Password Grant exchanges. " +
						"Options include `auth0-adldap`, `ad`, `auth0`, `email`, `sms`, `waad`, and `adfs`.",
				},
				"friendly_name": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Friendly name for the tenant.",
				},
				"picture_url": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Description: "URL of logo to be shown for the tenant. Recommended size is 150px x 150px. " +
						"If no URL is provided, the Auth0 logo will be used.",
				},
				"support_email": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Support email address for authenticating users.",
				},
				"support_url": schema.StringAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Support URL for authenticating users.",
				},
				"allowed_logout_urls": schema.ListAttribute{
					ElementType:        types.StringType,
					Optional:    true,
					Computed:    true,
					Description: "URLs that Auth0 may redirect to after logout.",
				},
				"sandbox_version": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Description: "Selected sandbox version for the extensibility environment, which allows you to " +
						"use custom scripts to extend parts of Auth0's functionality.",
				},
				"session_lifetime": schema.Float64_Attribute{
					Optional:    true,
					Default:     sessionLifetimeDefault,
					Validators: []validator.Float64{float64validator.AtLeast(0.01)},
					Description: "Number of hours during which a session will stay valid.",
				},
				"idle_session_lifetime": schema.Float64_Attribute{
					Optional:     true,
					Default:      idleSessionLifetimeDefault,
					Validators: []validator.Float64{float64validator.AtLeast(0.01)},
					Description:  "Number of hours during which a session can be inactive before the user must log in again.",
				},
				"enabled_locales": schema.ListAttribute{
					ElementType:        types.StringType,
					Optional: true,
					Computed: true,
					Description: "Supported locales for the user interface. The first locale in the list will be " +
						"used to set the default locale.",
				},
				"flags": schema.SingleNestedAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Configuration settings for tenant flags.",
					Attributes: map[string]schema.Attribute{
						"enable_client_connections": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether all current connections should be enabled when a new client is created.",
						},
						"enable_apis_section": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether the APIs section is enabled for the tenant.",
						},
						"enable_pipeline2": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether advanced API Authorization scenarios are enabled.",
						},
						"enable_dynamic_client_registration": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether the tenant allows dynamic client registration.",
						},
						"enable_custom_domain_in_emails": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Description: "Indicates whether the tenant allows custom domains in emails. " +
								"Before enabling this flag, you must have a custom domain with status: ready.",
							MarkdownDescription: "Indicates whether the tenant allows custom domains in emails. " +
								"Before enabling this flag, you must have a custom domain with status: `ready`.",
						},
						"enable_sso": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Description: "Flag indicating whether users will not be prompted to confirm log in before SSO redirection. " +
								"This flag applies to existing tenants only; new tenants have it enforced as true.",
						},
						"enable_legacy_logs_search_v2": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether to use the older v2 legacy logs search.",
						},
						"disable_clickjack_protection_headers": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether classic Universal Login prompts include additional security headers to prevent clickjacking.",
						},
						"enable_public_signup_user_exists_error": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether the public sign up process shows a user_exists error if the user already exists.",
							MarkdownDescription: "Indicates whether the public sign up process shows a `user_exists` error if the user already exists.",
						},
						"use_scope_descriptions_for_consent": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether to use scope descriptions for consent.",
						},
						"allow_legacy_delegation_grant_types": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether the legacy delegation endpoint will be enabled for your account (true) or not available (false).",
						},
						"allow_legacy_ro_grant_types": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether the legacy auth/ro endpoint (used with resource owner password and passwordless features) will be enabled for your account (true) or not available (false).",
							MarkdownDescription: "Whether the legacy `auth/ro` endpoint (used with resource owner password and passwordless features) will be enabled for your account (true) or not available (false).",
						},
						"allow_legacy_tokeninfo_endpoint": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "If enabled, customers can use Tokeninfo Endpoint, otherwise they can not use it.",
						},
						"enable_legacy_profile": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether ID tokens and the userinfo endpoint includes a complete user profile (true) or only OpenID Connect claims (false).",
						},
						"enable_idtoken_api2": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether ID tokens can be used to authorize some types of requests to API v2 (true) or not (false).",
						},
						"no_disclose_enterprise_connections": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Do not Publish Enterprise Connections Information with IdP domains on the lock configuration file.",
						},
						"disable_management_api_sms_obfuscation": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "If true, SMS phone numbers will not be obfuscated in Management API GET calls.",
						},
						"enable_adfs_waad_email_verification": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "If enabled, users will be presented with an email verification prompt during their first login when using Azure AD or ADFS connections.",
						},
						"revoke_refresh_token_grant": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Delete underlying grant when a refresh token is revoked via the Authentication API.",
						},
						"dashboard_log_streams_next": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Enables beta access to log streaming changes.",
						},
						"dashboard_insights_view": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Enables new insights activity page view.",
						},
						"disable_fields_map_fix": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Disables SAML fields map fix for bad mappings with repeated attributes.",
						},
						"mfa_show_factor_list_on_enrollment": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Used to allow users to pick which factor to enroll with from the list of available MFA factors.",
						},
						"require_pushed_authorization_requests": schema.BoolAttribute{
							Deprecated:  "This Flag is not supported by the Auth0 Management API and will be removed in the next major release.",
							Optional:    true,
							Computed:    true,
							Description: "This Flag is not supported by the Auth0 Management API and will be removed in the next major release.",
						},
						"remove_alg_from_jwks": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Remove alg from jwks(JSON Web Key Sets).",
							MarkdownDescription: "Remove `alg` from jwks(JSON Web Key Sets).",
						},
					},
				},
				"default_redirection_uri": schema.StringAttribute{
					Optional:     true,
					Computed:     true,
					Validators: []validator.String{internalValidation.IsURLWithHTTPSorEmptyString()},
					Description:  "The default absolute redirection URI. Must be HTTPS or an empty string.",
				},
				"session_cookie": schema.ListNestedAttribute{
					Optional:    true,
					Computed:    true,
					MaxItems:    1,
					Description: "Alters behavior of tenant's session cookie. Contains a single mode property.",
					MarkdownDescription: "Alters behavior of tenant's session cookie. Contains a single `mode` property.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"mode": schema.StringAttribute{
								Optional: true,
								Validators: []validator.String{stringvalidator.OneOf(
									"persistent",
									"non-persistent",
								)},
								Description: "Behavior of tenant session cookie. Accepts either \"persistent\" or \"non-persistent\".",
							},
						},
					},
				},
				"sessions": schema.ListNestedAttribute{
					Optional:    true,
					Computed:    true,
					MaxItems:    1,
					Description: "Sessions related settings for the tenant.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"oidc_logout_prompt_enabled": schema.BoolAttribute{
								Required: true,
								Description: "When active, users will be presented with a consent prompt to confirm the " +
									"logout request if the request is not trustworthy. Turn off the consent prompt to " +
									"bypass user confirmation.",
							},
						},
					},
				},
				"allow_organization_name_in_authentication_api": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Whether to accept an organization name instead of an ID on auth endpoints.",
				},
				"customize_mfa_in_postlogin_action": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Whether to enable flexible factors for MFA in the PostLogin action.",
				},
				"acr_values_supported": schema.SetAttribute{
					Optional:    true,
					Computed:    true,
					Description: "List of supported ACR values.",
					ElementType: types.StringType,
				},
				"disable_acr_values_supported": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Disable list of supported ACR values.",
				},
				"pushed_authorization_requests_supported": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Enable pushed authorization requests.",
				},
				"mtls": schema.SingleNestedAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Configuration for mTLS.",
					Attributes:   map[string]schema.Attribute{
						"enable_endpoint_aliases": schema.BoolAttribute{
							Optional:    true,
							Description: "Enable mTLS endpoint aliases.",
						},
						"disable": schema.BoolAttribute{
							Optional:    true,
							Description: "Disable mTLS settings.",
						},
					},
				},
			},
		}
	}
}

// ImportState will be called by the framework to import an existing auth0_tenant resource.
func (r *resourceType) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// Create will be called by the framework to initialise a new auth0_tenant resource.
func (r *resourceType) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	id, err := uuid.GenerateUUID()
	if err != nil {
		response.Diagnostics.Append(internalError.Diagnostics(err)...)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), id)...)
	response.Diagnostics.Append(updateTenant(ctx, r.cfg.GetAPI(), request.Plan, response.State, &response.State)...)
}

// Update will be called by the framework to update an auth0_tenant resource.
func (r *resourceType) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.Append(updateTenant(ctx, r.cfg.GetAPI(), request.Plan, request.State, &response.State)...)
}

func updateTenant(ctx context.Context, api *management.Management, requestPlan tfsdk.Plan, requestState tfsdk.State, responseState *tfsdk.State) diag.Diagnostics {
	tenant := expandTenant(data)
	if err := api.Tenant.Update(ctx, tenant); err != nil {
		return diag.FromErr(err)
	}
	// These call should NOT be needed, but the tests fail sometimes if it they not there.
	time.Sleep(800 * time.Millisecond)

	if isACRValuesSupportedNull(data) {
		if err := api.Request(ctx, http.MethodPatch, api.URI("tenants", "settings"), map[string]interface{}{
			"acr_values_supported": nil,
		}); err != nil {
			return diag.FromErr(err)
		}
	}
	time.Sleep(200 * time.Millisecond)

	if isMTLSConfigurationNull(data) {
		if err := api.Request(ctx, http.MethodPatch, api.URI("tenants", "settings"), map[string]interface{}{
			"mtls": nil,
		}); err != nil {
			return diag.FromErr(err)
		}
	}
	time.Sleep(800 * time.Millisecond)

	response.Diagnostics.Append(readTenant(ctx, api, &response.State)...)
}

// Read will be called by the framework to read an auth0_tenant resource.
func (r *resourceType) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	api := r.cfg.GetAPI()

	if !response.Diagnostics.HasError() {
		response.Diagnostics.Append(readTenant(ctx, api, &response.State)...)
	}
}

func readTenant(ctx context.Context, api *management.Management, responseState *tfsdk.State) diag.Diagnostics {
	tenant, err := api.Tenant.Read(ctx)
	if err != nil {
		return internalError.HandleAPIError(ctx, responseState, err)
	}

	return diag.FromErr(flattenTenant(data, tenant))
}

// Delete will be called by the framework to delete an auth0_tenant resource.
func (r *resourceType) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Tenants can't be deleted (or created), so do nothing.
}

/*
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
*/

