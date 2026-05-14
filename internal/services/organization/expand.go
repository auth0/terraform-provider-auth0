package organization

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// basetypesObjectAsOpts is kept as a thin local alias of framework.ObjectAsOpts
// so older call sites don't need touching. New code should prefer
// `framework.ObjectAsOpts()`.
func basetypesObjectAsOpts() basetypes.ObjectAsOptions { return framework.ObjectAsOpts() }

// =========================================================================
// expand: Terraform plan -> SDK request body
// =========================================================================

// expandMetadata converts the Terraform map into the SDK's
// `map[string]*string`. Returns nil when the value is null/unknown so that the
// existing metadata is left untouched on update.
func expandMetadata(ctx context.Context, m types.Map) (mgmt.OrganizationMetadata, diag.Diagnostics) {
	if m.IsNull() || m.IsUnknown() {
		return nil, nil
	}
	raw := make(map[string]string, len(m.Elements()))
	diags := m.ElementsAs(ctx, &raw, false)
	if diags.HasError() {
		return nil, diags
	}
	out := make(mgmt.OrganizationMetadata, len(raw))
	for k, v := range raw {
		v := v
		out[k] = &v
	}
	return out, nil
}

func expandBranding(ctx context.Context, o types.Object) (*mgmt.OrganizationBranding, diag.Diagnostics) {
	if o.IsNull() || o.IsUnknown() {
		return nil, nil
	}
	type brandingModel struct {
		LogoURL types.String `tfsdk:"logo_url"`
		Colors  types.Object `tfsdk:"colors"`
	}
	type colorsModel struct {
		Primary        types.String `tfsdk:"primary"`
		PageBackground types.String `tfsdk:"page_background"`
	}

	var b brandingModel
	if d := o.As(ctx, &b, basetypesObjectAsOpts()); d.HasError() {
		return nil, d
	}
	out := &mgmt.OrganizationBranding{}
	if !b.LogoURL.IsNull() && !b.LogoURL.IsUnknown() {
		out.LogoURL = b.LogoURL.ValueStringPointer()
	}
	if !b.Colors.IsNull() && !b.Colors.IsUnknown() {
		var c colorsModel
		if d := b.Colors.As(ctx, &c, basetypesObjectAsOpts()); d.HasError() {
			return nil, d
		}
		out.Colors = &mgmt.OrganizationBrandingColors{
			Primary:        c.Primary.ValueString(),
			PageBackground: c.PageBackground.ValueString(),
		}
	}
	return out, nil
}

// -- token_quota -----------------------------------------------------------

// tokenQuotaCCFromTF reads the typed nested object into the SDK type.
func tokenQuotaCCFromTF(o types.Object) *mgmt.TokenQuotaClientCredentials {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	attrs := o.Attributes()
	cc := &mgmt.TokenQuotaClientCredentials{}
	if v, ok := attrs["enforce"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		cc.Enforce = v.ValueBoolPointer()
	}
	if v, ok := attrs["per_day"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		cc.PerDay = framework.Int64ToIntPtr(v)
	}
	if v, ok := attrs["per_hour"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		cc.PerHour = framework.Int64ToIntPtr(v)
	}
	return cc
}

func expandTokenQuotaCreate(_ context.Context, o types.Object) (*mgmt.CreateTokenQuota, diag.Diagnostics) {
	if o.IsNull() || o.IsUnknown() {
		return nil, nil
	}
	cc := nestedObject(o, "client_credentials")
	return &mgmt.CreateTokenQuota{ClientCredentials: tokenQuotaCCFromTF(cc)}, nil
}

func expandTokenQuotaUpdate(_ context.Context, o types.Object) (*mgmt.UpdateTokenQuota, diag.Diagnostics) {
	if o.IsNull() || o.IsUnknown() {
		return nil, nil
	}
	cc := nestedObject(o, "client_credentials")
	return &mgmt.UpdateTokenQuota{ClientCredentials: tokenQuotaCCFromTF(cc)}, nil
}

// nestedObject pulls a child types.Object by name (returns a null Object if
// the child is missing or of the wrong type).
func nestedObject(parent types.Object, name string) types.Object {
	v, ok := parent.Attributes()[name]
	if !ok {
		return types.ObjectNull(nil)
	}
	o, ok := v.(types.Object)
	if !ok {
		return types.ObjectNull(nil)
	}
	return o
}

// =========================================================================
// enabled_connections (write-only — see resource.go schema docs)
// =========================================================================

// expandEnabledConnections converts the typed list into the SDK type. Returns
// nil for null/unknown so the field is omitted from the request.
func expandEnabledConnections(l types.List) []*mgmt.ConnectionForOrganization {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]*mgmt.ConnectionForOrganization, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		obj, ok := e.(types.Object)
		if !ok {
			continue
		}
		a := obj.Attributes()
		conn := &mgmt.ConnectionForOrganization{}
		if v, ok := a["connection_id"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
			conn.ConnectionID = v.ValueString()
		}
		if v, ok := a["assign_membership_on_login"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
			conn.AssignMembershipOnLogin = v.ValueBoolPointer()
		}
		if v, ok := a["show_as_button"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
			conn.ShowAsButton = v.ValueBoolPointer()
		}
		if v, ok := a["is_signup_enabled"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
			conn.IsSignupEnabled = v.ValueBoolPointer()
		}
		out = append(out, conn)
	}
	return out
}
