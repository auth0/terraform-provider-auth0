package auth0triggerbindings

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"
	core "github.com/auth0/go-auth0/v2/management/core"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// actionElemAttrTypes describes a single element of the `actions` list.
func actionElemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":           types.StringType,
		"display_name": types.StringType,
	}
}

// collectBindings drains the paginated bindings response into a flat slice,
// preserving the order returned by Auth0 (which is the execution order).
func collectBindings(
	ctx context.Context,
	page *core.Page[*int, *mgmt.ActionBinding, *mgmt.ListActionBindingsPaginatedResponseContent],
	diags *diag.Diagnostics,
) []*mgmt.ActionBinding {
	var out []*mgmt.ActionBinding
	if page == nil {
		return out
	}
	iter := page.Iterator()
	for iter.Next(ctx) {
		out = append(out, iter.Current())
	}
	if err := iter.Err(); err != nil {
		diags.AddError("Failed to paginate trigger bindings", err.Error())
	}
	return out
}

// flattenInto projects the API bindings onto the resource state model.
func flattenInto(m *model, bindings []*mgmt.ActionBinding, diags *diag.Diagnostics) {
	m.Trigger = types.StringValue(m.Trigger.ValueString())

	elems := make([]attr.Value, 0, len(bindings))
	for _, b := range bindings {
		var actionID string
		if b.Action != nil {
			actionID = b.Action.GetID()
		}
		obj, d := types.ObjectValue(actionElemAttrTypes(), map[string]attr.Value{
			"id":           types.StringValue(actionID),
			"display_name": types.StringValue(b.GetDisplayName()),
		})
		diags.Append(d...)
		if diags.HasError() {
			return
		}
		elems = append(elems, obj)
	}

	list, d := types.ListValue(types.ObjectType{AttrTypes: actionElemAttrTypes()}, elems)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	m.Actions = list
}
