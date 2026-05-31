package auth0client

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"
	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

var (
	_ resource.Resource                = (*clientResource)(nil)
	_ resource.ResourceWithConfigure   = (*clientResource)(nil)
	_ resource.ResourceWithImportState = (*clientResource)(nil)
)

// NewResource returns a fresh client resource implementation.
func NewResource() resource.Resource { return &clientResource{} }

type clientResource struct {
	mgmt *mgmtclient.Management
}

// model mirrors the auth0_client HCL schema. Every API-returned field has a
// corresponding attribute so the state file is a faithful projection of the
// resource.
type model struct {
	// Identity / read-only metadata.
	ID           types.String `tfsdk:"id"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Tenant       types.String `tfsdk:"tenant"`
	Global       types.Bool   `tfsdk:"global"`

	// Basic descriptors.
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AppType     types.String `tfsdk:"app_type"`
	LogoURI     types.String `tfsdk:"logo_uri"`

	// Behaviour flags.
	IsFirstParty                                   types.Bool `tfsdk:"is_first_party"`
	OIDCConformant                                 types.Bool `tfsdk:"oidc_conformant"`
	SSO                                            types.Bool `tfsdk:"sso"`
	SSODisabled                                    types.Bool `tfsdk:"sso_disabled"`
	CrossOriginAuthentication                      types.Bool `tfsdk:"cross_origin_authentication"`
	CustomLoginPageOn                              types.Bool `tfsdk:"custom_login_page_on"`
	IsTokenEndpointIPHeaderTrusted                 types.Bool `tfsdk:"is_token_endpoint_ip_header_trusted"`
	RequirePushedAuthorizationRequests             types.Bool `tfsdk:"require_pushed_authorization_requests"`
	RequireProofOfPossession                       types.Bool `tfsdk:"require_proof_of_possession"`
	SkipNonVerifiableCallbackURIConfirmationPrompt types.Bool `tfsdk:"skip_non_verifiable_callback_uri_confirmation_prompt"`

	// URL collections.
	Callbacks         types.List `tfsdk:"callbacks"`
	AllowedLogoutURLs types.List `tfsdk:"allowed_logout_urls"`
	AllowedOrigins    types.List `tfsdk:"allowed_origins"`
	WebOrigins        types.List `tfsdk:"web_origins"`
	ClientAliases     types.List `tfsdk:"client_aliases"`
	AllowedClients    types.List `tfsdk:"allowed_clients"`

	// OAuth / OIDC behaviour.
	GrantTypes              types.List   `tfsdk:"grant_types"`
	TokenEndpointAuthMethod types.String `tfsdk:"token_endpoint_auth_method"`
	CrossOriginLoc          types.String `tfsdk:"cross_origin_loc"`
	InitiateLoginURI        types.String `tfsdk:"initiate_login_uri"`
	FormTemplate            types.String `tfsdk:"form_template"`
	CustomLoginPage         types.String `tfsdk:"custom_login_page"`
	CustomLoginPagePreview  types.String `tfsdk:"custom_login_page_preview"`
	ParRequestExpiry        types.Int64  `tfsdk:"par_request_expiry"`
	ComplianceLevel         types.String `tfsdk:"compliance_level"`
	ThirdPartySecurityMode  types.String `tfsdk:"third_party_security_mode"`
	RedirectionPolicy       types.String `tfsdk:"redirection_policy"`
	JwksURI                 types.String `tfsdk:"jwks_uri"`

	// Organization integration.
	OrganizationUsage            types.String `tfsdk:"organization_usage"`
	OrganizationRequireBehavior  types.String `tfsdk:"organization_require_behavior"`
	OrganizationDiscoveryMethods types.List   `tfsdk:"organization_discovery_methods"`

	// External / CIMD identity.
	ResourceServerIdentifier  types.String `tfsdk:"resource_server_identifier"`
	ExternalMetadataType      types.String `tfsdk:"external_metadata_type"`
	ExternalMetadataCreatedBy types.String `tfsdk:"external_metadata_created_by"`
	ExternalClientID          types.String `tfsdk:"external_client_id"`

	// Free-form metadata.
	ClientMetadata types.Map `tfsdk:"client_metadata"`

	// Nested objects covered in this wave.
	SigningKeys types.List   `tfsdk:"signing_keys"`
	TokenQuota  types.Object `tfsdk:"token_quota"`

	// Wave 2 nested objects.
	JwtConfiguration            types.Object `tfsdk:"jwt_configuration"`
	RefreshToken                types.Object `tfsdk:"refresh_token"`
	OidcLogout                  types.Object `tfsdk:"oidc_logout"`
	OidcBackchannelLogout       types.Object `tfsdk:"oidc_backchannel_logout"` // deprecated alias
	EncryptionKey               types.Object `tfsdk:"encryption_key"`
	DefaultOrganization         types.Object `tfsdk:"default_organization"`
	NativeSocialLogin           types.Object `tfsdk:"native_social_login"`
	SessionTransfer             types.Object `tfsdk:"session_transfer"`
	Mobile                      types.Object `tfsdk:"mobile"`
	TokenExchange               types.Object `tfsdk:"token_exchange"`
	MyOrganizationConfiguration types.Object `tfsdk:"my_organization_configuration"`
	ExpressConfiguration        types.Object `tfsdk:"express_configuration"`

	// Wave 2 — list-of-enum scalars and JSON-string passthroughs.
	AsyncApprovalNotificationChannels types.List   `tfsdk:"async_approval_notification_channels"`
	SignedRequestObject               types.String `tfsdk:"signed_request_object"`
	Addons                            types.Object `tfsdk:"addons"`
	ClientAuthenticationMethods       types.String `tfsdk:"client_authentication_methods"`
}

// -- shared attr type maps for nested objects ------------------------------

func signingKeyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"pkcs7":   types.StringType,
		"cert":    types.StringType,
		"subject": types.StringType,
	}
}

func tokenQuotaCCAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enforce":  types.BoolType,
		"per_day":  types.Int64Type,
		"per_hour": types.Int64Type,
	}
}

func tokenQuotaAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"client_credentials": types.ObjectType{AttrTypes: tokenQuotaCCAttrTypes()},
	}
}

// -- schema ---------------------------------------------------------------

func (r *clientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

func (r *clientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computedString := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{Computed: true, Description: desc, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}}
	}
	optionalComputedString := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: desc}
	}
	optionalComputedBool := func(desc string) schema.BoolAttribute {
		return schema.BoolAttribute{Optional: true, Computed: true, Description: desc}
	}
	optionalComputedStringList := func(desc string) schema.ListAttribute {
		return schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: desc}
	}

	resp.Schema = schema.Schema{
		Description: "Manages an Auth0 application (a.k.a. client). See https://auth0.com/docs/get-started/applications.",
		Attributes: map[string]schema.Attribute{
			// -- identity --
			"id":            computedString("Auth0 client identifier (alias of `client_id`)."),
			"client_id":     computedString("Auth0-issued client identifier."),
			"client_secret": schema.StringAttribute{Computed: true, Sensitive: true, Description: "Auth0-issued client secret. Empty for public/native apps.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"tenant":        computedString("The tenant this client belongs to."),
			"global":        schema.BoolAttribute{Computed: true, Description: "True for the legacy 'All Applications' client representing tenant-wide settings."},

			// -- basic descriptors --
			"name":        schema.StringAttribute{Required: true, Description: "Application name (1+ chars, no `<` or `>`)."},
			"description": optionalComputedString("Free-text description (max 140 chars)."),
			"app_type":    optionalComputedString("Application type. One of: `native`, `spa`, `regular_web`, `non_interactive`, `sso_integration`, etc."),
			"logo_uri":    optionalComputedString("URL of the application logo (recommended 150x150)."),

			// -- behaviour flags --
			"is_first_party":                        optionalComputedBool("Whether this is a first-party client."),
			"oidc_conformant":                       optionalComputedBool("Whether this client conforms to strict OIDC specifications."),
			"sso":                                   optionalComputedBool("Whether Auth0 (rather than the IdP) handles Single Sign On."),
			"sso_disabled":                          optionalComputedBool("Whether Single Sign On is disabled."),
			"cross_origin_authentication":           optionalComputedBool("Whether cross-origin authentication is permitted."),
			"custom_login_page_on":                  optionalComputedBool("Whether a custom login page is used."),
			"is_token_endpoint_ip_header_trusted":   optionalComputedBool("Trust the `auth0-forwarded-for` header for brute-force protection."),
			"require_pushed_authorization_requests": optionalComputedBool("Require Pushed Authorization Requests for this client."),
			"require_proof_of_possession":           optionalComputedBool("Require Proof-of-Possession for this client."),
			"skip_non_verifiable_callback_uri_confirmation_prompt": optionalComputedBool("Skip confirmation prompt for non-verifiable callback URIs."),

			// -- URL collections --
			"callbacks":           optionalComputedStringList("URLs Auth0 may call back to after authentication."),
			"allowed_logout_urls": optionalComputedStringList("URLs that are valid to redirect to after logout."),
			"allowed_origins":     optionalComputedStringList("URLs allowed to make CORS requests to Auth0."),
			"web_origins":         optionalComputedStringList("Allowed origins for web message response mode and cross-origin auth."),
			"client_aliases":      optionalComputedStringList("Audiences/realms for the SAML / WS-Fed addons."),
			"allowed_clients":     optionalComputedStringList("Other clients allowed to make delegation requests."),

			// -- OAuth / OIDC --
			"grant_types":                optionalComputedStringList("OAuth grant types this application is permitted to use."),
			"token_endpoint_auth_method": optionalComputedString("Token endpoint authentication method (`none`, `client_secret_post`, `client_secret_basic`)."),
			"cross_origin_loc":           optionalComputedString("URL where cross-origin verification takes place."),
			"initiate_login_uri":         optionalComputedString("URL Auth0 redirects to to initiate login (must be HTTPS)."),
			"form_template":              optionalComputedString("HTML template used for WS-Federation."),
			"custom_login_page":          optionalComputedString("HTML/CSS/JS for the custom login page."),
			"custom_login_page_preview":  optionalComputedString("Preview HTML/CSS/JS for the custom login page."),
			"par_request_expiry":         schema.Int64Attribute{Optional: true, Computed: true, Description: "Pushed Authorization Request URI lifetime, in seconds."},
			"compliance_level":           optionalComputedString("Compliance level (e.g. `none`, `fapi1_adv_pkj_par`, `fapi1_adv_mtls_par`)."),
			"third_party_security_mode":  optionalComputedString("Third-party security mode for this client."),
			"redirection_policy":         optionalComputedString("Redirection policy for this client."),
			"jwks_uri":                   optionalComputedString("JWKS URI for CIMD clients using `private_key_jwt`."),

			// -- organization integration --
			"organization_usage":             optionalComputedString("Organization usage policy: `deny`, `allow`, or `require`."),
			"organization_require_behavior":  optionalComputedString("How organisations are required: `no_prompt`, `pre_login_prompt`, `post_login_prompt`."),
			"organization_discovery_methods": optionalComputedStringList("Allowed organization discovery methods (`email`, `organization_name`)."),

			// -- external / CIMD --
			"resource_server_identifier":   optionalComputedString("Identifier of the resource server this client is linked to."),
			"external_metadata_type":       optionalComputedString("External metadata type."),
			"external_metadata_created_by": optionalComputedString("External metadata `created_by` source."),
			"external_client_id":           optionalComputedString("Alternate client identifier for CIMD-based authorization flows."),

			// -- metadata --
			"client_metadata": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Free-form metadata attached to the client.",
			},

			// -- nested objects (Wave 1) --
			"signing_keys": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Signing certificates (read-only).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"pkcs7":   schema.StringAttribute{Computed: true, Description: "PKCS#7 (.P7B) public key + chain."},
						"cert":    schema.StringAttribute{Computed: true, Description: "X.509 (.CER) public key."},
						"subject": schema.StringAttribute{Computed: true, Description: "Certificate subject (`/CN={domain}`)."},
					},
				},
			},
			"token_quota": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Token quota configuration for this client.",
				Attributes: map[string]schema.Attribute{
					"client_credentials": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Quota applied to the client_credentials grant.",
						Attributes: map[string]schema.Attribute{
							"enforce":  schema.BoolAttribute{Optional: true, Computed: true, Description: "If true the quota is hard-enforced; if false, only logged."},
							"per_day":  schema.Int64Attribute{Optional: true, Computed: true, Description: "Maximum number of tokens issued per day."},
							"per_hour": schema.Int64Attribute{Optional: true, Computed: true, Description: "Maximum number of tokens issued per hour."},
						},
					},
				},
			},

			// -- Wave 2 nested objects -------------------------------------
			"jwt_configuration": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JWT issuance settings for this client.",
				Attributes: map[string]schema.Attribute{
					"lifetime_in_seconds": schema.Int64Attribute{Optional: true, Computed: true, Description: "JWT lifetime, affecting the `exp` claim."},
					"secret_encoded":      schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether the client secret is base64-encoded."},
					"alg":                 schema.StringAttribute{Optional: true, Computed: true, Description: "Signing algorithm (e.g. `RS256`, `HS256`)."},
					"scopes":              schema.StringAttribute{Optional: true, Computed: true, Description: "Free-form scopes object, encoded as JSON."},
				},
			},
			"refresh_token": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Refresh-token behaviour for this client.",
				Attributes: map[string]schema.Attribute{
					"rotation_type":                schema.StringAttribute{Optional: true, Computed: true, Description: "`rotating` or `non-rotating`."},
					"expiration_type":              schema.StringAttribute{Optional: true, Computed: true, Description: "`expiring` or `non-expiring`."},
					"leeway":                       schema.Int64Attribute{Optional: true, Computed: true, Description: "Period (s) where a previous refresh token can be re-exchanged."},
					"token_lifetime":               schema.Int64Attribute{Optional: true, Computed: true, Description: "Refresh-token lifetime (s)."},
					"infinite_token_lifetime":      schema.BoolAttribute{Optional: true, Computed: true, Description: "If true, refresh tokens never expire (overrides `token_lifetime`)."},
					"idle_token_lifetime":          schema.Int64Attribute{Optional: true, Computed: true, Description: "Refresh-token idle lifetime (s)."},
					"infinite_idle_token_lifetime": schema.BoolAttribute{Optional: true, Computed: true, Description: "If true, idle refresh tokens never expire."},
					"policies": schema.ListNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Multi-resource refresh-token policies.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"audience": schema.StringAttribute{Required: true, Description: "Resource server identifier."},
								"scope":    schema.ListAttribute{Required: true, ElementType: types.StringType, Description: "Permissions granted under this policy."},
							},
						},
					},
				},
			},
			"oidc_logout": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "OIDC backchannel-logout configuration.",
				Attributes:  oidcLogoutSchemaAttrs(),
			},
			"oidc_backchannel_logout": schema.SingleNestedAttribute{
				Optional:           true,
				DeprecationMessage: "Use `oidc_logout` instead. The Auth0 API still accepts this field but it is being phased out and is not echoed in responses.",
				Description:        "Deprecated alias for `oidc_logout`. Write-only — not returned in API responses.",
				Attributes:         oidcLogoutSchemaAttrs(),
			},
			"encryption_key": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Public key & certificate used to encrypt assertions sent to this client.",
				Attributes: map[string]schema.Attribute{
					"pub":     schema.StringAttribute{Optional: true, Computed: true, Description: "RSA public key."},
					"cert":    schema.StringAttribute{Optional: true, Computed: true, Description: "X.509 certificate."},
					"subject": schema.StringAttribute{Optional: true, Computed: true, Description: "Certificate subject (`/CN={domain}`)."},
				},
			},
			"default_organization": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Default organisation injected into specified flows.",
				Attributes: map[string]schema.Attribute{
					"organization_id": schema.StringAttribute{Required: true, Description: "Organisation ID."},
					"flows":           schema.ListAttribute{Required: true, ElementType: types.StringType, Description: "Flows in which this default applies (e.g. `client_credentials`)."},
				},
			},
			"native_social_login": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Native social login provider toggles.",
				Attributes: map[string]schema.Attribute{
					"apple":    nativeSocialLoginProviderSchemaAttr("Apple"),
					"facebook": nativeSocialLoginProviderSchemaAttr("Facebook"),
					"google":   nativeSocialLoginProviderSchemaAttr("Google"),
				},
			},
			"session_transfer": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Native ↔ Web Session Transfer Token settings.",
				Attributes: map[string]schema.Attribute{
					"can_create_session_transfer_token": schema.BoolAttribute{Optional: true, Computed: true, Description: "Allow this app to issue Session Transfer Tokens."},
					"enforce_cascade_revocation":        schema.BoolAttribute{Optional: true, Computed: true, Description: "Cascade-revoke dependants when the parent refresh token is revoked."},
					"allowed_authentication_methods":    schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: "Methods allowed to consume Session Transfer Tokens (`cookie`, `query`)."},
					"enforce_device_binding":            schema.StringAttribute{Optional: true, Computed: true, Description: "Device binding enforcement strategy."},
					"allow_refresh_token":               schema.BoolAttribute{Optional: true, Computed: true, Description: "Allow refresh tokens when authenticating with a Session Transfer Token."},
					"enforce_online_refresh_tokens":     schema.BoolAttribute{Optional: true, Computed: true, Description: "Tie refresh tokens to the Native↔Web session lifetime."},
					"delegation": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Delegation (impersonation) settings.",
						Attributes: map[string]schema.Attribute{
							"allow_delegated_access": schema.BoolAttribute{Optional: true, Computed: true, Description: "Allow delegated access via Session Transfer Tokens."},
							"enforce_device_binding": schema.StringAttribute{Optional: true, Computed: true, Description: "Device binding enforcement strategy for delegated flows."},
						},
					},
				},
			},
			"mobile": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Mobile-app configuration (Android & iOS).",
				Attributes: map[string]schema.Attribute{
					"android": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Android-specific settings.",
						Attributes: map[string]schema.Attribute{
							"app_package_name":         schema.StringAttribute{Optional: true, Computed: true, Description: "App package name from AndroidManifest.xml."},
							"sha256_cert_fingerprints": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: "SHA-256 fingerprints of the signing certificate."},
						},
					},
					"ios": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "iOS-specific settings.",
						Attributes: map[string]schema.Attribute{
							"team_id":               schema.StringAttribute{Optional: true, Computed: true, Description: "Apple developer team identifier."},
							"app_bundle_identifier": schema.StringAttribute{Optional: true, Computed: true, Description: "App bundle identifier (e.g. `com.example.MyApp`)."},
						},
					},
				},
			},
			"token_exchange": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Token-exchange configuration for this client.",
				Attributes: map[string]schema.Attribute{
					"allow_any_profile_of_type": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: "Enabled token-exchange profile types."},
				},
			},
			"my_organization_configuration": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "MyOrganization configuration for this client (read-only).",
				Attributes: map[string]schema.Attribute{
					"connection_profile_id":        schema.StringAttribute{Computed: true, Description: "Connection profile ID this client validates against."},
					"user_attribute_profile_id":    schema.StringAttribute{Computed: true, Description: "User-attribute profile ID this client validates against."},
					"allowed_strategies":           schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Allowed connection strategies."},
					"connection_deletion_behavior": schema.StringAttribute{Computed: true, Description: "Connection deletion behaviour."},
				},
			},
			"express_configuration": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Okta OIN express-configuration settings (read-only).",
				Attributes: map[string]schema.Attribute{
					"initiate_login_uri_template": schema.StringAttribute{Computed: true},
					"user_attribute_profile_id":   schema.StringAttribute{Computed: true},
					"connection_profile_id":       schema.StringAttribute{Computed: true},
					"enable_client":               schema.BoolAttribute{Computed: true},
					"enable_organization":         schema.BoolAttribute{Computed: true},
					"linked_clients": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"client_id": schema.StringAttribute{Computed: true},
							},
						},
					},
					"okta_oin_client_id": schema.StringAttribute{Computed: true},
					"admin_login_domain": schema.StringAttribute{Computed: true},
					"oin_submission_id":  schema.StringAttribute{Computed: true},
				},
			},

			// -- Wave 2 — list-of-enum + JSON-string passthroughs ----------
			"async_approval_notification_channels": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Notification channels for async approval (push, email, etc.).",
			},
			"signed_request_object": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "JAR (Signed Request Object) configuration as a JSON string. Shape depends on whether you supply `credentials` (PublicKey on Create / CredentialID on Update).",
			},
			"addons": addonsSchemaAttribute(),
			"client_authentication_methods": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Client authentication-methods configuration as a JSON string (`private_key_jwt`, `tls_client_auth`, `self_signed_tls_client_auth`).",
			},
		},
	}
}

func (r *clientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if m, ok := framework.ManagementFromResource(req, resp); ok {
		r.mgmt = m
	}
}

// -- CRUD ------------------------------------------------------------------

func (r *clientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &mgmt.CreateClientRequestContent{Name: plan.Name.ValueString()}
	applyCreateOptional(ctx, &plan, body, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.mgmt.Clients.Create(ctx, body)
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to create client", err)
		return
	}

	flattenCreate(ctx, &plan, created, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.mgmt.Clients.Get(ctx, state.ID.ValueString(), &mgmt.GetClientRequestParameters{})
	if err != nil {
		if framework.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		framework.AddAPIError(&resp.Diagnostics, "Failed to read client", err)
		return
	}

	flattenGet(ctx, &state, got, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &mgmt.UpdateClientRequestContent{}
	body.Name = plan.Name.ValueStringPointer()
	applyUpdateOptional(ctx, &plan, body, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.mgmt.Clients.Update(ctx, plan.ID.ValueString(), body)
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to update client", err)
		return
	}

	flattenUpdate(ctx, &plan, updated, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.mgmt.Clients.Delete(ctx, state.ID.ValueString()); err != nil && !framework.IsNotFound(err) {
		framework.AddAPIError(&resp.Diagnostics, "Failed to delete client", err)
	}
}

func (r *clientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// -- schema-only helpers ---------------------------------------------------

// oidcLogoutSchemaAttrs returns the attribute map shared by both
// `oidc_logout` and the deprecated `oidc_backchannel_logout` aliases.
func oidcLogoutSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"backchannel_logout_urls": schema.ListAttribute{
			Optional: true, Computed: true, ElementType: types.StringType,
			Description: "URLs Auth0 may call back for OIDC backchannel logout (currently a single URL).",
		},
		"backchannel_logout_initiators": schema.SingleNestedAttribute{
			Optional: true, Computed: true,
			Description: "Which logout initiators trigger the backchannel call.",
			Attributes: map[string]schema.Attribute{
				"mode":                schema.StringAttribute{Optional: true, Computed: true, Description: "`all` or `custom`."},
				"selected_initiators": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: "Initiators to include when mode is `custom`."},
			},
		},
		"backchannel_logout_session_metadata": schema.SingleNestedAttribute{
			Optional: true, Computed: true,
			Description: "Whether session metadata is included in the logout token.",
			Attributes: map[string]schema.Attribute{
				"include": schema.BoolAttribute{Optional: true, Computed: true, Description: "Include session metadata in the logout token."},
			},
		},
	}
}

// nativeSocialLoginProviderSchemaAttr returns the schema for one provider
// inside `native_social_login.{apple,facebook,google}`.
func nativeSocialLoginProviderSchemaAttr(provider string) schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Native " + provider + " sign-in toggle.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{Optional: true, Computed: true, Description: "Whether native " + provider + " sign-in is enabled."},
		},
	}
}
