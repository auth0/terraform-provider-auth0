package auth0action

import (
	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// flattenInto projects an Action API response onto the resource state model.
//
// The secrets list is deliberately left untouched: the API never returns secret
// values, so the value already present on `m` (from plan or prior state) is the
// authoritative source and is carried forward to avoid permanent drift.
func flattenInto(m *model, a *mgmt.GetActionResponseContent, diags *diag.Diagnostics) {
	m.ID = types.StringValue(a.GetID())
	m.Name = types.StringValue(a.GetName())
	m.Code = types.StringValue(a.GetCode())
	m.Runtime = types.StringValue(a.GetRuntime())

	m.SupportedTriggers = flattenTriggers(a.GetSupportedTriggers(), diags)
	if diags.HasError() {
		return
	}

	m.Dependencies = flattenDependencies(a.GetDependencies(), diags)
	if diags.HasError() {
		return
	}

	if dv := a.DeployedVersion; dv != nil {
		m.VersionID = types.StringValue(dv.GetID())
	} else {
		m.VersionID = types.StringValue("")
	}

	// Normalise an unknown/null deploy flag so state always has a concrete value.
	if m.Deploy.IsNull() || m.Deploy.IsUnknown() {
		m.Deploy = types.BoolValue(false)
	}

	// Ensure secrets is never unknown in state.
	if m.Secrets.IsUnknown() {
		m.Secrets = types.ListNull(types.ObjectType{AttrTypes: secretAttrTypes()})
	}
}

func flattenTriggers(triggers []*mgmt.ActionTrigger, diags *diag.Diagnostics) types.Object {
	if len(triggers) == 0 {
		return types.ObjectNull(triggerAttrTypes())
	}
	t := triggers[0]
	obj, d := types.ObjectValue(triggerAttrTypes(), map[string]attr.Value{
		"id":      types.StringValue(string(t.GetID())),
		"version": types.StringValue(t.GetVersion()),
	})
	diags.Append(d...)
	return obj
}

func flattenDependencies(deps []*mgmt.ActionVersionDependency, diags *diag.Diagnostics) types.Set {
	elemType := types.ObjectType{AttrTypes: dependencyAttrTypes()}
	if len(deps) == 0 {
		return types.SetNull(elemType)
	}
	elems := make([]attr.Value, 0, len(deps))
	for _, d := range deps {
		obj, dd := types.ObjectValue(dependencyAttrTypes(), map[string]attr.Value{
			"name":    types.StringValue(d.GetName()),
			"version": types.StringValue(d.GetVersion()),
		})
		diags.Append(dd...)
		if diags.HasError() {
			return types.SetNull(elemType)
		}
		elems = append(elems, obj)
	}
	set, d := types.SetValue(elemType, elems)
	diags.Append(d...)
	return set
}

// flattenSecretNames returns the list of secret names exposed by the API. Secret
// values are never returned, so only names are surfaced (used by the data source).
func flattenSecretNames(secrets []*mgmt.ActionSecretResponse, diags *diag.Diagnostics) types.List {
	if len(secrets) == 0 {
		return types.ListNull(types.StringType)
	}
	elems := make([]attr.Value, 0, len(secrets))
	for _, s := range secrets {
		elems = append(elems, types.StringValue(s.GetName()))
	}
	list, d := types.ListValue(types.StringType, elems)
	diags.Append(d...)
	return list
}
