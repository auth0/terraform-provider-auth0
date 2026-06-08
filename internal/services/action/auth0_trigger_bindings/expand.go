package auth0triggerbindings

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// expandBindings converts the planned `actions` list into the bulk
// UpdateActionBindings request body. Actions are referenced by their ID.
func expandBindings(ctx context.Context, list types.List) (*mgmt.UpdateActionBindingsRequestContent, diag.Diagnostics) {
	var diags diag.Diagnostics

	var elems []actionModel
	diags.Append(list.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil, diags
	}

	bindings := make([]*mgmt.ActionBindingWithRef, 0, len(elems))
	for i := range elems {
		ref := &mgmt.ActionBindingRef{
			Type:  mgmt.ActionBindingRefTypeEnumActionID.Ptr(),
			Value: elems[i].ID.ValueStringPointer(),
		}
		binding := &mgmt.ActionBindingWithRef{
			Ref:         ref,
			DisplayName: elems[i].DisplayName.ValueStringPointer(),
		}
		bindings = append(bindings, binding)
	}

	body := &mgmt.UpdateActionBindingsRequestContent{}
	// SetBindings marks the field explicit so an empty list is serialized as
	// `[]` (rather than omitted), which is required to unbind every action.
	body.SetBindings(bindings)
	return body, diags
}
