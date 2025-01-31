package tenant

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenTenant(ctx context.Context,
	state *tfsdk.State,
	tenant *management.Tenant,
	model *resourceModel,
) (diagnostics diag.Diagnostics) {
	diagnostics.Append(state.SetAttribute(ctx, path.Root("default_audience"), value.AttrString(tenant.GetDefaultAudience()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("default_directory"), value.AttrString(tenant.GetDefaultDirectory()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("default_redirection_uri"), value.AttrString(tenant.GetDefaultRedirectionURI()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("friendly_name"), value.AttrString(tenant.GetFriendlyName()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("picture_url"), value.AttrString(tenant.GetPictureURL()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("support_email"), value.AttrString(tenant.GetSupportEmail()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("support_url"), value.AttrString(tenant.GetSupportURL()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("allowed_logout_urls"), value.AttrStringList(tenant.GetAllowedLogoutURLs()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("session_lifetime"), value.AttrFloat64(tenant.GetSessionLifetime()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("idle_session_lifetime"), value.AttrFloat64(tenant.GetIdleSessionLifetime()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("sandbox_version"), value.AttrString(tenant.GetSandboxVersion()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("enabled_locales"), value.AttrString(tenant.GetEnabledLocales()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("flags"), flattenTenantFlags(tenant.GetFlags()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("session_cookie"), flattenTenantSessionCookie(tenant.GetSessionCookie()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("sessions"), flattenTenantSessions(tenant.GetSessions()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("allow_organization_name_in_authentication_api"), value.AttrBool(tenant.GetAllowOrgNameInAuthAPI()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("customize_mfa_in_postlogin_action"), value.AttrString(tenant.GetCustomizeMFAInPostLoginAction()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("pushed_authorization_requests_supported"), value.AttrBool(tenant.GetPushedAuthorizationRequestsSupported()))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("mtls"), flattenMTLSConfiguration(tenant.GetMTLS()))...)

	if tenant.GetIdleSessionLifetime() == 0 {
		diagnostics.Append(state.SetAttribute(ctx, path.Root("idle_session_lifetime"), value.AttrFloat64(idleSessionLifetimeDefault))...)
	}

	if tenant.GetSessionLifetime() == 0 {
		diagnostics.Append(state.SetAttribute(ctx, path.Root("session_lifetime"), value.AttrFloat64(sessionLifetimeDefault))...)
	}

	if tenant.GetACRValuesSupported() == nil {
		diagnostics.Append(state.SetAttribute(ctx, path.Root("disable_acr_values_supported"), value.AttrBool(true))...)
		diagnostics.Append(state.SetAttribute(ctx, path.Root("acr_values_supported"), value.SetNull(types.StringType))...)
	} else {
		diagnostics.Append(state.SetAttribute(ctx, path.Root("disable_acr_values_supported"), value.AttrBool(false))...)
		diagnostics.Append(state.SetAttribute(ctx, path.Root("acr_values_supported"), value.StringSet(tenant.GetACRValuesSupported()))...)
	}

	return result.ErrorOrNil()
}

func flattenTenantFlags(flags *management.TenantFlags) map[string]*string {
	if flags == nil {
		return nil
	}

	m := make(map[string]*string)
	m["enable_client_connections"] = flags.EnableClientConnections
	m["enable_apis_section"] = flags.EnableAPIsSection
	m["enable_pipeline2"] = flags.EnablePipeline2
	m["enable_dynamic_client_registration"] = flags.EnableDynamicClientRegistration
	m["enable_custom_domain_in_emails"] = flags.EnableCustomDomainInEmails
	m["enable_sso"] = flags.EnableSSO
	m["enable_legacy_logs_search_v2"] = flags.EnableLegacyLogsSearchV2
	m["disable_clickjack_protection_headers"] = flags.DisableClickjackProtectionHeaders
	m["enable_public_signup_user_exists_error"] = flags.EnablePublicSignupUserExistsError
	m["use_scope_descriptions_for_consent"] = flags.UseScopeDescriptionsForConsent
	m["allow_legacy_delegation_grant_types"] = flags.AllowLegacyDelegationGrantTypes
	m["allow_legacy_ro_grant_types"] = flags.AllowLegacyROGrantTypes
	m["allow_legacy_tokeninfo_endpoint"] = flags.AllowLegacyTokenInfoEndpoint
	m["enable_legacy_profile"] = flags.EnableLegacyProfile
	m["enable_idtoken_api2"] = flags.EnableIDTokenAPI2
	m["no_disclose_enterprise_connections"] = flags.NoDisclosureEnterpriseConnections
	m["disable_management_api_sms_obfuscation"] = flags.DisableManagementAPISMSObfuscation
	m["enable_adfs_waad_email_verification"] = flags.EnableADFSWAADEmailVerification
	m["revoke_refresh_token_grant"] = flags.RevokeRefreshTokenGrant
	m["dashboard_log_streams_next"] = flags.DashboardLogStreams
	m["dashboard_insights_view"] = flags.DashboardInsightsView
	m["disable_fields_map_fix"] = flags.DisableFieldsMapFix
	m["mfa_show_factor_list_on_enrollment"] = flags.MFAShowFactorListOnEnrollment
	m["remove_alg_from_jwks"] = flags.RemoveAlgFromJWKS

	return m
}

func flattenTenantSessionCookie(sessionCookie *management.TenantSessionCookie) map[string]*string {
	m := make(map[string]*string)
	m["mode"] = sessionCookie.GetMode()

	return m
}

func flattenTenantSessions(sessions *management.TenantSessions) map[string]*string {
	m := make(map[string]*string)
	m["oidc_logout_prompt_enabled"] = sessions.GetOIDCLogoutPromptEnabled()

	return m
}

func flattenMTLSConfiguration(mtls *management.TenantMTLSConfiguration) map[string]*string {
	m := make(map[string]*string)
	if mtls == nil {
		m["disable"] = true
	} else {
		m["enable_endpoint_aliases"] = mtls.EnableEndpointAliases
	}

	return m
}
