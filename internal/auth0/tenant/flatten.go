package tenant

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenTenant(data *schema.ResourceData, tenant *management.Tenant) error {
	result := multierror.Append(
		data.Set("default_audience", tenant.GetDefaultAudience()),
		data.Set("default_directory", tenant.GetDefaultDirectory()),
		data.Set("default_redirection_uri", tenant.GetDefaultRedirectionURI()),
		data.Set("friendly_name", tenant.GetFriendlyName()),
		data.Set("picture_url", tenant.GetPictureURL()),
		data.Set("support_email", tenant.GetSupportEmail()),
		data.Set("support_url", tenant.GetSupportURL()),
		data.Set("allowed_logout_urls", tenant.GetAllowedLogoutURLs()),
		data.Set("session_lifetime", tenant.GetSessionLifetime()),
		data.Set("sandbox_version", tenant.GetSandboxVersion()),
		data.Set("enabled_locales", tenant.GetEnabledLocales()),
		data.Set("flags", flattenTenantFlags(tenant.GetFlags())),
		data.Set("session_cookie", flattenTenantSessionCookie(tenant.GetSessionCookie())),
		data.Set("sessions", flattenTenantSessions(tenant.GetSessions())),
		data.Set("allow_organization_name_in_authentication_api", tenant.GetAllowOrgNameInAuthAPI()),
	)

	if tenant.GetIdleSessionLifetime() == 0 {
		idleSessionLifetimeDefault := NewResource().Schema["idle_session_lifetime"].Default
		result = multierror.Append(result, data.Set("idle_session_lifetime", idleSessionLifetimeDefault))
	} else {
		result = multierror.Append(result, data.Set("idle_session_lifetime", tenant.GetIdleSessionLifetime()))
	}

	return result.ErrorOrNil()
}

func flattenTenantFlags(flags *management.TenantFlags) []interface{} {
	if flags == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["enable_client_connections"] = flags.EnableClientConnections
	m["enable_apis_section"] = flags.EnableAPIsSection
	m["enable_pipeline2"] = flags.EnablePipeline2
	m["enable_dynamic_client_registration"] = flags.EnableDynamicClientRegistration
	m["enable_custom_domain_in_emails"] = flags.EnableCustomDomainInEmails
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
	m["require_pushed_authorization_requests"] = flags.RequirePushedAuthorizationRequests

	return []interface{}{m}
}

func flattenTenantSessionCookie(sessionCookie *management.TenantSessionCookie) []interface{} {
	m := make(map[string]interface{})
	m["mode"] = sessionCookie.GetMode()

	return []interface{}{m}
}

func flattenTenantSessions(sessions *management.TenantSessions) []interface{} {
	m := make(map[string]interface{})
	m["oidc_logout_prompt_enabled"] = sessions.GetOIDCLogoutPromptEnabled()

	return []interface{}{m}
}
