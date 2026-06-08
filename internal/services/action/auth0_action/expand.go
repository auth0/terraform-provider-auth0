package auth0action

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// -- nested attribute type maps --------------------------------------------

func triggerAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":      types.StringType,
		"version": types.StringType,
	}
}

func dependencyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"version": types.StringType,
	}
}

func secretAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	}
}

// -- plan helper models ----------------------------------------------------

type triggerModel struct {
	ID      types.String `tfsdk:"id"`
	Version types.String `tfsdk:"version"`
}

type dependencyModel struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
}

type secretModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// expandActionInto populates the shared body fields used by both the Create and
// Update requests from the planned model.
func expandActionInto(
	ctx context.Context,
	plan *model,
	triggers *[]*mgmt.ActionTrigger,
	code **string,
	runtime **string,
	deps *[]*mgmt.ActionVersionDependency,
	secrets *[]*mgmt.ActionSecretRequest,
	diags *diag.Diagnostics,
) {
	*triggers = expandTriggers(ctx, plan.SupportedTriggers, diags)
	if diags.HasError() {
		return
	}

	*code = plan.Code.ValueStringPointer()

	if !plan.Runtime.IsNull() && !plan.Runtime.IsUnknown() {
		*runtime = plan.Runtime.ValueStringPointer()
	}

	*deps = expandDependencies(ctx, plan.Dependencies, diags)
	if diags.HasError() {
		return
	}

	*secrets = expandSecrets(ctx, plan.Secrets, diags)
}

func expandTriggers(ctx context.Context, obj types.Object, diags *diag.Diagnostics) []*mgmt.ActionTrigger {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}
	var t triggerModel
	diags.Append(obj.As(ctx, &t, framework.ObjectAsOpts())...)
	if diags.HasError() {
		return nil
	}
	return []*mgmt.ActionTrigger{{
		ID:      mgmt.ActionTriggerTypeEnum(t.ID.ValueString()),
		Version: t.Version.ValueStringPointer(),
	}}
}

func expandDependencies(ctx context.Context, set types.Set, diags *diag.Diagnostics) []*mgmt.ActionVersionDependency {
	if set.IsNull() || set.IsUnknown() {
		return nil
	}
	var elems []dependencyModel
	diags.Append(set.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil
	}
	out := make([]*mgmt.ActionVersionDependency, 0, len(elems))
	for i := range elems {
		out = append(out, &mgmt.ActionVersionDependency{
			Name:    elems[i].Name.ValueStringPointer(),
			Version: elems[i].Version.ValueStringPointer(),
		})
	}
	return out
}

func expandSecrets(ctx context.Context, list types.List, diags *diag.Diagnostics) []*mgmt.ActionSecretRequest {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var elems []secretModel
	diags.Append(list.ElementsAs(ctx, &elems, false)...)
	if diags.HasError() {
		return nil
	}
	out := make([]*mgmt.ActionSecretRequest, 0, len(elems))
	for i := range elems {
		out = append(out, &mgmt.ActionSecretRequest{
			Name:  elems[i].Name.ValueStringPointer(),
			Value: elems[i].Value.ValueStringPointer(),
		})
	}
	return out
}
