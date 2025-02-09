package resourceserver

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/framework/error"
)

type dataSourceType struct {
	cfg *config.Config
}

type dataSourceModel struct {
	resourceModel
	Scopes types.Set `tfsdk:"scopes"`
}

// NewDataSource will return a new auth0_resource_server data source.
func NewDataSource() datasource.DataSource {
	return &dataSourceType{}
}

// Configure will be called by the framework to configure the auth0_resource_server datasource.
func (r *dataSourceType) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if request.ProviderData != nil {
		r.cfg = request.ProviderData.(*config.Config)
	}
}

// Metadata will be called by the framework to get the type name for the auth0_resource_server datasource.
func (r *dataSourceType) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "auth0_resource_server"
}

// Schema will be called by the framework to get the schema for the auth0_resource_server datasource.
func (r *dataSourceType) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	if response != nil {
		response.Schema = schema.Schema{
			Description: "With this datasource, you can set up APIs that can be consumed from your authorized applications.",
			Attributes: map[string]schema.Attribute{
				"resource_server_id": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.AtLeastOneOf(
							path.MatchRelative().AtParent().AtName("identifier"),
						),
					},
					Description: "The ID of the resource server. If not provided, `identifier` must be set.",
				},
				"identifier": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.AtLeastOneOf(
							path.MatchRelative().AtParent().AtName("resource_server_id"),
						),
					},
					Description: "Unique identifier for the resource server. Used as the audience parameter " +
						"for authorization calls. If not provided, resource_server_id must be set. ",
					MarkdownDescription: "Unique identifier for the resource server. Used as the audience parameter " +
						"for authorization calls. If not provided, `resource_server_id` must be set. ",
				},
				"scopes": schema.SetNestedAttribute{
					Description: "List of permissions (scopes) used by this resource server.",
					Computed:    true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed:            true,
								Description:         "Name of the permission (scope). Examples include read:appointments or delete:appointments.",
								MarkdownDescription: "Name of the permission (scope). Examples include `read:appointments` or `delete:appointments`.",
							},
							"description": schema.StringAttribute{
								Computed:    true,
								Description: "Description of the permission (scope).",
							},
						},
					},
				},
				"verification_location": schema.StringAttribute{
					Computed: true,
					Description: "URL from which to retrieve JWKs for this resource server. " +
						"Used for verifying the JWT sent to Auth0 for token introspection.",
				},
				"enforce_policies": schema.BoolAttribute{
					Computed: true,
					Description: "If this setting is enabled, RBAC authorization policies will be enforced for this API. " +
						"Role and permission assignments will be evaluated during the login transaction.",
				},
				"token_dialect": schema.StringAttribute{
					Computed: true,
					Description: "Dialect of access tokens that should be issued for this resource server. " +
						"Options include access_token, rfc9068_profile, access_token_authz, and rfc9068_profile_authz. " +
						"access_token is a JWT containing standard Auth0 claims. rfc9068_profile is a JWT conforming to the IETF JWT Access Token Profile. " +
						"access_token_authz is a JWT containing standard Auth0 claims, including RBAC permissions claims. rfc9068_profile_authz is a JWT conforming to the IETF JWT Access Token Profile, including RBAC permissions claims. " +
						"RBAC permissions claims are available if RBAC (enforce_policies) is enabled for this API. " +
						"For more details, refer to Access Token Profiles(https://auth0.com/docs/secure/tokens/access-tokens/access-token-profiles).",
					MarkdownDescription: "Dialect of access tokens that should be issued for this resource server. " +
						"Options include `access_token`, `rfc9068_profile`, `access_token_authz`, and `rfc9068_profile_authz`. " +
						"`access_token` is a JWT containing standard Auth0 claims. `rfc9068_profile` is a JWT conforming to the IETF JWT Access Token Profile. " +
						"`access_token_authz` is a JWT containing standard Auth0 claims, including RBAC permissions claims. `rfc9068_profile_authz` is a JWT conforming to the IETF JWT Access Token Profile, including RBAC permissions claims. " +
						"RBAC permissions claims are available if RBAC (`enforce_policies`) is enabled for this API. " +
						"For more details, refer to [Access Token Profiles](https://auth0.com/docs/secure/tokens/access-tokens/access-token-profiles).",
				},
				"name": schema.StringAttribute{
					Computed:    true,
					Description: "Friendly name for the resource server.",
				},
				"signing_alg": schema.StringAttribute{
					Computed:            true,
					Description:         "Algorithm used to sign JWTs. Options include HS256, RS256, and PS256.",
					MarkdownDescription: "Algorithm used to sign JWTs. Options include `HS256`, `RS256`, and `PS256`.",
				},
				"signing_secret": schema.StringAttribute{
					Computed:    true,
					Description: "Secret used to sign tokens when using symmetric algorithms (HS256).",
				},
				"allow_offline_access": schema.BoolAttribute{
					Computed:    true,
					Description: "Indicates whether refresh tokens can be issued for this resource server.",
				},
				"token_lifetime": schema.Int64Attribute{
					Computed: true,
					Description: "Number of seconds during which access tokens issued for this resource " +
						"server from the token endpoint remain valid.",
				},
				"token_lifetime_for_web": schema.Int64Attribute{
					Computed: true,
					Description: "Number of seconds during which access tokens issued for this resource server via " +
						"implicit or hybrid flows remain valid. Cannot be greater than the token_lifetime value.",
					MarkdownDescription: "Number of seconds during which access tokens issued for this resource server via " +
						"implicit or hybrid flows remain valid. Cannot be greater than the `token_lifetime` value.",
				},
				"skip_consent_for_verifiable_first_party_clients": schema.BoolAttribute{
					Computed:    true,
					Description: "Indicates whether to skip user consent for applications flagged as first party.",
				},
				"consent_policy": schema.StringAttribute{
					Computed: true,
					Description: "Consent policy for this resource server. " +
						"Options include transactional-authorization-with-mfa, or null to disable.",
					MarkdownDescription: "Consent policy for this resource server. " +
						"Options include `transactional-authorization-with-mfa`, or `null` to disable.",
				},
				"authorization_details": schema.ListNestedAttribute{
					Computed:    true,
					Description: "Authorization details for this resource server.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Computed:    true,
								Description: "Type of authorization details.",
							},
						},
					},
				},
				"token_encryption": schema.SingleNestedAttribute{
					Optional:    true,
					Description: "Configuration for JSON Web Encryption(JWE) of tokens for this resource server.",
					Attributes: map[string]schema.Attribute{
						"format": schema.StringAttribute{
							Computed: true,
							Description: "Format of the token encryption. " +
								"Only compact-nested-jwe is supported.",
							MarkdownDescription: "Format of the token encryption. " +
								"Only `compact-nested-jwe` is supported.",
						},
						"encryption_key": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Authorization details for this resource server.",
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed:    true,
									Description: "Name of the encryption key.",
								},
								"kid": schema.StringAttribute{
									Computed:    true,
									Description: "Key ID.",
								},
								"algorithm": schema.StringAttribute{
									Computed:    true,
									Description: "Algorithm used to encrypt the token.",
								},
								"pem": schema.StringAttribute{
									Computed:    true,
									Description: "PEM-formatted public key. Must be JSON escaped.",
								},
							},
						},
					},
				},
				"proof_of_possession": schema.SingleNestedAttribute{
					Computed:    true,
					Description: "Configuration settings for proof-of-possession for this resource server.",
					Attributes: map[string]schema.Attribute{
						"mechanism": schema.StringAttribute{
							Computed: true,
							Description: "Mechanism used for proof-of-possession. " +
								"Only mtls is supported.",
							MarkdownDescription: "Mechanism used for proof-of-possession. " +
								"Only `mtls` is supported.",
						},
						"required": schema.BoolAttribute{
							Computed:    true,
							Description: "Indicates whether proof-of-possession is required with this resource server.",
						},
					},
				},
			},
		}
	}
}

// Read will be called by the framework to read an auth0_resource_server datasource.
func (r *dataSourceType) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	api := r.cfg.GetAPI()

	var resourceID *string
	response.Diagnostics.Append(request.Config.GetAttribute(ctx, path.Root("resource_server_id"), &resourceID)...)
	if response.Diagnostics.HasError() {
		return
	}
	if resourceID == nil {
		var identifier string
		response.Diagnostics.Append(request.Config.GetAttribute(ctx, path.Root("identifier"), &identifier)...)
		if response.Diagnostics.HasError() {
			return
		}
		identifier = url.PathEscape(identifier)
		resourceID = &identifier
	}

	resourceServer, err := api.ResourceServer.Read(ctx, *resourceID)
	if err != nil {
		response.Diagnostics.Append(internalError.Diagnostics(err)...)
		return
	}

	response.Diagnostics.Append(flattenResourceServerForDataSource(ctx, &response.State, resourceServer)...)
}
