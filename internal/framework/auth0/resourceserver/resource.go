package resourceserver

import (
	"context"
	"net/http"
	"time"

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
)

const auth0ManagementAPIName = "Auth0 Management API"

type resourceType struct {
	cfg *config.Config
}

type authorizationDetailsModel struct {
	Type types.String `tfsdk:"type"`
}

var authorizationDetailsElementTypeMap = map[string]attr.Type{
	"type": types.StringType,
}

var authorizationDetailsElementType = types.ObjectType{
	AttrTypes: authorizationDetailsElementTypeMap,
}

type encryptionKeyModel struct {
	Name      types.String `tfsdk:"name"`
	Algorithm types.String `tfsdk:"algorithm"`
	KID       types.String `tfsdk:"kid"`
	PEM       types.String `tfsdk:"pem"`
}

var encryptionKeyTypeMap = map[string]attr.Type{
	"name":      types.StringType,
	"algorithm": types.StringType,
	"kid":       types.StringType,
	"pem":       types.StringType,
}

var encryptionKeyType = types.ObjectType{
	AttrTypes: encryptionKeyTypeMap,
}

type tokenEncryptionModel struct {
	Format        types.String `tfsdk:"format"`
	EncryptionKey types.Object `tfsdk:"encryption_key"`
}

var tokenEncryptionTypeMap = map[string]attr.Type{
	"format":         types.StringType,
	"encryption_key": encryptionKeyType,
}

type proofOfPossessionModel struct {
	Mechanism types.String `tfsdk:"mechanism"`
	Required  types.Bool   `tfsdk:"required"`
}

var proofOfPossessionTypeMap = map[string]attr.Type{
	"mechanism": types.StringType,
	"required":  types.BoolType,
}

type resourceModel struct {
	ResourceServerID     types.String `tfsdk:"resource_server_id"`
	Identifier           types.String `tfsdk:"identifier"`
	TokenLifetime        types.Int64  `tfsdk:"token_lifetime"`
	SkipConsent          types.Bool   `tfsdk:"skip_consent_for_verifiable_first_party_clients"`
	Name                 types.String `tfsdk:"name"`
	SigningAlgorithm     types.String `tfsdk:"signing_alg"`
	SigningSecret        types.String `tfsdk:"signing_secret"`
	AllowOfflineAccess   types.Bool   `tfsdk:"allow_offline_access"`
	TokenLifetimeForWeb  types.Int64  `tfsdk:"token_lifetime_for_web"`
	EnforcePolicies      types.Bool   `tfsdk:"enforce_policies"`
	TokenDialect         types.String `tfsdk:"token_dialect"`
	VerificationLocation types.String `tfsdk:"verification_location"`
	AuthorizationDetails types.List   `tfsdk:"authorization_details"`
	TokenEncryption      types.Object `tfsdk:"token_encryption"`
	ConsentPolicy        types.String `tfsdk:"consent_policy"`
	ProofOfPossession    types.Object `tfsdk:"proof_of_possession"`
}

// NewResource will return a new auth0_resource_server resource.
func NewResource() resource.Resource {
	return &resourceType{}
}

// Configure will be called by the framework to configure the auth0_resource_server resource.
func (r *resourceType) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData != nil {
		r.cfg = request.ProviderData.(*config.Config)
	}
}

// Metadata will be called by the framework to get the type name for the auth0_resource_server resource.
func (r *resourceType) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "auth0_resource_server"
}

// Schema will be called by the framework to get the schema for the auth0_resource_server resource.
func (r *resourceType) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	if response != nil {
		response.Schema = schema.Schema{
			Description: "With this resource, you can set up APIs that can be consumed from your authorized applications.",
			Attributes: map[string]schema.Attribute{
				"resource_server_id": schema.StringAttribute{
					Computed:    true,
					Description: "A generated string identifying the resource server.",
				},
				"identifier": schema.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Description: "Unique identifier for the resource server. Used as the audience parameter " +
						"for authorization calls. Cannot be changed once set.",
				},
				"name": schema.StringAttribute{
					Optional:            true,
					Computed:            true,
					Description:         "Friendly name for the resource server. Cannot include < or > characters.",
					MarkdownDescription: "Friendly name for the resource server. Cannot include `<` or `>` characters.",
				},
				"signing_alg": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.OneOf(
							"HS256",
							"RS256",
							"PS256",
						),
					},
					Description:         "Algorithm used to sign JWTs. Options include HS256, RS256, and PS256.",
					MarkdownDescription: "Algorithm used to sign JWTs. Options include `HS256`, `RS256`, and `PS256`.",
				},
				"signing_secret": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.UTF8LengthAtLeast(16),
					},
					Description: "Secret used to sign tokens when using symmetric algorithms (HS256).",
				},
				"allow_offline_access": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Indicates whether refresh tokens can be issued for this resource server.",
				},
				"token_lifetime": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Description: "Number of seconds during which access tokens issued for this resource " +
						"server from the token endpoint remain valid.",
				},
				"token_lifetime_for_web": schema.Int64Attribute{
					Optional: true,
					Computed: true,
					Description: "Number of seconds during which access tokens issued for this resource server via " +
						"implicit or hybrid flows remain valid. Cannot be greater than the token_lifetime value.",
					MarkdownDescription: "Number of seconds during which access tokens issued for this resource server via " +
						"implicit or hybrid flows remain valid. Cannot be greater than the `token_lifetime` value.",
				},
				"skip_consent_for_verifiable_first_party_clients": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Description: "Indicates whether to skip user consent for applications flagged as first party.",
				},
				"verification_location": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Description: "URL from which to retrieve JWKs for this resource server. " +
						"Used for verifying the JWT sent to Auth0 for token introspection.",
				},
				"enforce_policies": schema.BoolAttribute{
					Computed: true,
					Optional: true,
					Description: "If this setting is enabled, RBAC authorization policies will be enforced for this API. " +
						"Role and permission assignments will be evaluated during the login transaction.",
				},
				"token_dialect": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.OneOf(
							"access_token",
							"access_token_authz",
							"rfc9068_profile",
							"rfc9068_profile_authz",
						),
					},
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
				"consent_policy": schema.StringAttribute{
					Optional: true,
					Computed: true,
					Validators: []validator.String{
						stringvalidator.OneOf(
							"transactional-authorization-with-mfa",
							"null",
						),
					},
					Description: "Consent policy for this resource server. " +
						"Options include transactional-authorization-with-mfa, or null to disable.",
					MarkdownDescription: "Consent policy for this resource server. " +
						"Options include `transactional-authorization-with-mfa`, or `null` to disable.",
				},
				"authorization_details": schema.ListNestedAttribute{
					Optional:    true,
					Description: "Authorization details for this resource server.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Required:    true,
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
							Optional: true, // This is actually false unless disabled, but then the block is required. It is being enforced with AlsoRequires.
							Validators: []validator.String{
								stringvalidator.OneOf(
									"compact-nested-jwe",
								),
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("encryption_key"),
								),
							},
							Description: "Format of the token encryption. " +
								"Only compact-nested-jwe is supported.",
							MarkdownDescription: "Format of the token encryption. " +
								"Only `compact-nested-jwe` is supported.",
						},
						"encryption_key": schema.SingleNestedAttribute{
							Optional: true,
							Validators: []validator.Object{
								objectvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("format"),
								),
							},
							Description: "Authorization details for this resource server.",
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Optional: true,
									Computed: true,
									Validators: []validator.String{
										stringvalidator.AlsoRequires(
											path.MatchRelative().AtParent().AtName("algorithm"),
											path.MatchRelative().AtParent().AtName("pem"),
										),
									},
									Description: "Name of the encryption key.",
								},
								"kid": schema.StringAttribute{
									Optional: true,
									Computed: true,
									Validators: []validator.String{
										stringvalidator.AlsoRequires(
											path.MatchRelative().AtParent().AtName("algorithm"),
											path.MatchRelative().AtParent().AtName("pem"),
										),
									},
									Description: "Key ID.",
								},
								"algorithm": schema.StringAttribute{
									Optional: true, // This is actually false, but then the block is required. It is being enforced with AlsoRequires.
									Validators: []validator.String{
										stringvalidator.AlsoRequires(
											path.MatchRelative().AtParent().AtName("pem"),
										),
									},
									Description: "Algorithm used to encrypt the token.",
								},
								"pem": schema.StringAttribute{
									Optional: true, // This is actually false, but then the block is required. It is being enforced with AlsoRequires.
									Validators: []validator.String{
										stringvalidator.AlsoRequires(
											path.MatchRelative().AtParent().AtName("algorithm"),
										),
									},
									Description: "PEM-formatted public key. Must be JSON escaped.",
								},
							},
						},
					},
				},
				"proof_of_possession": schema.SingleNestedAttribute{
					Optional:    true,
					Description: "Configuration settings for proof-of-possession for this resource server.",
					Attributes: map[string]schema.Attribute{
						"mechanism": schema.StringAttribute{
							Optional: true, // This is actually false unless disabled, but then the block is required. It is being enforced with AlsoRequires.
							Validators: []validator.String{
								stringvalidator.OneOf(
									"mtls",
								),
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("required"),
								),
							},
							Description: "Mechanism used for proof-of-possession. " +
								"Only mtls is supported.",
							MarkdownDescription: "Mechanism used for proof-of-possession. " +
								"Only `mtls` is supported.",
						},
						"required": schema.BoolAttribute{
							Optional: true, // This is actually false unless disabled, but then the block is required. It is being enforced with AlsoRequires.
							Computed: true,
							Validators: []validator.Bool{
								boolvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("mechanism"),
								),
							},
							Description: "Indicates whether proof-of-possession is required with this resource server.",
						},
					},
				},
			},
		}
	}
}

// ImportState will be called by the framework to import an existing auth0_resource_server resource.
func (r *resourceType) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_server_id"), request, response)
}

// Create will be called by the framework to initialise a new auth0_resource_server resource.
func (r *resourceType) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	api := r.cfg.GetAPI()
	configData, planData, stateData, diagnostics := internalSchema.GetRequestModels[resourceModel](ctx, &request.Config, &request.Plan, nil)
	response.Diagnostics.Append(diagnostics...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServer := expandResourceServer(ctx, configData, planData, stateData)

	if err := api.ResourceServer.Create(ctx, resourceServer); err != nil {
		response.Diagnostics.Append(internalError.Diagnostics(err)...)
		return
	}

	resourceID := resourceServer.GetID()

	time.Sleep(200 * time.Millisecond)
	if err := fixNullableAttributes(ctx, api, resourceID,
		isConsentPolicyNull(stateData.ConsentPolicy, planData.ConsentPolicy),
		isAuthorizationDetailsNull(ctx, stateData.AuthorizationDetails, planData.AuthorizationDetails),
		isTokenEncryptionNull(ctx, stateData.TokenEncryption, planData.TokenEncryption),
		isProofOfPossessionNull(ctx, stateData.ProofOfPossession, planData.ProofOfPossession),
		resourceServer,
	); err != nil {
		response.Diagnostics.Append(internalError.Diagnostics(err)...)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("resource_server_id"), resourceID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("consent_policy"), configData.ConsentPolicy)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("authorization_details"), configData.AuthorizationDetails)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("token_encryption"), configData.TokenEncryption)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("proof_of_possession"), configData.ProofOfPossession)...)
	response.Diagnostics.Append(flattenResourceServer(ctx, &response.State, resourceServer)...)
	if response.Diagnostics.HasError() {
		return
	}

	time.Sleep(200 * time.Millisecond)

	response.Diagnostics.Append(readResource(ctx, api, resourceID, &response.State)...)
}

// Update will be called by the framework to update an auth0_resource_server resource.
func (r *resourceType) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	api := r.cfg.GetAPI()

	configData, planData, stateData, diagnostics := internalSchema.GetRequestModels[resourceModel](ctx, &request.Config, &request.Plan, &request.State)
	response.Diagnostics.Append(diagnostics...)
	if response.Diagnostics.HasError() {
		return
	}
	resourceServer := expandResourceServer(ctx, configData, planData, stateData)
	resourceID := stateData.ResourceServerID.ValueString()

	if err := api.ResourceServer.Update(ctx, resourceID, resourceServer); err != nil {
		response.Diagnostics.Append(internalError.HandleAPIError(ctx, &response.State, err)...)
		return
	}

	time.Sleep(200 * time.Millisecond)
	if err := fixNullableAttributes(ctx, api, resourceID,
		isConsentPolicyNull(stateData.ConsentPolicy, planData.ConsentPolicy),
		isAuthorizationDetailsNull(ctx, stateData.AuthorizationDetails, planData.AuthorizationDetails),
		isTokenEncryptionNull(ctx, stateData.TokenEncryption, planData.TokenEncryption),
		isProofOfPossessionNull(ctx, stateData.ProofOfPossession, planData.ProofOfPossession),
		resourceServer,
	); err != nil {
		response.Diagnostics.Append(internalError.HandleAPIError(ctx, &response.State, err)...)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("consent_policy"), configData.ConsentPolicy)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("authorization_details"), configData.AuthorizationDetails)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("token_encryption"), configData.TokenEncryption)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("proof_of_possession"), configData.ProofOfPossession)...)
	response.Diagnostics.Append(flattenResourceServer(ctx, &response.State, resourceServer)...)

	time.Sleep(200 * time.Millisecond)

	response.Diagnostics.Append(readResource(ctx, api, resourceID, &response.State)...)
}

func fixNullableAttributes(
	ctx context.Context,
	api *management.Management,
	resourceID string,
	consentPolicyNull, authorizationDetailsNull, tokenEncryptionNull, proofOfPossessionNull bool,
	resourceServer *management.ResourceServer,
) error {
	nullMap := make(map[string]interface{})

	if consentPolicyNull {
		nullMap["consent_policy"] = nil
		resourceServer.ConsentPolicy = nil
	}
	if authorizationDetailsNull {
		nullMap["authorization_details"] = nil
		resourceServer.AuthorizationDetails = nil
	}
	if tokenEncryptionNull {
		nullMap["token_encryption"] = nil
		resourceServer.TokenEncryption = nil
	}
	if proofOfPossessionNull {
		nullMap["proof_of_possession"] = nil
		resourceServer.ProofOfPossession = nil
	}
	if len(nullMap) > 0 {
		return api.Request(ctx, http.MethodPatch, api.URI("resource-servers", resourceID), nullMap)
	}
	return nil
}

// Read will be called by the framework to read an auth0_resource_server resource.
func (r *resourceType) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	api := r.cfg.GetAPI()

	var resourceID string
	response.Diagnostics.Append(request.State.GetAttribute(ctx, path.Root("resource_server_id"), &resourceID)...)
	if !response.Diagnostics.HasError() {
		response.Diagnostics.Append(readResource(ctx, api, resourceID, &response.State)...)
	}
}

func readResource(ctx context.Context, api *management.Management, resourceID string, responseState *tfsdk.State) diag.Diagnostics {
	resourceServer, err := api.ResourceServer.Read(ctx, resourceID)
	if err != nil {
		return internalError.HandleAPIError(ctx, responseState, err)
	}

	rval := flattenResourceServer(ctx, responseState, resourceServer)
	return rval
}

// Delete will be called by the framework to delete an auth0_resource_server resource.
func (r *resourceType) Delete(ctx context.Context, _ resource.DeleteRequest, response *resource.DeleteResponse) {
	api := r.cfg.GetAPI()

	var resourceID string
	response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root("resource_server_id"), &resourceID)...)
	var name *string
	response.Diagnostics.Append(response.State.GetAttribute(ctx, path.Root("name"), &name)...)
	if response.Diagnostics.HasError() || name != nil && *name == auth0ManagementAPIName {
		return
	}

	if err := api.ResourceServer.Delete(ctx, resourceID); err != nil {
		response.Diagnostics.Append(internalError.HandleAPIError(ctx, &response.State, err)...)
	}
}
