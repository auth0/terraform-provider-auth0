package client

import (
	"context"
	"encoding/json"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// =========================================================================
// expand: Terraform plan -> SDK request body
// =========================================================================

// applyCreateOptional copies optional fields from the Terraform plan onto the
// CreateClient request body.
//
// Required field `Name` is set by the caller.
func applyCreateOptional(ctx context.Context, plan *model, body *mgmt.CreateClientRequestContent, diags *diag.Diagnostics) {
	// scalars
	body.Description = strPtrFromTF(plan.Description)
	body.LogoURI = strPtrFromTF(plan.LogoURI)
	body.IsFirstParty = boolPtrFromTF(plan.IsFirstParty)
	body.OidcConformant = boolPtrFromTF(plan.OIDCConformant)
	body.SSO = boolPtrFromTF(plan.SSO)
	body.SSODisabled = boolPtrFromTF(plan.SSODisabled)
	body.CrossOriginAuthentication = boolPtrFromTF(plan.CrossOriginAuthentication)
	body.CustomLoginPageOn = boolPtrFromTF(plan.CustomLoginPageOn)
	body.IsTokenEndpointIPHeaderTrusted = boolPtrFromTF(plan.IsTokenEndpointIPHeaderTrusted)
	body.RequirePushedAuthorizationRequests = boolPtrFromTF(plan.RequirePushedAuthorizationRequests)
	body.RequireProofOfPossession = boolPtrFromTF(plan.RequireProofOfPossession)
	body.SkipNonVerifiableCallbackURIConfirmationPrompt = boolPtrFromTF(plan.SkipNonVerifiableCallbackURIConfirmationPrompt)

	body.CrossOriginLoc = strPtrFromTF(plan.CrossOriginLoc)
	body.InitiateLoginURI = strPtrFromTF(plan.InitiateLoginURI)
	body.FormTemplate = strPtrFromTF(plan.FormTemplate)
	body.CustomLoginPage = strPtrFromTF(plan.CustomLoginPage)
	body.CustomLoginPagePreview = strPtrFromTF(plan.CustomLoginPagePreview)
	body.ParRequestExpiry = framework.Int64ToIntPtr(plan.ParRequestExpiry)
	body.ResourceServerIdentifier = strPtrFromTF(plan.ResourceServerIdentifier)

	// list / map
	body.Callbacks = framework.StringListToGo(ctx, plan.Callbacks, diags)
	body.AllowedLogoutURLs = framework.StringListToGo(ctx, plan.AllowedLogoutURLs, diags)
	body.AllowedOrigins = framework.StringListToGo(ctx, plan.AllowedOrigins, diags)
	body.WebOrigins = framework.StringListToGo(ctx, plan.WebOrigins, diags)
	body.ClientAliases = framework.StringListToGo(ctx, plan.ClientAliases, diags)
	body.AllowedClients = framework.StringListToGo(ctx, plan.AllowedClients, diags)
	body.GrantTypes = framework.StringListToGo(ctx, plan.GrantTypes, diags)

	// enums
	if v := plan.AppType; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientAppTypeEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("app_type"), "Invalid app_type", err.Error())
			return
		}
		body.AppType = &enum
	}
	if v := plan.TokenEndpointAuthMethod; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientTokenEndpointAuthMethodEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("token_endpoint_auth_method"), "Invalid token_endpoint_auth_method", err.Error())
			return
		}
		body.TokenEndpointAuthMethod = &enum
	}
	if v := plan.ComplianceLevel; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientComplianceLevelEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("compliance_level"), "Invalid compliance_level", err.Error())
			return
		}
		body.ComplianceLevel = &enum
	}
	if v := plan.OrganizationUsage; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientOrganizationUsageEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("organization_usage"), "Invalid organization_usage", err.Error())
			return
		}
		body.OrganizationUsage = &enum
	}
	if v := plan.OrganizationRequireBehavior; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientOrganizationRequireBehaviorEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("organization_require_behavior"), "Invalid organization_require_behavior", err.Error())
			return
		}
		body.OrganizationRequireBehavior = &enum
	}
	if discovery := framework.StringListToGo(ctx, plan.OrganizationDiscoveryMethods, diags); discovery != nil {
		out := make([]mgmt.ClientOrganizationDiscoveryEnum, 0, len(discovery))
		for _, s := range discovery {
			enum, err := mgmt.NewClientOrganizationDiscoveryEnumFromString(s)
			if err != nil {
				diags.AddAttributeError(path.Root("organization_discovery_methods"), "Invalid organization_discovery_methods value", err.Error())
				return
			}
			out = append(out, enum)
		}
		body.OrganizationDiscoveryMethods = out
	}

	// metadata
	if md, d := expandClientMetadata(ctx, plan.ClientMetadata); d.HasError() {
		diags.Append(d...)
	} else if md != nil {
		body.ClientMetadata = &md
	}

	// nested: token_quota
	if tq, d := expandTokenQuotaCreate(ctx, plan.TokenQuota); d.HasError() {
		diags.Append(d...)
	} else if tq != nil {
		body.TokenQuota = tq
	}

	// Wave 2 nested objects
	if v := expandJwtConfiguration(ctx, plan.JwtConfiguration, diags); v != nil {
		body.JwtConfiguration = v
	}
	if v := expandRefreshToken(ctx, plan.RefreshToken, diags); v != nil {
		body.RefreshToken = v
	}
	if v := expandOidcLogout(ctx, plan.OidcLogout, diags); v != nil {
		body.OidcLogout = v
	}
	if v := expandOidcLogout(ctx, plan.OidcBackchannelLogout, diags); v != nil {
		body.OidcBackchannelLogout = v
	}
	if v := expandEncryptionKey(ctx, plan.EncryptionKey); v != nil {
		body.EncryptionKey = v
	}
	if v := expandDefaultOrganization(ctx, plan.DefaultOrganization, diags); v != nil {
		body.DefaultOrganization = v
	}
	if v := expandNativeSocialLogin(ctx, plan.NativeSocialLogin); v != nil {
		body.NativeSocialLogin = v
	}
	if v := expandSessionTransfer(ctx, plan.SessionTransfer, diags); v != nil {
		body.SessionTransfer = v
	}
	if v := expandMobile(ctx, plan.Mobile, diags); v != nil {
		body.Mobile = v
	}
	if v := expandTokenExchangeCreate(ctx, plan.TokenExchange, diags); v != nil {
		body.TokenExchange = v
	}

	// list-of-enum
	if !plan.AsyncApprovalNotificationChannels.IsNull() && !plan.AsyncApprovalNotificationChannels.IsUnknown() {
		raw := framework.StringListToGo(ctx, plan.AsyncApprovalNotificationChannels, diags)
		out := make(mgmt.ClientAsyncApprovalNotificationsChannelsAPIPostConfiguration, 0, len(raw))
		for _, s := range raw {
			enum, err := mgmt.NewAsyncApprovalNotificationsChannelsEnumFromString(s)
			if err != nil {
				diags.AddAttributeError(path.Root("async_approval_notification_channels"), "Invalid channel", err.Error())
				return
			}
			out = append(out, enum)
		}
		body.AsyncApprovalNotificationChannels = &out
	}

	// JSON-string passthroughs
	if pubKey, _, ok := signedRequestObjectFromJSON(plan.SignedRequestObject, diags); ok {
		body.SignedRequestObject = pubKey
	}
	if v := expandAddons(ctx, plan.Addons, diags); v != nil {
		body.Addons = v
	}
	if v, ok := framework.ParseJSONString(plan.ClientAuthenticationMethods, "client_authentication_methods", diags); ok {
		cam := &mgmt.ClientCreateAuthenticationMethod{}
		if _, err := jsonRoundTrip(v, cam); err != nil {
			diags.AddError("Failed to decode client_authentication_methods", err.Error())
		} else {
			body.ClientAuthenticationMethods = cam
		}
	}
}

// applyUpdateOptional is the PATCH-equivalent.
func applyUpdateOptional(ctx context.Context, plan *model, body *mgmt.UpdateClientRequestContent, diags *diag.Diagnostics) {
	body.Description = strPtrFromTF(plan.Description)
	body.LogoURI = strPtrFromTF(plan.LogoURI)
	body.IsFirstParty = boolPtrFromTF(plan.IsFirstParty)
	body.OidcConformant = boolPtrFromTF(plan.OIDCConformant)
	body.SSO = boolPtrFromTF(plan.SSO)
	body.SSODisabled = boolPtrFromTF(plan.SSODisabled)
	body.CrossOriginAuthentication = boolPtrFromTF(plan.CrossOriginAuthentication)
	body.CustomLoginPageOn = boolPtrFromTF(plan.CustomLoginPageOn)
	body.IsTokenEndpointIPHeaderTrusted = boolPtrFromTF(plan.IsTokenEndpointIPHeaderTrusted)
	body.RequirePushedAuthorizationRequests = boolPtrFromTF(plan.RequirePushedAuthorizationRequests)
	body.RequireProofOfPossession = boolPtrFromTF(plan.RequireProofOfPossession)
	body.SkipNonVerifiableCallbackURIConfirmationPrompt = boolPtrFromTF(plan.SkipNonVerifiableCallbackURIConfirmationPrompt)
	body.CrossOriginLoc = strPtrFromTF(plan.CrossOriginLoc)
	body.InitiateLoginURI = strPtrFromTF(plan.InitiateLoginURI)
	body.FormTemplate = strPtrFromTF(plan.FormTemplate)
	body.CustomLoginPage = strPtrFromTF(plan.CustomLoginPage)
	body.CustomLoginPagePreview = strPtrFromTF(plan.CustomLoginPagePreview)
	body.ParRequestExpiry = framework.Int64ToIntPtr(plan.ParRequestExpiry)

	body.Callbacks = framework.StringListToGo(ctx, plan.Callbacks, diags)
	body.AllowedLogoutURLs = framework.StringListToGo(ctx, plan.AllowedLogoutURLs, diags)
	body.AllowedOrigins = framework.StringListToGo(ctx, plan.AllowedOrigins, diags)
	body.WebOrigins = framework.StringListToGo(ctx, plan.WebOrigins, diags)
	body.ClientAliases = framework.StringListToGo(ctx, plan.ClientAliases, diags)
	body.AllowedClients = framework.StringListToGo(ctx, plan.AllowedClients, diags)
	body.GrantTypes = framework.StringListToGo(ctx, plan.GrantTypes, diags)

	if v := plan.AppType; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientAppTypeEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("app_type"), "Invalid app_type", err.Error())
			return
		}
		body.AppType = &enum
	}
	if v := plan.TokenEndpointAuthMethod; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientTokenEndpointAuthMethodOrNullEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("token_endpoint_auth_method"), "Invalid token_endpoint_auth_method", err.Error())
			return
		}
		body.TokenEndpointAuthMethod = &enum
	}
	if v := plan.ComplianceLevel; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientComplianceLevelEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("compliance_level"), "Invalid compliance_level", err.Error())
			return
		}
		body.ComplianceLevel = &enum
	}
	if v := plan.OrganizationUsage; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientOrganizationUsagePatchEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("organization_usage"), "Invalid organization_usage", err.Error())
			return
		}
		body.OrganizationUsage = &enum
	}
	if v := plan.OrganizationRequireBehavior; !v.IsNull() && !v.IsUnknown() {
		enum, err := mgmt.NewClientOrganizationRequireBehaviorPatchEnumFromString(v.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("organization_require_behavior"), "Invalid organization_require_behavior", err.Error())
			return
		}
		body.OrganizationRequireBehavior = &enum
	}
	if discovery := framework.StringListToGo(ctx, plan.OrganizationDiscoveryMethods, diags); discovery != nil {
		out := make([]mgmt.ClientOrganizationDiscoveryEnum, 0, len(discovery))
		for _, s := range discovery {
			enum, err := mgmt.NewClientOrganizationDiscoveryEnumFromString(s)
			if err != nil {
				diags.AddAttributeError(path.Root("organization_discovery_methods"), "Invalid organization_discovery_methods value", err.Error())
				return
			}
			out = append(out, enum)
		}
		body.OrganizationDiscoveryMethods = out
	}

	if md, d := expandClientMetadata(ctx, plan.ClientMetadata); d.HasError() {
		diags.Append(d...)
	} else if md != nil {
		body.ClientMetadata = &md
	}

	if tq, d := expandTokenQuotaUpdate(ctx, plan.TokenQuota); d.HasError() {
		diags.Append(d...)
	} else if tq != nil {
		body.TokenQuota = tq
	}

	// Wave 2 nested objects
	if v := expandJwtConfiguration(ctx, plan.JwtConfiguration, diags); v != nil {
		body.JwtConfiguration = v
	}
	if v := expandRefreshToken(ctx, plan.RefreshToken, diags); v != nil {
		body.RefreshToken = v
	}
	if v := expandOidcLogout(ctx, plan.OidcLogout, diags); v != nil {
		body.OidcLogout = v
	}
	if v := expandOidcLogout(ctx, plan.OidcBackchannelLogout, diags); v != nil {
		body.OidcBackchannelLogout = v
	}
	if v := expandEncryptionKey(ctx, plan.EncryptionKey); v != nil {
		body.EncryptionKey = v
	}
	if v := expandDefaultOrganization(ctx, plan.DefaultOrganization, diags); v != nil {
		body.DefaultOrganization = v
	}
	if v := expandNativeSocialLogin(ctx, plan.NativeSocialLogin); v != nil {
		body.NativeSocialLogin = v
	}
	if v := expandSessionTransfer(ctx, plan.SessionTransfer, diags); v != nil {
		body.SessionTransfer = v
	}
	if v := expandMobile(ctx, plan.Mobile, diags); v != nil {
		body.Mobile = v
	}
	if v := expandTokenExchangeUpdate(ctx, plan.TokenExchange, diags); v != nil {
		body.TokenExchange = v
	}

	if !plan.AsyncApprovalNotificationChannels.IsNull() && !plan.AsyncApprovalNotificationChannels.IsUnknown() {
		raw := framework.StringListToGo(ctx, plan.AsyncApprovalNotificationChannels, diags)
		out := make(mgmt.ClientAsyncApprovalNotificationsChannelsAPIPostConfiguration, 0, len(raw))
		for _, s := range raw {
			enum, err := mgmt.NewAsyncApprovalNotificationsChannelsEnumFromString(s)
			if err != nil {
				diags.AddAttributeError(path.Root("async_approval_notification_channels"), "Invalid channel", err.Error())
				return
			}
			out = append(out, enum)
		}
		body.AsyncApprovalNotificationChannels = &out
	}

	// JSON-string passthroughs — Update uses different polymorphic types.
	if _, withCred, ok := signedRequestObjectFromJSON(plan.SignedRequestObject, diags); ok {
		body.SignedRequestObject = withCred
	}
	if v := expandAddons(ctx, plan.Addons, diags); v != nil {
		body.Addons = v
	}
	if v, ok := framework.ParseJSONString(plan.ClientAuthenticationMethods, "client_authentication_methods", diags); ok {
		cam := &mgmt.ClientAuthenticationMethod{}
		if _, err := jsonRoundTrip(v, cam); err != nil {
			diags.AddError("Failed to decode client_authentication_methods", err.Error())
		} else {
			body.ClientAuthenticationMethods = cam
		}
	}
}

// =========================================================================
// expand helpers (top-level + nested)
// =========================================================================

// expandClientMetadata converts the TF map into mgmt.ClientMetadata
// (`map[string]any`). Returns nil for null/unknown so the field is omitted.
func expandClientMetadata(ctx context.Context, m types.Map) (mgmt.ClientMetadata, diag.Diagnostics) {
	if m.IsNull() || m.IsUnknown() {
		return nil, nil
	}
	raw := make(map[string]string, len(m.Elements()))
	d := m.ElementsAs(ctx, &raw, false)
	if d.HasError() {
		return nil, d
	}
	out := make(mgmt.ClientMetadata, len(raw))
	for k, v := range raw {
		out[k] = v
	}
	return out, nil
}

func tokenQuotaCCFromTF(o types.Object) (*mgmt.TokenQuotaClientCredentials, diag.Diagnostics) {
	var diags diag.Diagnostics
	if o.IsNull() || o.IsUnknown() {
		return nil, diags
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
	return cc, diags
}

// expandTokenQuotaCreate converts the typed object to *mgmt.CreateTokenQuota.
func expandTokenQuotaCreate(_ context.Context, o types.Object) (*mgmt.CreateTokenQuota, diag.Diagnostics) {
	var diags diag.Diagnostics
	if o.IsNull() || o.IsUnknown() {
		return nil, diags
	}
	cc, _ := childObject(o, "client_credentials")
	ccVal, d := tokenQuotaCCFromTF(cc)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return &mgmt.CreateTokenQuota{ClientCredentials: ccVal}, diags
}

// expandTokenQuotaUpdate converts the typed object to *mgmt.UpdateTokenQuota.
func expandTokenQuotaUpdate(_ context.Context, o types.Object) (*mgmt.UpdateTokenQuota, diag.Diagnostics) {
	var diags diag.Diagnostics
	if o.IsNull() || o.IsUnknown() {
		return nil, diags
	}
	cc, _ := childObject(o, "client_credentials")
	ccVal, d := tokenQuotaCCFromTF(cc)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	return &mgmt.UpdateTokenQuota{ClientCredentials: ccVal}, diags
}

// =========================================================================
// tiny boilerplate eliminators (used by expand)
// =========================================================================

// strPtrFromTF returns a *string from a TF string (nil for null/unknown).
func strPtrFromTF(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueStringPointer()
}

// boolPtrFromTF returns a *bool from a TF bool (nil for null/unknown).
func boolPtrFromTF(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueBoolPointer()
}

// childObject pulls a nested attribute as a types.Object. The boolean is true
// when present (we currently ignore it; the caller treats absence as null).
func childObject(parent types.Object, name string) (types.Object, bool) {
	v, ok := parent.Attributes()[name]
	if !ok {
		return types.ObjectNull(nil), false
	}
	o, ok := v.(types.Object)
	if !ok {
		return types.ObjectNull(nil), false
	}
	return o, true
}

// jsonRoundTrip marshals `v` to JSON then unmarshals it into `into`. Used for
// the JSON-string passthrough fields (addons, client_authentication_methods,
// signed_request_object) where the user supplies arbitrary JSON and we hand
// it to a typed SDK struct via JSON tags.
func jsonRoundTrip(v any, into any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, into); err != nil {
		return nil, err
	}
	return b, nil
}
