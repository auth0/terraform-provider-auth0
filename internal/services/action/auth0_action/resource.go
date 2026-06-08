// Package auth0action implements the auth0_action resource and data source.
//
// An Auth0 Action is a custom Node.js function that runs at an extensibility
// point (trigger) during an Auth0 flow. The lifecycle is:
//
//	create -> deploy -> bind to a trigger
//
// This resource manages the create/deploy half. Binding an action to a trigger
// is handled by the separate `auth0_trigger_bindings` resource, mirroring the
// API's separation of concerns (an action can exist and be deployed without
// being bound to any trigger).
//
// Deploy semantics: the Create endpoint accepts a `deploy` flag and deploys
// inline. The Update endpoint does NOT, so when `deploy = true` we issue an
// explicit Deploy call after updating to materialise a new immutable version.
package auth0action

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"
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
	_ resource.Resource                = (*actionResource)(nil)
	_ resource.ResourceWithConfigure   = (*actionResource)(nil)
	_ resource.ResourceWithImportState = (*actionResource)(nil)
)

// NewResource returns a fresh auth0_action resource implementation.
func NewResource() resource.Resource { return &actionResource{} }

type actionResource struct {
	mgmt *mgmtclient.Management
}

// model mirrors the auth0_action HCL schema.
type model struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	SupportedTriggers types.Object `tfsdk:"supported_triggers"`
	Code              types.String `tfsdk:"code"`
	Runtime           types.String `tfsdk:"runtime"`
	Dependencies      types.Set    `tfsdk:"dependencies"`
	Secrets           types.List   `tfsdk:"secrets"`
	Deploy            types.Bool   `tfsdk:"deploy"`
	VersionID         types.String `tfsdk:"version_id"`
}

func (r *actionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

func (r *actionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Auth0 Action — a custom Node.js function executed at an " +
			"extensibility point. See https://auth0.com/docs/customize/actions. Bind the " +
			"action to a trigger with the `auth0_trigger_bindings` resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "The Auth0-assigned action identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the action.",
			},
			"supported_triggers": schema.SingleNestedAttribute{
				Required: true,
				Description: "The trigger to which this action targets. An action can only " +
					"target a single trigger at a time.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required: true,
						Description: "The trigger identifier (e.g. `post-login`, " +
							"`credentials-exchange`).",
					},
					"version": schema.StringAttribute{
						Required:    true,
						Description: "The trigger version (e.g. `v1`, `v2`, `v3`).",
					},
				},
			},
			"code": schema.StringAttribute{
				Required:    true,
				Description: "The source code of the action.",
			},
			"runtime": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The Node runtime (e.g. `node22`). Defaults to the trigger's default runtime.",
			},
			"dependencies": schema.SetNestedAttribute{
				Optional:    true,
				Description: "The list of third-party npm modules this action depends on.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":    schema.StringAttribute{Required: true, Description: "The npm module name (e.g. `lodash`)."},
						"version": schema.StringAttribute{Required: true, Description: "The npm module version (e.g. `4.17.21`)."},
					},
				},
			},
			"secrets": schema.ListNestedAttribute{
				Optional: true,
				Description: "The list of secrets injected into the action's runtime. Secret " +
					"values are write-only: the API never returns them, so the configured value " +
					"is retained in state.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{Required: true, Description: "The secret name (e.g. `API_KEY`)."},
						"value": schema.StringAttribute{
							Required:    true,
							Sensitive:   true,
							Description: "The secret value. Write-only; never returned by the API.",
						},
					},
				},
			},
			"deploy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to deploy the action after creating or updating it. Defaults to false.",
			},
			"version_id": schema.StringAttribute{
				Computed:      true,
				Description:   "The ID of the currently deployed action version (empty if never deployed).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *actionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if m, ok := framework.ManagementFromResource(req, resp); ok {
		r.mgmt = m
	}
}

// -- CRUD ------------------------------------------------------------------

func (r *actionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &mgmt.CreateActionRequestContent{Name: plan.Name.ValueString()}
	expandActionInto(ctx, &plan, &body.SupportedTriggers, &body.Code, &body.Runtime, &body.Dependencies, &body.Secrets, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// The Create endpoint can deploy inline.
	if plan.Deploy.ValueBool() {
		body.Deploy = plan.Deploy.ValueBoolPointer()
	}

	created, err := r.mgmt.Actions.Create(ctx, body)
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to create action", err)
		return
	}

	got, err := r.mgmt.Actions.Get(ctx, created.GetID())
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to read action after create", err)
		return
	}

	flattenInto(&plan, got, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *actionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.mgmt.Actions.Get(ctx, state.ID.ValueString())
	if err != nil {
		if framework.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		framework.AddAPIError(&resp.Diagnostics, "Failed to read action", err)
		return
	}

	flattenInto(&state, got, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *actionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := &mgmt.UpdateActionRequestContent{Name: plan.Name.ValueStringPointer()}
	expandActionInto(ctx, &plan, &body.SupportedTriggers, &body.Code, &body.Runtime, &body.Dependencies, &body.Secrets, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if _, err := r.mgmt.Actions.Update(ctx, plan.ID.ValueString(), body); err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to update action", err)
		return
	}

	// The Update endpoint does not deploy; do it explicitly when requested.
	if plan.Deploy.ValueBool() {
		if _, err := r.mgmt.Actions.Deploy(ctx, plan.ID.ValueString()); err != nil {
			framework.AddAPIError(&resp.Diagnostics, "Failed to deploy action", err)
			return
		}
	}

	got, err := r.mgmt.Actions.Get(ctx, plan.ID.ValueString())
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to read action after update", err)
		return
	}

	flattenInto(&plan, got, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *actionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.mgmt.Actions.Delete(ctx, state.ID.ValueString(), &mgmt.DeleteActionRequestParameters{}); err != nil && !framework.IsNotFound(err) {
		framework.AddAPIError(&resp.Diagnostics, "Failed to delete action", err)
	}
}

func (r *actionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
