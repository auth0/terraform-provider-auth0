// Package auth0clientgrant implements the auth0_client_grant resource.
//
// A client grant authorizes an Auth0 application (client) to call a specific
// API (resource server) under the Client Credentials flow, optionally with a
// set of scopes and organization-related constraints.
//
// Mutable fields (allowed in PATCH):
//   - scopes
//   - organization_usage
//   - allow_any_organization
//   - authorization_details_types
//   - allow_all_scopes
//
// Immutable fields (force-replace on change):
//   - client_id, audience, subject_type, default_for
package auth0clientgrant

import (
	"context"

	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

var (
	_ resource.Resource                = (*clientGrantResource)(nil)
	_ resource.ResourceWithConfigure   = (*clientGrantResource)(nil)
	_ resource.ResourceWithImportState = (*clientGrantResource)(nil)
)

// NewResource returns a fresh auth0_client_grant resource implementation.
func NewResource() resource.Resource { return &clientGrantResource{} }

type clientGrantResource struct {
	mgmt *mgmtclient.Management
}

// clientGrantResourceModel mirrors the HCL schema.
type clientGrantResourceModel struct {
	ID types.String `tfsdk:"id"`

	// Immutable identity fields.
	ClientID    types.String `tfsdk:"client_id"`
	Audience    types.String `tfsdk:"audience"`
	SubjectType types.String `tfsdk:"subject_type"`
	DefaultFor  types.String `tfsdk:"default_for"`

	// Mutable fields.
	Scopes                    types.List   `tfsdk:"scopes"`
	OrganizationUsage         types.String `tfsdk:"organization_usage"`
	AllowAnyOrganization      types.Bool   `tfsdk:"allow_any_organization"`
	AuthorizationDetailsTypes types.List   `tfsdk:"authorization_details_types"`
	AllowAllScopes            types.Bool   `tfsdk:"allow_all_scopes"`

	// Read-only.
	IsSystem types.Bool `tfsdk:"is_system"`
}

// -- metadata / schema --------------------------------------------------------

func (r *clientGrantResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_grant"
}

func (r *clientGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	requiresReplaceString := func() planmodifier.String {
		return stringplanmodifier.RequiresReplace()
	}
	useStateForUnknownString := func() planmodifier.String {
		return stringplanmodifier.UseStateForUnknown()
	}

	resp.Schema = schema.Schema{
		Description: "Manages a client grant: an authorization for an Auth0 client to call a specific API (resource server) using the Client Credentials flow. " +
			"See https://auth0.com/docs/get-started/applications/application-access-to-apis-client-grants.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Auth0-issued client grant identifier.",
				PlanModifiers: []planmodifier.String{
					useStateForUnknownString(),
				},
			},

			// Immutable identity.
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the Auth0 client (application) being granted access.",
				PlanModifiers: []planmodifier.String{
					requiresReplaceString(),
				},
			},
			"audience": schema.StringAttribute{
				Required:    true,
				Description: "Audience (API identifier) the client is granted access to.",
				PlanModifiers: []planmodifier.String{
					requiresReplaceString(),
				},
			},
			"subject_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Subject type for the grant. One of `client` or `user`.",
				PlanModifiers: []planmodifier.String{
					requiresReplaceString(),
					useStateForUnknownString(),
				},
			},
			"default_for": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Marks this grant as the default for a given category. Currently only `third_party_clients` is supported.",
				PlanModifiers: []planmodifier.String{
					requiresReplaceString(),
					useStateForUnknownString(),
				},
			},

			// Mutable.
			"scopes": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Scopes allowed for this client grant.",
			},
			"organization_usage": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Defines how organizations may be used with this grant. One of `deny`, `allow`, or `require`.",
				PlanModifiers: []planmodifier.String{
					useStateForUnknownString(),
				},
			},
			"allow_any_organization": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, any organization may be used with this grant. When false (default), the grant must be explicitly assigned to specific organizations.",
			},
			"authorization_details_types": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Types of `authorization_details` allowed for this grant (RFC 9396).",
			},
			"allow_all_scopes": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, all scopes configured on the resource server are allowed for this grant — overrides `scopes`.",
			},

			// Read-only.
			"is_system": schema.BoolAttribute{
				Computed:    true,
				Description: "When true, this grant is Auth0-managed and cannot be modified or deleted directly.",
			},
		},
	}
}

// -- configure ---------------------------------------------------------------

func (r *clientGrantResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if m, ok := framework.ManagementFromResource(req, resp); ok {
		r.mgmt = m
	}
}

// -- CRUD --------------------------------------------------------------------

func (r *clientGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clientGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, diags := expandCreate(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.mgmt.ClientGrants.Create(ctx, body)
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to create client grant", err)
		return
	}

	flattenCreate(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clientGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.mgmt.ClientGrants.Get(ctx, state.ID.ValueString())
	if err != nil {
		if framework.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		framework.AddAPIError(&resp.Diagnostics, "Failed to read client grant", err)
		return
	}

	flattenGet(&state, got)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clientGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state clientGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// is_system grants cannot be patched.
	if state.IsSystem.ValueBool() {
		resp.Diagnostics.AddError(
			"Client grant is system-managed",
			"This client grant is marked as `is_system: true` by Auth0 and cannot be modified.",
		)
		return
	}

	body, diags := expandUpdate(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated, err := r.mgmt.ClientGrants.Update(ctx, state.ID.ValueString(), body)
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to update client grant", err)
		return
	}

	flattenUpdate(&plan, updated)
	plan.ID = state.ID // preserve ID across update
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clientGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clientGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.IsSystem.ValueBool() {
		resp.Diagnostics.AddError(
			"Client grant is system-managed",
			"This client grant is marked as `is_system: true` by Auth0 and cannot be deleted.",
		)
		return
	}

	if err := r.mgmt.ClientGrants.Delete(ctx, state.ID.ValueString()); err != nil {
		if framework.IsNotFound(err) {
			return
		}
		framework.AddAPIError(&resp.Diagnostics, "Failed to delete client grant", err)
		return
	}
}

func (r *clientGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
