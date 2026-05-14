package organization

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// =========================================================================
// flatten: SDK response -> Terraform model
// =========================================================================

// flattenInto copies the API response onto the model in place.
func flattenInto(
	_ context.Context,
	m *model,
	name *string,
	displayName *string,
	metadata *mgmt.OrganizationMetadata,
	branding *mgmt.OrganizationBranding,
	tokenQuota *mgmt.TokenQuota,
	diags *diag.Diagnostics,
) {
	m.Name = types.StringPointerValue(name)
	m.DisplayName = types.StringPointerValue(displayName)

	if metadata == nil {
		m.Metadata = types.MapNull(types.StringType)
	} else {
		els := make(map[string]attr.Value, len(*metadata))
		for k, v := range *metadata {
			els[k] = types.StringPointerValue(v)
		}
		mv, d := types.MapValue(types.StringType, els)
		diags.Append(d...)
		m.Metadata = mv
	}

	if branding == nil {
		m.Branding = types.ObjectNull(brandingAttrTypes())
	} else {
		colorsVal := types.ObjectNull(colorsAttrTypes())
		if branding.Colors != nil {
			cv, d := types.ObjectValue(colorsAttrTypes(), map[string]attr.Value{
				"primary":         types.StringValue(branding.Colors.Primary),
				"page_background": types.StringValue(branding.Colors.PageBackground),
			})
			diags.Append(d...)
			colorsVal = cv
		}
		bv, d := types.ObjectValue(brandingAttrTypes(), map[string]attr.Value{
			"logo_url": types.StringPointerValue(branding.LogoURL),
			"colors":   colorsVal,
		})
		diags.Append(d...)
		m.Branding = bv
	}

	m.TokenQuota = flattenTokenQuota(tokenQuota, diags)
}

func flattenTokenQuota(in *mgmt.TokenQuota, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(tokenQuotaAttrTypes())
	}
	ccVal := types.ObjectNull(tokenQuotaCCAttrTypes())
	if in.ClientCredentials != nil {
		cc := in.ClientCredentials
		v, d := types.ObjectValue(tokenQuotaCCAttrTypes(), map[string]attr.Value{
			"enforce":  types.BoolPointerValue(cc.Enforce),
			"per_day":  framework.IntPtrToInt64(cc.PerDay),
			"per_hour": framework.IntPtrToInt64(cc.PerHour),
		})
		diags.Append(d...)
		ccVal = v
	}
	out, d := types.ObjectValue(tokenQuotaAttrTypes(), map[string]attr.Value{
		"client_credentials": ccVal,
	})
	diags.Append(d...)
	return out
}
