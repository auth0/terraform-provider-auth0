// Package organization implements the auth0_organization resource.
//
// The schema mirrors every field returned by Auth0's Get/Update Organization
// endpoints so users always see a faithful projection in state. The Create
// endpoint additionally returns `enabled_connections`; that field is exposed
// as a separate `auth0_organization_connections` resource (planned), the same
// pattern the SDKv2 provider uses, to avoid permanent drift between Create
// and Get responses.
package organization

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
	_ resource.Resource                = (*organizationResource)(nil)
	_ resource.ResourceWithConfigure   = (*organizationResource)(nil)
	_ resource.ResourceWithImportState = (*organizationResource)(nil)
)

// NewResource returns a fresh organization resource implementation.
func NewResource() resource.Resource { return &organizationResource{} }

type organizationResource struct {
	mgmt *mgmtclient.Management
}

// model mirrors the auth0_organization HCL schema.
type model struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Metadata    types.Map    `tfsdk:"metadata"`
	Branding    types.Object `tfsdk:"branding"`
	TokenQuota  types.Object `tfsdk:"token_quota"`

	// Write-only — accepted by the Create endpoint to bootstrap initial
	// connections, but never returned by Get/Update. Use the dedicated
	// `auth0_organization_connections` resource (planned) for ongoing
	// management.
	EnabledConnections types.List `tfsdk:"enabled_connections"`
}

// -- nested attribute type maps --------------------------------------------

func colorsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"primary":         types.StringType,
		"page_background": types.StringType,
	}
}

func brandingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"logo_url": types.StringType,
		"colors":   types.ObjectType{AttrTypes: colorsAttrTypes()},
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

func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Auth0 Organization. See https://auth0.com/docs/manage-users/organizations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "The Auth0-assigned organization identifier (e.g. `org_xxx`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The logical name of the organization (lowercase, no spaces).",
			},
			"display_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Human-friendly name of the organization.",
			},
			"metadata": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Free-form metadata key/value pairs. Maximum 10 keys.",
			},
			"branding": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Branding overrides applied to the Universal Login experience.",
				Attributes: map[string]schema.Attribute{
					"logo_url": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "URL of the organization logo shown on the login page.",
					},
					"colors": schema.SingleNestedAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Color customisations.",
						Attributes: map[string]schema.Attribute{
							"primary": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Hex color used for primary elements (e.g. `#0059d6`).",
							},
							"page_background": schema.StringAttribute{
								Optional:    true,
								Computed:    true,
								Description: "Hex color used as the page background.",
							},
						},
					},
				},
			},
			"token_quota": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Token quota configuration for this organization.",
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
			"enabled_connections": schema.ListNestedAttribute{
				Optional: true,
				Description: "Connections enabled for this organization at create-time. " +
					"Write-only: the Auth0 Get/Update Organization endpoints do NOT return " +
					"this field, so changes after the initial apply are not picked up. " +
					"Use the (planned) `auth0_organization_connections` resource for ongoing management.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"connection_id":              schema.StringAttribute{Required: true, Description: "Connection ID."},
						"assign_membership_on_login": schema.BoolAttribute{Optional: true, Description: "Auto-grant org membership on login."},
						"show_as_button":             schema.BoolAttribute{Optional: true, Description: "Show this connection as a button on the org's login prompt (enterprise only)."},
						"is_signup_enabled":          schema.BoolAttribute{Optional: true, Description: "Allow signup via this connection (database connections only)."},
					},
				},
			},
		},
	}
}

func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if m, ok := framework.ManagementFromResource(req, resp); ok {
		r.mgmt = m
	}
}

// -- CRUD ------------------------------------------------------------------

func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &mgmt.CreateOrganizationRequestContent{Name: plan.Name.ValueString()}
	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		body.DisplayName = plan.DisplayName.ValueStringPointer()
	}
	if md, d := expandMetadata(ctx, plan.Metadata); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	} else if md != nil {
		body.Metadata = &md
	}
	if b, d := expandBranding(ctx, plan.Branding); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	} else {
		body.Branding = b
	}
	if tq, d := expandTokenQuotaCreate(ctx, plan.TokenQuota); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	} else if tq != nil {
		body.TokenQuota = tq
	}
	if conns := expandEnabledConnections(plan.EnabledConnections); conns != nil {
		body.EnabledConnections = conns
	}

	created, err := r.mgmt.Organizations.Create(ctx, body)
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to create organization", err)
		return
	}

	plan.ID = types.StringValue(created.GetID())
	flattenInto(ctx, &plan, created.Name, created.DisplayName, created.Metadata, created.Branding, created.TokenQuota, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.mgmt.Organizations.Get(ctx, state.ID.ValueString())
	if err != nil {
		if framework.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		framework.AddAPIError(&resp.Diagnostics, "Failed to read organization", err)
		return
	}

	flattenInto(ctx, &state, got.Name, got.DisplayName, got.Metadata, got.Branding, got.TokenQuota, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &mgmt.UpdateOrganizationRequestContent{}
	body.Name = plan.Name.ValueStringPointer()
	if !plan.DisplayName.IsNull() && !plan.DisplayName.IsUnknown() {
		body.DisplayName = plan.DisplayName.ValueStringPointer()
	}
	if md, d := expandMetadata(ctx, plan.Metadata); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	} else if md != nil {
		body.Metadata = &md
	}
	if b, d := expandBranding(ctx, plan.Branding); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	} else {
		body.Branding = b
	}
	if tq, d := expandTokenQuotaUpdate(ctx, plan.TokenQuota); d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	} else if tq != nil {
		body.TokenQuota = tq
	}

	updated, err := r.mgmt.Organizations.Update(ctx, plan.ID.ValueString(), body)
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to update organization", err)
		return
	}

	flattenInto(ctx, &plan, updated.Name, updated.DisplayName, updated.Metadata, updated.Branding, updated.TokenQuota, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.mgmt.Organizations.Delete(ctx, state.ID.ValueString()); err != nil && !framework.IsNotFound(err) {
		framework.AddAPIError(&resp.Diagnostics, "Failed to delete organization", err)
	}
}

func (r *organizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
