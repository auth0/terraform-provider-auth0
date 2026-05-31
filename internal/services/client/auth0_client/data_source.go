// data_source.go — implements the auth0_client data source.
//
// The data source schema mirrors the resource's attribute set (every API
// field is exposed) but every attribute is `Computed`. Lookup is by
// `client_id`.
package auth0client

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"
	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

var (
	_ datasource.DataSource              = (*clientDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*clientDataSource)(nil)
)

// NewDataSource returns a fresh auth0_client data source implementation.
func NewDataSource() datasource.DataSource { return &clientDataSource{} }

type clientDataSource struct {
	mgmt *mgmtclient.Management
}

func (d *clientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client"
}

func (d *clientDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if m, ok := framework.ManagementFromDataSource(req, resp); ok {
		d.mgmt = m
	}
}

// dataSourceSchemaFromResource clones the resource schema as a data-source
// schema, marking every attribute Computed-only.
//
// We do this by hand (rather than reflectively) to:
//   - Keep `client_id` Required (it's the lookup key)
//   - Drop write-only / sensitive markers that don't apply to data sources
//   - Preserve all descriptions
func (d *clientDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Look up an Auth0 application (client) by its `client_id`.",
		Attributes: map[string]dsschema.Attribute{
			"id":        dsschema.StringAttribute{Computed: true, Description: "Auth0 client identifier (alias of `client_id`)."},
			"client_id": dsschema.StringAttribute{Required: true, Description: "Auth0-issued client identifier to look up."},

			"client_secret": dsschema.StringAttribute{Computed: true, Sensitive: true, Description: "Auth0-issued client secret. Empty for public/native apps."},
			"tenant":        dsschema.StringAttribute{Computed: true, Description: "The tenant this client belongs to."},
			"global":        dsschema.BoolAttribute{Computed: true, Description: "True for the legacy 'All Applications' client."},

			"name":        dsschema.StringAttribute{Computed: true, Description: "Application name."},
			"description": dsschema.StringAttribute{Computed: true, Description: "Free-text description."},
			"app_type":    dsschema.StringAttribute{Computed: true, Description: "Application type."},
			"logo_uri":    dsschema.StringAttribute{Computed: true, Description: "URL of the application logo."},

			"is_first_party":                        dsschema.BoolAttribute{Computed: true},
			"oidc_conformant":                       dsschema.BoolAttribute{Computed: true},
			"sso":                                   dsschema.BoolAttribute{Computed: true},
			"sso_disabled":                          dsschema.BoolAttribute{Computed: true},
			"cross_origin_authentication":           dsschema.BoolAttribute{Computed: true},
			"custom_login_page_on":                  dsschema.BoolAttribute{Computed: true},
			"is_token_endpoint_ip_header_trusted":   dsschema.BoolAttribute{Computed: true},
			"require_pushed_authorization_requests": dsschema.BoolAttribute{Computed: true},
			"require_proof_of_possession":           dsschema.BoolAttribute{Computed: true},
			"skip_non_verifiable_callback_uri_confirmation_prompt": dsschema.BoolAttribute{Computed: true},

			"callbacks":           dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
			"allowed_logout_urls": dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
			"allowed_origins":     dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
			"web_origins":         dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
			"client_aliases":      dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
			"allowed_clients":     dsschema.ListAttribute{Computed: true, ElementType: types.StringType},

			"grant_types":                dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
			"token_endpoint_auth_method": dsschema.StringAttribute{Computed: true},
			"cross_origin_loc":           dsschema.StringAttribute{Computed: true},
			"initiate_login_uri":         dsschema.StringAttribute{Computed: true},
			"form_template":              dsschema.StringAttribute{Computed: true},
			"custom_login_page":          dsschema.StringAttribute{Computed: true},
			"custom_login_page_preview":  dsschema.StringAttribute{Computed: true},
			"par_request_expiry":         dsschema.Int64Attribute{Computed: true},
			"compliance_level":           dsschema.StringAttribute{Computed: true},
			"third_party_security_mode":  dsschema.StringAttribute{Computed: true},
			"redirection_policy":         dsschema.StringAttribute{Computed: true},
			"jwks_uri":                   dsschema.StringAttribute{Computed: true},

			"organization_usage":             dsschema.StringAttribute{Computed: true},
			"organization_require_behavior":  dsschema.StringAttribute{Computed: true},
			"organization_discovery_methods": dsschema.ListAttribute{Computed: true, ElementType: types.StringType},

			"resource_server_identifier":   dsschema.StringAttribute{Computed: true},
			"external_metadata_type":       dsschema.StringAttribute{Computed: true},
			"external_metadata_created_by": dsschema.StringAttribute{Computed: true},
			"external_client_id":           dsschema.StringAttribute{Computed: true},

			"client_metadata": dsschema.MapAttribute{Computed: true, ElementType: types.StringType},

			// Nested objects — computed equivalents of the resource schema.
			"signing_keys": dsschema.ListNestedAttribute{
				Computed: true,
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"pkcs7":   dsschema.StringAttribute{Computed: true},
						"cert":    dsschema.StringAttribute{Computed: true},
						"subject": dsschema.StringAttribute{Computed: true},
					},
				},
			},
			"token_quota": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"client_credentials": dsschema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]dsschema.Attribute{
							"enforce":  dsschema.BoolAttribute{Computed: true},
							"per_day":  dsschema.Int64Attribute{Computed: true},
							"per_hour": dsschema.Int64Attribute{Computed: true},
						},
					},
				},
			},

			"jwt_configuration": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"lifetime_in_seconds": dsschema.Int64Attribute{Computed: true},
					"secret_encoded":      dsschema.BoolAttribute{Computed: true},
					"alg":                 dsschema.StringAttribute{Computed: true},
					"scopes":              dsschema.StringAttribute{Computed: true},
				},
			},
			"refresh_token": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"rotation_type":                dsschema.StringAttribute{Computed: true},
					"expiration_type":              dsschema.StringAttribute{Computed: true},
					"leeway":                       dsschema.Int64Attribute{Computed: true},
					"token_lifetime":               dsschema.Int64Attribute{Computed: true},
					"infinite_token_lifetime":      dsschema.BoolAttribute{Computed: true},
					"idle_token_lifetime":          dsschema.Int64Attribute{Computed: true},
					"infinite_idle_token_lifetime": dsschema.BoolAttribute{Computed: true},
					"policies": dsschema.ListNestedAttribute{
						Computed: true,
						NestedObject: dsschema.NestedAttributeObject{
							Attributes: map[string]dsschema.Attribute{
								"audience": dsschema.StringAttribute{Computed: true},
								"scope":    dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
							},
						},
					},
				},
			},
			"oidc_logout":             oidcLogoutDSAttrs(),
			"oidc_backchannel_logout": oidcLogoutDSAttrs(),
			"encryption_key": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"pub":     dsschema.StringAttribute{Computed: true},
					"cert":    dsschema.StringAttribute{Computed: true},
					"subject": dsschema.StringAttribute{Computed: true},
				},
			},
			"default_organization": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"organization_id": dsschema.StringAttribute{Computed: true},
					"flows":           dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
				},
			},
			"native_social_login": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"apple":    nativeSocialProviderDSAttr(),
					"facebook": nativeSocialProviderDSAttr(),
					"google":   nativeSocialProviderDSAttr(),
				},
			},
			"session_transfer": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"can_create_session_transfer_token": dsschema.BoolAttribute{Computed: true},
					"enforce_cascade_revocation":        dsschema.BoolAttribute{Computed: true},
					"allowed_authentication_methods":    dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
					"enforce_device_binding":            dsschema.StringAttribute{Computed: true},
					"allow_refresh_token":               dsschema.BoolAttribute{Computed: true},
					"enforce_online_refresh_tokens":     dsschema.BoolAttribute{Computed: true},
					"delegation": dsschema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]dsschema.Attribute{
							"allow_delegated_access": dsschema.BoolAttribute{Computed: true},
							"enforce_device_binding": dsschema.StringAttribute{Computed: true},
						},
					},
				},
			},
			"mobile": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"android": dsschema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]dsschema.Attribute{
							"app_package_name":         dsschema.StringAttribute{Computed: true},
							"sha256_cert_fingerprints": dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
						},
					},
					"ios": dsschema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]dsschema.Attribute{
							"team_id":               dsschema.StringAttribute{Computed: true},
							"app_bundle_identifier": dsschema.StringAttribute{Computed: true},
						},
					},
				},
			},
			"token_exchange": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"allow_any_profile_of_type": dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
				},
			},
			"my_organization_configuration": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"connection_profile_id":        dsschema.StringAttribute{Computed: true},
					"user_attribute_profile_id":    dsschema.StringAttribute{Computed: true},
					"allowed_strategies":           dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
					"connection_deletion_behavior": dsschema.StringAttribute{Computed: true},
				},
			},
			"express_configuration": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"initiate_login_uri_template": dsschema.StringAttribute{Computed: true},
					"user_attribute_profile_id":   dsschema.StringAttribute{Computed: true},
					"connection_profile_id":       dsschema.StringAttribute{Computed: true},
					"enable_client":               dsschema.BoolAttribute{Computed: true},
					"enable_organization":         dsschema.BoolAttribute{Computed: true},
					"linked_clients": dsschema.ListNestedAttribute{
						Computed: true,
						NestedObject: dsschema.NestedAttributeObject{
							Attributes: map[string]dsschema.Attribute{
								"client_id": dsschema.StringAttribute{Computed: true},
							},
						},
					},
					"okta_oin_client_id": dsschema.StringAttribute{Computed: true},
					"admin_login_domain": dsschema.StringAttribute{Computed: true},
					"oin_submission_id":  dsschema.StringAttribute{Computed: true},
				},
			},

			"async_approval_notification_channels": dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
			"signed_request_object":                dsschema.StringAttribute{Computed: true, Description: "JAR configuration as a JSON string."},
			"addons":                               dsschema.StringAttribute{Computed: true, Description: "Addons configuration as a JSON string."},
			"client_authentication_methods":        dsschema.StringAttribute{Computed: true, Description: "Client authentication methods configuration as a JSON string."},
		},
	}
}

// oidcLogoutDSAttrs is a small helper that returns the shared schema for the
// oidc_logout / oidc_backchannel_logout attributes.
func oidcLogoutDSAttrs() dsschema.SingleNestedAttribute {
	return dsschema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]dsschema.Attribute{
			"backchannel_logout_urls": dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
			"backchannel_logout_initiators": dsschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dsschema.Attribute{
					"mode":                dsschema.StringAttribute{Computed: true},
					"selected_initiators": dsschema.ListAttribute{Computed: true, ElementType: types.StringType},
				},
			},
		},
	}
}

func nativeSocialProviderDSAttr() dsschema.SingleNestedAttribute {
	return dsschema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]dsschema.Attribute{
			"enabled": dsschema.BoolAttribute{Computed: true},
		},
	}
}

// Read fetches the client by ID and flattens it into state using the same
// flattenGet helper the resource uses (data sources share the model struct).
func (d *clientDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg model
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if cfg.ClientID.IsNull() || cfg.ClientID.IsUnknown() || cfg.ClientID.ValueString() == "" {
		resp.Diagnostics.AddError("Missing client_id", "`client_id` is required to look up a client.")
		return
	}

	got, err := d.mgmt.Clients.Get(ctx, cfg.ClientID.ValueString(), &mgmt.GetClientRequestParameters{})
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to read client", err)
		return
	}

	state := cfg // start from config to keep `client_id`
	flattenGet(ctx, &state, got, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
