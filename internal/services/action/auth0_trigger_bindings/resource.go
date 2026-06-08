// Package auth0triggerbindings implements the auth0_trigger_bindings resource.
//
// Auth0 exposes only a *bulk* API for trigger bindings:
//
//	GET   /api/v2/actions/triggers/{trigger}/bindings   -> list the ordered bindings
//	PATCH /api/v2/actions/triggers/{trigger}/bindings   -> replace the entire ordered list
//
// There is no endpoint to create or delete a single binding. Because of that
// this resource is *authoritative*: it owns the complete, ordered list of
// actions bound to one trigger. The order of `actions` in HCL determines the
// execution order during the trigger's flow.
//
// The SDKv2 provider additionally shipped a non-authoritative, per-binding
// resource (auth0_trigger_action) implemented via read-modify-write on the
// shared list. That pattern is race-prone under concurrent applies, so the
// framework-based provider intentionally exposes only the authoritative
// resource here.
package auth0triggerbindings

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"
	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

var (
	_ resource.Resource                = (*triggerBindingsResource)(nil)
	_ resource.ResourceWithConfigure   = (*triggerBindingsResource)(nil)
	_ resource.ResourceWithImportState = (*triggerBindingsResource)(nil)
)

// NewResource returns a fresh auth0_trigger_bindings resource.
func NewResource() resource.Resource { return &triggerBindingsResource{} }

type triggerBindingsResource struct {
	mgmt *mgmtclient.Management
}

// model mirrors the auth0_trigger_bindings HCL schema.
type model struct {
	ID      types.String `tfsdk:"id"`
	Trigger types.String `tfsdk:"trigger"`
	Actions types.List   `tfsdk:"actions"`
}

// actionModel mirrors a single element of the ordered `actions` list.
type actionModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
}

func (r *triggerBindingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_trigger_bindings"
}

func (r *triggerBindingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the complete, ordered set of actions bound to an Auth0 trigger. " +
			"This resource is authoritative: it owns every binding for the given trigger, so any " +
			"binding not declared here will be removed. See " +
			"https://auth0.com/docs/customize/actions/flows-and-triggers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "The trigger identifier. Equal to the `trigger` value.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"trigger": schema.StringAttribute{
				Required: true,
				Description: "The trigger to bind the actions to (e.g. `post-login`, " +
					"`credentials-exchange`, `pre-user-registration`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators:    []validator.String{triggerValidator{}},
			},
			"actions": schema.ListNestedAttribute{
				Required: true,
				Description: "The ordered list of actions bound to the trigger. The order " +
					"determines the execution order during the trigger's flow.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the action to bind to the trigger.",
						},
						"display_name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the binding as shown in the dashboard.",
						},
					},
				},
			},
		},
	}
}

func (r *triggerBindingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if m, ok := framework.ManagementFromResource(req, resp); ok {
		r.mgmt = m
	}
}

// -- CRUD ------------------------------------------------------------------

func (r *triggerBindingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.upsert(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *triggerBindingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	trigger := mgmt.ActionTriggerTypeEnum(state.Trigger.ValueString())
	page, err := r.mgmt.Actions.Triggers.Bindings.List(ctx, &trigger, &mgmt.ListActionTriggerBindingsRequestParameters{})
	if err != nil {
		if framework.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		framework.AddAPIError(&resp.Diagnostics, "Failed to read trigger bindings", err)
		return
	}

	bindings := collectBindings(ctx, page, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ID = types.StringValue(state.Trigger.ValueString())
	flattenInto(&state, bindings, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *triggerBindingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if r.upsert(ctx, &plan, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *triggerBindingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	trigger := mgmt.ActionTriggerTypeEnum(state.Trigger.ValueString())
	body := &mgmt.UpdateActionBindingsRequestContent{}
	body.SetBindings([]*mgmt.ActionBindingWithRef{}) // explicit empty list -> unbind everything
	if _, err := r.mgmt.Actions.Triggers.Bindings.UpdateMany(ctx, &trigger, body); err != nil && !framework.IsNotFound(err) {
		framework.AddAPIError(&resp.Diagnostics, "Failed to delete trigger bindings", err)
	}
}

func (r *triggerBindingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by the trigger id, e.g. `terraform import auth0_trigger_bindings.example post-login`.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("trigger"), req.ID)...)
}

// upsert is shared by Create and Update: it replaces the whole binding list for
// the trigger, then re-reads it so state reflects exactly what the API stored.
// It returns true when the plan was populated successfully and state should be
// written by the caller.
func (r *triggerBindingsResource) upsert(ctx context.Context, plan *model, diags *diag.Diagnostics) bool {
	trigger := mgmt.ActionTriggerTypeEnum(plan.Trigger.ValueString())

	body, d := expandBindings(ctx, plan.Actions)
	diags.Append(d...)
	if diags.HasError() {
		return false
	}

	if _, err := r.mgmt.Actions.Triggers.Bindings.UpdateMany(ctx, &trigger, body); err != nil {
		framework.AddAPIError(diags, "Failed to set trigger bindings", err)
		return false
	}

	page, err := r.mgmt.Actions.Triggers.Bindings.List(ctx, &trigger, &mgmt.ListActionTriggerBindingsRequestParameters{})
	if err != nil {
		framework.AddAPIError(diags, "Failed to read trigger bindings after update", err)
		return false
	}
	bindings := collectBindings(ctx, page, diags)
	if diags.HasError() {
		return false
	}

	plan.ID = types.StringValue(plan.Trigger.ValueString())
	flattenInto(plan, bindings, diags)
	return !diags.HasError()
}
