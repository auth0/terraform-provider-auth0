package auth0clientcredentials

import (
	"context"
	"fmt"

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
	_ resource.Resource                = (*credentialsResource)(nil)
	_ resource.ResourceWithConfigure   = (*credentialsResource)(nil)
	_ resource.ResourceWithImportState = (*credentialsResource)(nil)
)

// NewResource returns a fresh auth0_client_credentials resource implementation.
func NewResource() resource.Resource { return &credentialsResource{} }

type credentialsResource struct {
	mgmt *mgmtclient.Management
}

// secretBasedMethods is the set of `token_endpoint_auth_method` values that do
// not require credentials.
var secretBasedMethods = map[string]struct{}{
	"none":                {},
	"client_secret_post":  {},
	"client_secret_basic": {},
}

// credentialBasedMethods are the keys on `client_authentication_methods` that
// reference one or more credential IDs.
var credentialBasedMethods = map[string]struct{}{
	"private_key_jwt":             {},
	"tls_client_auth":             {},
	"self_signed_tls_client_auth": {},
}

func isSecretBased(m string) bool {
	_, ok := secretBasedMethods[m]
	return ok
}

func isCredentialBased(m string) bool {
	_, ok := credentialBasedMethods[m]
	return ok
}

// -- model --------------------------------------------------------------------

// model is the Terraform-facing state for one auth0_client_credentials.
type model struct {
	ID                   types.String `tfsdk:"id"`
	ClientID             types.String `tfsdk:"client_id"`
	AuthenticationMethod types.String `tfsdk:"authentication_method"`

	// Exactly one of the three credential-block lists is non-null when the
	// auth method is credential-based. Each is a list with max 1 element.
	PrivateKeyJWT           types.List `tfsdk:"private_key_jwt"`
	TLSClientAuth           types.List `tfsdk:"tls_client_auth"`
	SelfSignedTLSClientAuth types.List `tfsdk:"self_signed_tls_client_auth"`

	// Computed: the credential IDs owned by this resource. Useful for
	// references / audits.
	ManagedCredentialIDs types.List `tfsdk:"managed_credential_ids"`
}

// credentialAttrTypes returns the attr.Type map for one credential entry.
// Shared by all three credential-bearing blocks.
func credentialAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                     types.StringType,
		"name":                   types.StringType,
		"credential_type":        types.StringType,
		"pem":                    types.StringType,
		"algorithm":              types.StringType,
		"parse_expiry_from_cert": types.BoolType,
		"expires_at":             types.StringType,
		"kid":                    types.StringType,
		"subject_dn":             types.StringType,
		"thumbprint_sha256":      types.StringType,
		"created_at":             types.StringType,
		"updated_at":             types.StringType,
	}
}

// credentialBlockAttrTypes returns the attr.Type for the *wrapping* block
// object (which contains a `credentials` list).
func credentialBlockAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"credentials": types.ListType{
			ElemType: types.ObjectType{AttrTypes: credentialAttrTypes()},
		},
	}
}

// -- metadata / schema --------------------------------------------------------

func (r *credentialsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_credentials"
}

func (r *credentialsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	credentialNested := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Auth0-issued credential ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Friendly name for the credential.",
			},
			"credential_type": schema.StringAttribute{
				Required:    true,
				Description: "Credential type. One of `public_key`, `x509_cert`, `cert_subject_dn`.",
			},
			"pem": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "PEM-encoded public key or X.509 certificate. Mutually exclusive with `subject_dn`. Write-only — not returned on read.",
			},
			"algorithm": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Signing algorithm. One of `RS256`, `RS384`, `PS256`. Only applies to `public_key`.",
			},
			"parse_expiry_from_cert": schema.BoolAttribute{
				Optional:    true,
				Description: "When true, the credential's expiry is parsed from the supplied PEM (X.509). Only applies to `public_key`. Write-only.",
			},
			"expires_at": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ISO 8601 expiry timestamp. If unset and `parse_expiry_from_cert` is false, the credential never expires.",
			},
			"kid": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Custom key identifier. Auto-generated if not supplied.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subject_dn": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Subject Distinguished Name. Required for `cert_subject_dn` if `pem` is not provided. Mutually exclusive with `pem`.",
			},
			"thumbprint_sha256": schema.StringAttribute{
				Computed:    true,
				Description: "SHA-256 thumbprint of the X.509 certificate.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp at which the credential was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp at which the credential was last updated.",
			},
		},
	}

	credentialBlock := func(desc string) schema.ListNestedAttribute {
		return schema.ListNestedAttribute{
			Optional:    true,
			Description: desc,
			Validators:  nil, // size-1 validation enforced in CRUD for clarity.
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"credentials": schema.ListNestedAttribute{
						Required:     true,
						Description:  "Credentials enabled for this authentication method (1+ allowed).",
						NestedObject: credentialNested,
					},
				},
			},
		}
	}

	resp.Schema = schema.Schema{
		Description: "Manages the authentication method of an Auth0 client (application). Use this to switch a client between secret-based (`none`, `client_secret_post`, `client_secret_basic`) and credential-based (`private_key_jwt`, `tls_client_auth`, `self_signed_tls_client_auth`) authentication, and to manage the underlying credentials.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Same as `client_id`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the client whose authentication method is managed by this resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"authentication_method": schema.StringAttribute{
				Required: true,
				Description: "The authentication method to use. One of: " +
					"`none`, `client_secret_post`, `client_secret_basic`, " +
					"`private_key_jwt`, `tls_client_auth`, `self_signed_tls_client_auth`.",
			},
			"private_key_jwt":             credentialBlock("Private Key JWT authentication settings. Required when `authentication_method` is `private_key_jwt`."),
			"tls_client_auth":             credentialBlock("Mutual TLS (CA-signed) authentication settings. Required when `authentication_method` is `tls_client_auth`."),
			"self_signed_tls_client_auth": credentialBlock("Mutual TLS (self-signed) authentication settings. Required when `authentication_method` is `self_signed_tls_client_auth`."),
			"managed_credential_ids": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "IDs of credentials managed by this resource.",
			},
		},
	}
}

// -- configure ---------------------------------------------------------------

func (r *credentialsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if m, ok := framework.ManagementFromResource(req, resp); ok {
		r.mgmt = m
	}
}

// -- CRUD --------------------------------------------------------------------

func (r *credentialsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	method := plan.AuthenticationMethod.ValueString()
	if err := validateMethodAndBlocks(method, plan); err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}

	clientID := plan.ClientID.ValueString()

	// For credential-based methods, create the credentials first, then PATCH
	// the client to reference their IDs. For secret-based methods, just PATCH.
	var createdCredentials []*mgmt.PostClientCredentialResponseContent
	var createdIDs []string

	if isCredentialBased(method) {
		credList := credentialsListForMethod(method, plan)
		toCreate, d := credentialEntriesFromList(ctx, credList)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(toCreate) == 0 {
			resp.Diagnostics.AddError("Invalid configuration", fmt.Sprintf("`%s` block must contain at least one credential.", method))
			return
		}

		for i, ce := range toCreate {
			body, d := buildPostCredentialBody(ce)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				r.rollbackCredentials(ctx, clientID, createdIDs)
				return
			}
			created, err := r.mgmt.Clients.Credentials.Create(ctx, clientID, body)
			if err != nil {
				framework.AddAPIError(&resp.Diagnostics, fmt.Sprintf("Failed to create credential #%d", i+1), err)
				r.rollbackCredentials(ctx, clientID, createdIDs)
				return
			}
			createdCredentials = append(createdCredentials, created)
			if created != nil && created.ID != nil {
				createdIDs = append(createdIDs, *created.ID)
			}
		}

		// Now PATCH the client to point at these credentials.
		body := &mgmt.UpdateClientRequestContent{
			ClientAuthenticationMethods: buildClientAuthMethodsForRefs(method, createdIDs),
		}
		if _, err := r.mgmt.Clients.Update(ctx, clientID, body); err != nil {
			framework.AddAPIError(&resp.Diagnostics, "Failed to update client to reference new credentials", err)
			r.rollbackCredentials(ctx, clientID, createdIDs)
			return
		}
	} else {
		// Secret-based: PATCH token_endpoint_auth_method.
		enum, err := mgmt.NewClientTokenEndpointAuthMethodOrNullEnumFromString(method)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("authentication_method"), "Invalid authentication_method", err.Error())
			return
		}
		body := &mgmt.UpdateClientRequestContent{
			TokenEndpointAuthMethod: &enum,
		}
		if _, err := r.mgmt.Clients.Update(ctx, clientID, body); err != nil {
			framework.AddAPIError(&resp.Diagnostics, "Failed to set token_endpoint_auth_method", err)
			return
		}
	}

	// Build the state. Start from the plan (preserves write-only `pem`,
	// `parse_expiry_from_cert`) and overlay API-returned fields per credential.
	plan.ID = plan.ClientID
	plan.ManagedCredentialIDs = framework.StringSliceToList(createdIDs)

	// Re-flatten credential blocks if this was a credential-based create, so
	// computed fields (id, kid, thumbprint_sha256, created_at, etc.) are set.
	if isCredentialBased(method) {
		newList, d := flattenCreatedCredentialsIntoPlan(ctx, credentialsListForMethod(method, plan), createdCredentials)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		setCredentialsListForMethod(method, &plan, newList)
		nullifyOtherBlocks(method, &plan)
	} else {
		plan.PrivateKeyJWT = types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
		plan.TLSClientAuth = types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
		plan.SelfSignedTLSClientAuth = types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *credentialsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := state.ClientID.ValueString()
	got, err := r.mgmt.Clients.Get(ctx, clientID, &mgmt.GetClientRequestParameters{})
	if err != nil {
		if framework.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		framework.AddAPIError(&resp.Diagnostics, "Failed to read client", err)
		return
	}

	// Determine the current auth method and refresh state.
	if err := refreshStateFromClient(ctx, r.mgmt, clientID, got, &state); err != nil {
		resp.Diagnostics.AddError("Failed to refresh client_credentials state", err.Error())
		return
	}

	state.ID = state.ClientID
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *credentialsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newMethod := plan.AuthenticationMethod.ValueString()
	oldMethod := state.AuthenticationMethod.ValueString()
	if err := validateMethodAndBlocks(newMethod, plan); err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}

	clientID := plan.ClientID.ValueString()

	// Strategy: compute a definitive set of credentials we want AFTER the
	// update. Create the ones that don't exist yet (we identify "exists" by
	// the `id` attribute being known in state), then PATCH the client to
	// reference exactly those, then delete the ones we no longer want.

	var (
		newCreated      []*mgmt.PostClientCredentialResponseContent
		newCreatedIDs   []string
		finalCredIDs    []string // full final reference list
		toDeleteIDs     []string
		newCredsByIndex = map[int]*mgmt.PostClientCredentialResponseContent{}
	)

	// Helper to compute the prior credential IDs.
	priorIDs := stringListToGo(state.ManagedCredentialIDs)

	if isCredentialBased(newMethod) {
		planList := credentialsListForMethod(newMethod, plan)
		toResolve, d := credentialEntriesFromList(ctx, planList)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(toResolve) == 0 {
			resp.Diagnostics.AddError("Invalid configuration", fmt.Sprintf("`%s` block must contain at least one credential.", newMethod))
			return
		}

		// Walk the planned credentials. Any with a known ID (UseStateForUnknown
		// preserves IDs across updates) and unchanged immutable fields is
		// reused; anything else triggers a creation.
		stateList := credentialsListForMethod(oldMethod, state)
		stateEntries, _ := credentialEntriesFromList(ctx, stateList)
		stateByID := map[string]credentialEntry{}
		for _, sc := range stateEntries {
			if sc.ID != "" {
				stateByID[sc.ID] = sc
			}
		}

		for i, pe := range toResolve {
			// Reuse if a stable ID is present and the immutable fields match.
			if pe.ID != "" {
				if sc, ok := stateByID[pe.ID]; ok && credentialImmutableEqual(sc, pe) {
					finalCredIDs = append(finalCredIDs, pe.ID)
					continue
				}
			}
			body, d := buildPostCredentialBody(pe)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				r.rollbackCredentials(ctx, clientID, newCreatedIDs)
				return
			}
			created, err := r.mgmt.Clients.Credentials.Create(ctx, clientID, body)
			if err != nil {
				framework.AddAPIError(&resp.Diagnostics, fmt.Sprintf("Failed to create credential #%d", i+1), err)
				r.rollbackCredentials(ctx, clientID, newCreatedIDs)
				return
			}
			newCreated = append(newCreated, created)
			if created != nil && created.ID != nil {
				newCreatedIDs = append(newCreatedIDs, *created.ID)
				finalCredIDs = append(finalCredIDs, *created.ID)
				newCredsByIndex[i] = created
			}
		}

		// PATCH the client.
		body := &mgmt.UpdateClientRequestContent{
			ClientAuthenticationMethods: buildClientAuthMethodsForRefs(newMethod, finalCredIDs),
		}
		if _, err := r.mgmt.Clients.Update(ctx, clientID, body); err != nil {
			framework.AddAPIError(&resp.Diagnostics, "Failed to update client authentication methods", err)
			r.rollbackCredentials(ctx, clientID, newCreatedIDs)
			return
		}

		// Apply mutable changes (expires_at) to reused credentials.
		for _, pe := range toResolve {
			if pe.ID == "" {
				continue
			}
			sc, ok := stateByID[pe.ID]
			if !ok {
				continue
			}
			if pe.ExpiresAt != sc.ExpiresAt {
				patchBody := &mgmt.PatchClientCredentialRequestContent{}
				if pe.ExpiresAt != "" {
					ts, err := parseTimeRFC3339(pe.ExpiresAt)
					if err != nil {
						resp.Diagnostics.AddError("Invalid expires_at", err.Error())
						return
					}
					patchBody.ExpiresAt = &ts
				}
				if _, err := r.mgmt.Clients.Credentials.Update(ctx, clientID, pe.ID, patchBody); err != nil {
					framework.AddAPIError(&resp.Diagnostics, fmt.Sprintf("Failed to patch credential %s", pe.ID), err)
					return
				}
			}
		}

		// Compute deletion set: anything in priorIDs but not in finalCredIDs.
		finalSet := stringSet(finalCredIDs)
		for _, id := range priorIDs {
			if _, keep := finalSet[id]; !keep {
				toDeleteIDs = append(toDeleteIDs, id)
			}
		}
	} else {
		// New method is secret-based. PATCH client to set
		// token_endpoint_auth_method and (if previously credential-based)
		// implicitly drop client_authentication_methods.
		enum, err := mgmt.NewClientTokenEndpointAuthMethodOrNullEnumFromString(newMethod)
		if err != nil {
			resp.Diagnostics.AddAttributeError(path.Root("authentication_method"), "Invalid authentication_method", err.Error())
			return
		}
		body := &mgmt.UpdateClientRequestContent{
			TokenEndpointAuthMethod: &enum,
		}
		if _, err := r.mgmt.Clients.Update(ctx, clientID, body); err != nil {
			framework.AddAPIError(&resp.Diagnostics, "Failed to set token_endpoint_auth_method", err)
			return
		}
		// All previously-managed credentials are now orphaned.
		toDeleteIDs = append(toDeleteIDs, priorIDs...)
	}

	// Best-effort delete old credentials. We don't fail the update if
	// deletion fails — surface it as a warning so the user knows.
	for _, id := range toDeleteIDs {
		if err := r.mgmt.Clients.Credentials.Delete(ctx, clientID, id); err != nil && !framework.IsNotFound(err) {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Failed to delete old credential %s", id),
				err.Error()+" — you may need to delete it manually.",
			)
		}
	}

	// Build the new state.
	plan.ID = plan.ClientID
	plan.ManagedCredentialIDs = framework.StringSliceToList(finalCredIDs)
	if isCredentialBased(newMethod) {
		newList, d := flattenUpdatedCredentialsIntoPlan(ctx, credentialsListForMethod(newMethod, plan), newCredsByIndex)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		setCredentialsListForMethod(newMethod, &plan, newList)
		nullifyOtherBlocks(newMethod, &plan)
	} else {
		plan.PrivateKeyJWT = types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
		plan.TLSClientAuth = types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
		plan.SelfSignedTLSClientAuth = types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *credentialsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := state.ClientID.ValueString()

	// Reset the client to a safe default (`client_secret_post` — Auth0's
	// default for regular web apps). This implicitly drops any
	// client_authentication_methods reference.
	defaultMethod := mgmt.ClientTokenEndpointAuthMethodOrNullEnumClientSecretPost
	body := &mgmt.UpdateClientRequestContent{
		TokenEndpointAuthMethod: &defaultMethod,
	}
	if _, err := r.mgmt.Clients.Update(ctx, clientID, body); err != nil {
		// Don't fail destroy if the client itself is gone.
		if !framework.IsNotFound(err) {
			resp.Diagnostics.AddWarning("Failed to reset client authentication method", err.Error())
		}
	}

	for _, id := range stringListToGo(state.ManagedCredentialIDs) {
		if err := r.mgmt.Clients.Credentials.Delete(ctx, clientID, id); err != nil && !framework.IsNotFound(err) {
			resp.Diagnostics.AddWarning(
				fmt.Sprintf("Failed to delete credential %s", id),
				err.Error()+" — you may need to delete it manually.",
			)
		}
	}
}

// ImportState supports `terraform import auth0_client_credentials.<name> <client_id>`.
func (r *credentialsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("client_id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// -- helpers -----------------------------------------------------------------

func (r *credentialsResource) rollbackCredentials(ctx context.Context, clientID string, ids []string) {
	for _, id := range ids {
		_ = r.mgmt.Clients.Credentials.Delete(ctx, clientID, id)
	}
}
