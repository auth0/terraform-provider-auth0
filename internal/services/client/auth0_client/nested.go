package auth0client

import (
	"context"
	"encoding/json"
	"time"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// =========================================================================
// jwt_configuration
// =========================================================================

func jwtConfigurationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"lifetime_in_seconds": types.Int64Type,
		"secret_encoded":      types.BoolType,
		"alg":                 types.StringType,
		// scopes is the SDK's `map[string]any`, stored as a JSON string.
		"scopes": types.StringType,
	}
}

func expandJwtConfiguration(_ context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientJwtConfiguration {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientJwtConfiguration{}
	if v, ok := a["lifetime_in_seconds"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		out.LifetimeInSeconds = framework.Int64ToIntPtr(v)
	}
	if v, ok := a["secret_encoded"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		out.SecretEncoded = v.ValueBoolPointer()
	}
	if v, ok := a["alg"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewSigningAlgorithmEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("jwt_configuration").AtName("alg"), "Invalid alg", err.Error())
			return nil
		}
		out.Alg = &enum
	}
	if v, ok := a["scopes"].(types.String); ok {
		if parsed, present := framework.ParseJSONString(v, "jwt_configuration.scopes", diags); present {
			scopes := mgmt.ClientJwtConfigurationScopes{}
			if m, ok := parsed.(map[string]any); ok {
				scopes = m
			}
			out.Scopes = &scopes
		}
	}
	return out
}

func flattenJwtConfiguration(in *mgmt.ClientJwtConfiguration, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(jwtConfigurationAttrTypes())
	}
	scopesStr := types.StringNull()
	if in.Scopes != nil {
		scopesStr = framework.FlattenJSONToString(map[string]any(*in.Scopes), diags)
	}
	algStr := framework.EnumPtrToString(in.Alg)
	v, d := types.ObjectValue(jwtConfigurationAttrTypes(), map[string]attr.Value{
		"lifetime_in_seconds": framework.IntPtrToInt64(in.LifetimeInSeconds),
		"secret_encoded":      types.BoolPointerValue(in.SecretEncoded),
		"alg":                 algStr,
		"scopes":              scopesStr,
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// refresh_token (with policies list)
// =========================================================================

func refreshTokenPolicyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"audience": types.StringType,
		"scope":    types.ListType{ElemType: types.StringType},
	}
}

func refreshTokenAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"rotation_type":                types.StringType,
		"expiration_type":              types.StringType,
		"leeway":                       types.Int64Type,
		"token_lifetime":               types.Int64Type,
		"infinite_token_lifetime":      types.BoolType,
		"idle_token_lifetime":          types.Int64Type,
		"infinite_idle_token_lifetime": types.BoolType,
		"policies":                     types.ListType{ElemType: types.ObjectType{AttrTypes: refreshTokenPolicyAttrTypes()}},
	}
}

func expandRefreshToken(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientRefreshTokenConfiguration {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientRefreshTokenConfiguration{}
	if v, ok := a["rotation_type"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewRefreshTokenRotationTypeEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("refresh_token").AtName("rotation_type"), "Invalid rotation_type", err.Error())
			return nil
		}
		out.RotationType = enum
	}
	if v, ok := a["expiration_type"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewRefreshTokenExpirationTypeEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("refresh_token").AtName("expiration_type"), "Invalid expiration_type", err.Error())
			return nil
		}
		out.ExpirationType = enum
	}
	if v, ok := a["leeway"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		out.Leeway = framework.Int64ToIntPtr(v)
	}
	if v, ok := a["token_lifetime"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		out.TokenLifetime = framework.Int64ToIntPtr(v)
	}
	if v, ok := a["infinite_token_lifetime"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		out.InfiniteTokenLifetime = v.ValueBoolPointer()
	}
	if v, ok := a["idle_token_lifetime"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		out.IdleTokenLifetime = framework.Int64ToIntPtr(v)
	}
	if v, ok := a["infinite_idle_token_lifetime"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		out.InfiniteIdleTokenLifetime = v.ValueBoolPointer()
	}
	if v, ok := a["policies"].(types.List); ok && !v.IsNull() && !v.IsUnknown() {
		out.Policies = expandRefreshTokenPolicies(ctx, v, diags)
	}
	return out
}

func expandRefreshTokenPolicies(ctx context.Context, l types.List, diags *diag.Diagnostics) []*mgmt.ClientRefreshTokenPolicy {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]*mgmt.ClientRefreshTokenPolicy, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		obj, ok := e.(types.Object)
		if !ok {
			continue
		}
		a := obj.Attributes()
		p := &mgmt.ClientRefreshTokenPolicy{}
		if v, ok := a["audience"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
			p.Audience = v.ValueString()
		}
		if v, ok := a["scope"].(types.List); ok && !v.IsNull() && !v.IsUnknown() {
			scope := make([]string, 0, len(v.Elements()))
			diags.Append(v.ElementsAs(ctx, &scope, false)...)
			p.Scope = scope
		}
		out = append(out, p)
	}
	return out
}

func flattenRefreshToken(in *mgmt.ClientRefreshTokenConfiguration, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(refreshTokenAttrTypes())
	}
	policies := flattenRefreshTokenPolicies(in.Policies, diags)
	v, d := types.ObjectValue(refreshTokenAttrTypes(), map[string]attr.Value{
		"rotation_type":                types.StringValue(string(in.RotationType)),
		"expiration_type":              types.StringValue(string(in.ExpirationType)),
		"leeway":                       framework.IntPtrToInt64(in.Leeway),
		"token_lifetime":               framework.IntPtrToInt64(in.TokenLifetime),
		"infinite_token_lifetime":      types.BoolPointerValue(in.InfiniteTokenLifetime),
		"idle_token_lifetime":          framework.IntPtrToInt64(in.IdleTokenLifetime),
		"infinite_idle_token_lifetime": types.BoolPointerValue(in.InfiniteIdleTokenLifetime),
		"policies":                     policies,
	})
	diags.Append(d...)
	return v
}

func flattenRefreshTokenPolicies(in []*mgmt.ClientRefreshTokenPolicy, diags *diag.Diagnostics) types.List {
	objType := types.ObjectType{AttrTypes: refreshTokenPolicyAttrTypes()}
	if in == nil {
		return types.ListNull(objType)
	}
	vals := make([]attr.Value, 0, len(in))
	for _, p := range in {
		if p == nil {
			continue
		}
		v, d := types.ObjectValue(refreshTokenPolicyAttrTypes(), map[string]attr.Value{
			"audience": types.StringValue(p.Audience),
			"scope":    framework.StringSliceToList(p.Scope),
		})
		diags.Append(d...)
		vals = append(vals, v)
	}
	lv, d := types.ListValue(objType, vals)
	diags.Append(d...)
	return lv
}

// =========================================================================
// oidc_logout / oidc_backchannel_logout
// =========================================================================

func oidcLogoutInitiatorsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"mode":                types.StringType,
		"selected_initiators": types.ListType{ElemType: types.StringType},
	}
}

func oidcLogoutSessionMetadataAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{"include": types.BoolType}
}

func oidcLogoutAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"backchannel_logout_urls":             types.ListType{ElemType: types.StringType},
		"backchannel_logout_initiators":       types.ObjectType{AttrTypes: oidcLogoutInitiatorsAttrTypes()},
		"backchannel_logout_session_metadata": types.ObjectType{AttrTypes: oidcLogoutSessionMetadataAttrTypes()},
	}
}

func expandOidcLogout(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientOidcBackchannelLogoutSettings {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientOidcBackchannelLogoutSettings{}
	if v, ok := a["backchannel_logout_urls"].(types.List); ok && !v.IsNull() && !v.IsUnknown() {
		urls := make([]string, 0, len(v.Elements()))
		diags.Append(v.ElementsAs(ctx, &urls, false)...)
		out.BackchannelLogoutURLs = urls
	}
	if v, ok := a["backchannel_logout_initiators"].(types.Object); ok && !v.IsNull() && !v.IsUnknown() {
		ia := v.Attributes()
		ini := &mgmt.ClientOidcBackchannelLogoutInitiators{}
		if mv, ok := ia["mode"].(types.String); ok && !mv.IsNull() && !mv.IsUnknown() {
			enum, err := mgmt.NewClientOidcBackchannelLogoutInitiatorsModeEnumFromString(mv.ValueString())
			if err != nil {
				diags.AddAttributeError(path.Root("oidc_logout").AtName("backchannel_logout_initiators").AtName("mode"), "Invalid mode", err.Error())
				return nil
			}
			ini.Mode = &enum
		}
		if sv, ok := ia["selected_initiators"].(types.List); ok && !sv.IsNull() && !sv.IsUnknown() {
			raw := make([]string, 0, len(sv.Elements()))
			diags.Append(sv.ElementsAs(ctx, &raw, false)...)
			out2 := make([]mgmt.ClientOidcBackchannelLogoutInitiatorsEnum, 0, len(raw))
			for _, s := range raw {
				enum, err := mgmt.NewClientOidcBackchannelLogoutInitiatorsEnumFromString(s)
				if err != nil {
					diags.AddAttributeError(path.Root("oidc_logout").AtName("backchannel_logout_initiators").AtName("selected_initiators"), "Invalid initiator", err.Error())
					return nil
				}
				out2 = append(out2, enum)
			}
			ini.SelectedInitiators = out2
		}
		out.BackchannelLogoutInitiators = ini
	}
	if v, ok := a["backchannel_logout_session_metadata"].(types.Object); ok && !v.IsNull() && !v.IsUnknown() {
		ma := v.Attributes()
		meta := &mgmt.ClientOidcBackchannelLogoutSessionMetadata{}
		if iv, ok := ma["include"].(types.Bool); ok && !iv.IsNull() && !iv.IsUnknown() {
			meta.Include = iv.ValueBoolPointer()
		}
		out.BackchannelLogoutSessionMetadata = meta
	}
	return out
}

func flattenOidcLogout(in *mgmt.ClientOidcBackchannelLogoutSettings, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(oidcLogoutAttrTypes())
	}
	initiators := types.ObjectNull(oidcLogoutInitiatorsAttrTypes())
	if in.BackchannelLogoutInitiators != nil {
		v, d := types.ObjectValue(oidcLogoutInitiatorsAttrTypes(), map[string]attr.Value{
			"mode":                framework.EnumPtrToString(in.BackchannelLogoutInitiators.Mode),
			"selected_initiators": framework.EnumSliceToList(in.BackchannelLogoutInitiators.SelectedInitiators),
		})
		diags.Append(d...)
		initiators = v
	}
	sessionMeta := types.ObjectNull(oidcLogoutSessionMetadataAttrTypes())
	if in.BackchannelLogoutSessionMetadata != nil {
		v, d := types.ObjectValue(oidcLogoutSessionMetadataAttrTypes(), map[string]attr.Value{
			"include": types.BoolPointerValue(in.BackchannelLogoutSessionMetadata.Include),
		})
		diags.Append(d...)
		sessionMeta = v
	}
	v, d := types.ObjectValue(oidcLogoutAttrTypes(), map[string]attr.Value{
		"backchannel_logout_urls":             framework.StringSliceToList(in.BackchannelLogoutURLs),
		"backchannel_logout_initiators":       initiators,
		"backchannel_logout_session_metadata": sessionMeta,
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// encryption_key
// =========================================================================

func encryptionKeyAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"pub":     types.StringType,
		"cert":    types.StringType,
		"subject": types.StringType,
	}
}

func expandEncryptionKey(_ context.Context, o types.Object) *mgmt.ClientEncryptionKey {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientEncryptionKey{}
	if v, ok := a["pub"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		out.Pub = v.ValueStringPointer()
	}
	if v, ok := a["cert"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		out.Cert = v.ValueStringPointer()
	}
	if v, ok := a["subject"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		out.Subject = v.ValueStringPointer()
	}
	return out
}

func flattenEncryptionKey(in *mgmt.ClientEncryptionKey, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(encryptionKeyAttrTypes())
	}
	v, d := types.ObjectValue(encryptionKeyAttrTypes(), map[string]attr.Value{
		"pub":     types.StringPointerValue(in.Pub),
		"cert":    types.StringPointerValue(in.Cert),
		"subject": types.StringPointerValue(in.Subject),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// default_organization
// =========================================================================

func defaultOrganizationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"organization_id": types.StringType,
		"flows":           types.ListType{ElemType: types.StringType},
	}
}

func expandDefaultOrganization(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientDefaultOrganization {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientDefaultOrganization{}
	if v, ok := a["organization_id"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		out.OrganizationID = v.ValueString()
	}
	if v, ok := a["flows"].(types.List); ok && !v.IsNull() && !v.IsUnknown() {
		raw := make([]string, 0, len(v.Elements()))
		diags.Append(v.ElementsAs(ctx, &raw, false)...)
		flows := make([]mgmt.ClientDefaultOrganizationFlowsEnum, 0, len(raw))
		for _, s := range raw {
			enum, err := mgmt.NewClientDefaultOrganizationFlowsEnumFromString(s)
			if err != nil {
				diags.AddAttributeError(path.Root("default_organization").AtName("flows"), "Invalid flow", err.Error())
				return nil
			}
			flows = append(flows, enum)
		}
		out.Flows = flows
	}
	return out
}

func flattenDefaultOrganization(in *mgmt.ClientDefaultOrganization, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(defaultOrganizationAttrTypes())
	}
	v, d := types.ObjectValue(defaultOrganizationAttrTypes(), map[string]attr.Value{
		"organization_id": types.StringValue(in.OrganizationID),
		"flows":           framework.EnumSliceToList(in.Flows),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// native_social_login
// =========================================================================

func nativeSocialLoginProviderAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{"enabled": types.BoolType}
}

func nativeSocialLoginAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"apple":    types.ObjectType{AttrTypes: nativeSocialLoginProviderAttrTypes()},
		"facebook": types.ObjectType{AttrTypes: nativeSocialLoginProviderAttrTypes()},
		"google":   types.ObjectType{AttrTypes: nativeSocialLoginProviderAttrTypes()},
	}
}

func extractNSLProviderEnabled(o types.Object) *bool {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	v, ok := o.Attributes()["enabled"].(types.Bool)
	if !ok || v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueBoolPointer()
}

func expandNativeSocialLogin(_ context.Context, o types.Object) *mgmt.NativeSocialLogin {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.NativeSocialLogin{}
	if obj, ok := a["apple"].(types.Object); ok {
		if e := extractNSLProviderEnabled(obj); e != nil {
			out.Apple = &mgmt.NativeSocialLoginApple{Enabled: e}
		}
	}
	if obj, ok := a["facebook"].(types.Object); ok {
		if e := extractNSLProviderEnabled(obj); e != nil {
			out.Facebook = &mgmt.NativeSocialLoginFacebook{Enabled: e}
		}
	}
	if obj, ok := a["google"].(types.Object); ok {
		if e := extractNSLProviderEnabled(obj); e != nil {
			out.Google = &mgmt.NativeSocialLoginGoogle{Enabled: e}
		}
	}
	return out
}

func flattenNSLProvider(enabled *bool, diags *diag.Diagnostics) types.Object {
	if enabled == nil {
		return types.ObjectNull(nativeSocialLoginProviderAttrTypes())
	}
	v, d := types.ObjectValue(nativeSocialLoginProviderAttrTypes(), map[string]attr.Value{
		"enabled": types.BoolPointerValue(enabled),
	})
	diags.Append(d...)
	return v
}

func flattenNativeSocialLogin(in *mgmt.NativeSocialLogin, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(nativeSocialLoginAttrTypes())
	}
	var apple, fb, google *bool
	if in.Apple != nil {
		apple = in.Apple.Enabled
	}
	if in.Facebook != nil {
		fb = in.Facebook.Enabled
	}
	if in.Google != nil {
		google = in.Google.Enabled
	}
	v, d := types.ObjectValue(nativeSocialLoginAttrTypes(), map[string]attr.Value{
		"apple":    flattenNSLProvider(apple, diags),
		"facebook": flattenNSLProvider(fb, diags),
		"google":   flattenNSLProvider(google, diags),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// session_transfer
// =========================================================================

func sessionTransferDelegationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allow_delegated_access": types.BoolType,
		"enforce_device_binding": types.StringType,
	}
}

func sessionTransferAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"can_create_session_transfer_token": types.BoolType,
		"enforce_cascade_revocation":        types.BoolType,
		"allowed_authentication_methods":    types.ListType{ElemType: types.StringType},
		"enforce_device_binding":            types.StringType,
		"allow_refresh_token":               types.BoolType,
		"enforce_online_refresh_tokens":     types.BoolType,
		"delegation":                        types.ObjectType{AttrTypes: sessionTransferDelegationAttrTypes()},
	}
}

func expandSessionTransfer(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientSessionTransferConfiguration {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientSessionTransferConfiguration{}
	if v, ok := a["can_create_session_transfer_token"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		out.CanCreateSessionTransferToken = v.ValueBoolPointer()
	}
	if v, ok := a["enforce_cascade_revocation"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		out.EnforceCascadeRevocation = v.ValueBoolPointer()
	}
	if v, ok := a["allowed_authentication_methods"].(types.List); ok && !v.IsNull() && !v.IsUnknown() {
		raw := make([]string, 0, len(v.Elements()))
		diags.Append(v.ElementsAs(ctx, &raw, false)...)
		out2 := make([]mgmt.ClientSessionTransferAllowedAuthenticationMethodsEnum, 0, len(raw))
		for _, s := range raw {
			enum, err := mgmt.NewClientSessionTransferAllowedAuthenticationMethodsEnumFromString(s)
			if err != nil {
				diags.AddAttributeError(path.Root("session_transfer").AtName("allowed_authentication_methods"), "Invalid value", err.Error())
				return nil
			}
			out2 = append(out2, enum)
		}
		out.AllowedAuthenticationMethods = out2
	}
	if v, ok := a["enforce_device_binding"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientSessionTransferDeviceBindingEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("session_transfer").AtName("enforce_device_binding"), "Invalid value", err.Error())
			return nil
		}
		out.EnforceDeviceBinding = &enum
	}
	if v, ok := a["allow_refresh_token"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		out.AllowRefreshToken = v.ValueBoolPointer()
	}
	if v, ok := a["enforce_online_refresh_tokens"].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		out.EnforceOnlineRefreshTokens = v.ValueBoolPointer()
	}
	if v, ok := a["delegation"].(types.Object); ok && !v.IsNull() && !v.IsUnknown() {
		da := v.Attributes()
		del := &mgmt.ClientSessionTransferDelegationConfiguration{}
		if dv, ok := da["allow_delegated_access"].(types.Bool); ok && !dv.IsNull() && !dv.IsUnknown() {
			del.AllowDelegatedAccess = dv.ValueBoolPointer()
		}
		if dv, ok := da["enforce_device_binding"].(types.String); ok && !dv.IsNull() && !dv.IsUnknown() {
			enum, err := mgmt.NewClientSessionTransferDelegationDeviceBindingEnumFromString(dv.ValueString())
			if err != nil {
				diags.AddAttributeError(path.Root("session_transfer").AtName("delegation").AtName("enforce_device_binding"), "Invalid value", err.Error())
				return nil
			}
			del.EnforceDeviceBinding = &enum
		}
		out.Delegation = del
	}
	return out
}

func flattenSessionTransfer(in *mgmt.ClientSessionTransferConfiguration, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(sessionTransferAttrTypes())
	}
	deleg := types.ObjectNull(sessionTransferDelegationAttrTypes())
	if in.Delegation != nil {
		v, d := types.ObjectValue(sessionTransferDelegationAttrTypes(), map[string]attr.Value{
			"allow_delegated_access": types.BoolPointerValue(in.Delegation.AllowDelegatedAccess),
			"enforce_device_binding": framework.EnumPtrToString(in.Delegation.EnforceDeviceBinding),
		})
		diags.Append(d...)
		deleg = v
	}
	v, d := types.ObjectValue(sessionTransferAttrTypes(), map[string]attr.Value{
		"can_create_session_transfer_token": types.BoolPointerValue(in.CanCreateSessionTransferToken),
		"enforce_cascade_revocation":        types.BoolPointerValue(in.EnforceCascadeRevocation),
		"allowed_authentication_methods":    framework.EnumSliceToList(in.AllowedAuthenticationMethods),
		"enforce_device_binding":            framework.EnumPtrToString(in.EnforceDeviceBinding),
		"allow_refresh_token":               types.BoolPointerValue(in.AllowRefreshToken),
		"enforce_online_refresh_tokens":     types.BoolPointerValue(in.EnforceOnlineRefreshTokens),
		"delegation":                        deleg,
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// mobile
// =========================================================================

func mobileAndroidAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"app_package_name":         types.StringType,
		"sha256_cert_fingerprints": types.ListType{ElemType: types.StringType},
	}
}

func mobileIosAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"team_id":               types.StringType,
		"app_bundle_identifier": types.StringType,
	}
}

func mobileAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"android": types.ObjectType{AttrTypes: mobileAndroidAttrTypes()},
		"ios":     types.ObjectType{AttrTypes: mobileIosAttrTypes()},
	}
}

func expandMobile(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientMobile {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientMobile{}
	if v, ok := a["android"].(types.Object); ok && !v.IsNull() && !v.IsUnknown() {
		aa := v.Attributes()
		droid := &mgmt.ClientMobileAndroid{}
		if pv, ok := aa["app_package_name"].(types.String); ok && !pv.IsNull() && !pv.IsUnknown() {
			droid.AppPackageName = pv.ValueStringPointer()
		}
		if pv, ok := aa["sha256_cert_fingerprints"].(types.List); ok && !pv.IsNull() && !pv.IsUnknown() {
			fps := make([]string, 0, len(pv.Elements()))
			diags.Append(pv.ElementsAs(ctx, &fps, false)...)
			droid.Sha256CertFingerprints = fps
		}
		out.Android = droid
	}
	if v, ok := a["ios"].(types.Object); ok && !v.IsNull() && !v.IsUnknown() {
		ia := v.Attributes()
		ios := &mgmt.ClientMobileiOs{}
		if pv, ok := ia["team_id"].(types.String); ok && !pv.IsNull() && !pv.IsUnknown() {
			ios.TeamID = pv.ValueStringPointer()
		}
		if pv, ok := ia["app_bundle_identifier"].(types.String); ok && !pv.IsNull() && !pv.IsUnknown() {
			ios.AppBundleIdentifier = pv.ValueStringPointer()
		}
		out.Ios = ios
	}
	return out
}

func flattenMobile(in *mgmt.ClientMobile, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(mobileAttrTypes())
	}
	droid := types.ObjectNull(mobileAndroidAttrTypes())
	if in.Android != nil {
		v, d := types.ObjectValue(mobileAndroidAttrTypes(), map[string]attr.Value{
			"app_package_name":         types.StringPointerValue(in.Android.AppPackageName),
			"sha256_cert_fingerprints": framework.StringSliceToList(in.Android.Sha256CertFingerprints),
		})
		diags.Append(d...)
		droid = v
	}
	ios := types.ObjectNull(mobileIosAttrTypes())
	if in.Ios != nil {
		v, d := types.ObjectValue(mobileIosAttrTypes(), map[string]attr.Value{
			"team_id":               types.StringPointerValue(in.Ios.TeamID),
			"app_bundle_identifier": types.StringPointerValue(in.Ios.AppBundleIdentifier),
		})
		diags.Append(d...)
		ios = v
	}
	v, d := types.ObjectValue(mobileAttrTypes(), map[string]attr.Value{
		"android": droid, "ios": ios,
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// token_exchange (Create + Update use different SDK types but identical fields)
// =========================================================================

func tokenExchangeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"allow_any_profile_of_type": types.ListType{ElemType: types.StringType},
	}
}

func tokenExchangeProfilesFromTF(ctx context.Context, o types.Object, diags *diag.Diagnostics) []mgmt.ClientTokenExchangeTypeEnum {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	v, ok := o.Attributes()["allow_any_profile_of_type"].(types.List)
	if !ok || v.IsNull() || v.IsUnknown() {
		return nil
	}
	raw := make([]string, 0, len(v.Elements()))
	diags.Append(v.ElementsAs(ctx, &raw, false)...)
	out := make([]mgmt.ClientTokenExchangeTypeEnum, 0, len(raw))
	for _, s := range raw {
		enum, err := mgmt.NewClientTokenExchangeTypeEnumFromString(s)
		if err != nil {
			diags.AddAttributeError(path.Root("token_exchange").AtName("allow_any_profile_of_type"), "Invalid value", err.Error())
			return nil
		}
		out = append(out, enum)
	}
	return out
}

func expandTokenExchangeCreate(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientTokenExchangeConfiguration {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	return &mgmt.ClientTokenExchangeConfiguration{
		AllowAnyProfileOfType: tokenExchangeProfilesFromTF(ctx, o, diags),
	}
}

func expandTokenExchangeUpdate(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientTokenExchangeConfigurationOrNull {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	return &mgmt.ClientTokenExchangeConfigurationOrNull{
		AllowAnyProfileOfType: tokenExchangeProfilesFromTF(ctx, o, diags),
	}
}

func flattenTokenExchange(in *mgmt.ClientTokenExchangeConfiguration, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(tokenExchangeAttrTypes())
	}
	v, d := types.ObjectValue(tokenExchangeAttrTypes(), map[string]attr.Value{
		"allow_any_profile_of_type": framework.EnumSliceToList(in.AllowAnyProfileOfType),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// signed_request_object  (JSON-string passthrough — Create vs Update use
// different SDK polymorphic types: WithPublicKey vs WithCredentialID.)
// =========================================================================

func signedRequestObjectFromJSON(jsonStr types.String, diags *diag.Diagnostics) (*mgmt.ClientSignedRequestObjectWithPublicKey, *mgmt.ClientSignedRequestObjectWithCredentialID, bool) {
	if jsonStr.IsNull() || jsonStr.IsUnknown() {
		return nil, nil, false
	}
	parsed, ok := framework.ParseJSONString(jsonStr, "signed_request_object", diags)
	if !ok {
		return nil, nil, false
	}
	b, err := json.Marshal(parsed)
	if err != nil {
		diags.AddError("Failed to re-encode signed_request_object", err.Error())
		return nil, nil, false
	}
	withPub := &mgmt.ClientSignedRequestObjectWithPublicKey{}
	withCred := &mgmt.ClientSignedRequestObjectWithCredentialID{}
	_ = json.Unmarshal(b, withPub)
	_ = json.Unmarshal(b, withCred)
	return withPub, withCred, true
}

func flattenSignedRequestObject(in any, diags *diag.Diagnostics) types.String {
	if in == nil {
		return types.StringNull()
	}
	return framework.FlattenJSONToString(in, diags)
}

// =========================================================================
// my_organization_configuration
// =========================================================================

func myOrgConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"connection_profile_id":        types.StringType,
		"user_attribute_profile_id":    types.StringType,
		"allowed_strategies":           types.ListType{ElemType: types.StringType},
		"connection_deletion_behavior": types.StringType,
	}
}

func flattenMyOrgConfig(in *mgmt.ClientMyOrganizationResponseConfiguration, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(myOrgConfigAttrTypes())
	}
	v, d := types.ObjectValue(myOrgConfigAttrTypes(), map[string]attr.Value{
		"connection_profile_id":        types.StringPointerValue(in.ConnectionProfileID),
		"user_attribute_profile_id":    types.StringPointerValue(in.UserAttributeProfileID),
		"allowed_strategies":           framework.EnumSliceToList(in.AllowedStrategies),
		"connection_deletion_behavior": types.StringValue(string(in.ConnectionDeletionBehavior)),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// express_configuration  (read-only — only returned on Get for OIN apps)
// =========================================================================

func linkedClientAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{"client_id": types.StringType}
}

func expressConfigurationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"initiate_login_uri_template": types.StringType,
		"user_attribute_profile_id":   types.StringType,
		"connection_profile_id":       types.StringType,
		"enable_client":               types.BoolType,
		"enable_organization":         types.BoolType,
		"linked_clients":              types.ListType{ElemType: types.ObjectType{AttrTypes: linkedClientAttrTypes()}},
		"okta_oin_client_id":          types.StringType,
		"admin_login_domain":          types.StringType,
		"oin_submission_id":           types.StringType,
	}
}

func flattenExpressConfiguration(in *mgmt.ExpressConfiguration, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(expressConfigurationAttrTypes())
	}
	objType := types.ObjectType{AttrTypes: linkedClientAttrTypes()}
	linked := types.ListNull(objType)
	if in.LinkedClients != nil {
		vals := make([]attr.Value, 0, len(in.LinkedClients))
		for _, l := range in.LinkedClients {
			if l == nil {
				continue
			}
			v, d := types.ObjectValue(linkedClientAttrTypes(), map[string]attr.Value{
				"client_id": types.StringValue(l.ClientID),
			})
			diags.Append(d...)
			vals = append(vals, v)
		}
		lv, d := types.ListValue(objType, vals)
		diags.Append(d...)
		linked = lv
	}
	v, d := types.ObjectValue(expressConfigurationAttrTypes(), map[string]attr.Value{
		"initiate_login_uri_template": types.StringValue(in.InitiateLoginURITemplate),
		"user_attribute_profile_id":   types.StringValue(in.UserAttributeProfileID),
		"connection_profile_id":       types.StringValue(in.ConnectionProfileID),
		"enable_client":               types.BoolValue(in.EnableClient),
		"enable_organization":         types.BoolValue(in.EnableOrganization),
		"linked_clients":              linked,
		"okta_oin_client_id":          types.StringValue(in.OktaOinClientID),
		"admin_login_domain":          types.StringValue(in.AdminLoginDomain),
		"oin_submission_id":           types.StringPointerValue(in.OinSubmissionID),
	})
	diags.Append(d...)
	return v
}

// suppress unused-import warnings for `time` until expires_at lands somewhere.
var _ = time.Time{}
