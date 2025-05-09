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
		data.Set("idle_session_lifetime", tenant.GetIdleSessionLifetime()),
		data.Set("sandbox_version", tenant.GetSandboxVersion()),
		data.Set("enabled_locales", tenant.GetEnabledLocales()),
		data.Set("flags", flattenTenantFlags(tenant.GetFlags())),
		data.Set("session_cookie", flattenTenantSessionCookie(tenant.GetSessionCookie())),
		data.Set("sessions", flattenTenantSessions(tenant.GetSessions())),
		data.Set("oidc_logout", flattenTenantOidcLogout(tenant.GetOIDCLogout())),
		data.Set("allow_organization_name_in_authentication_api", tenant.GetAllowOrgNameInAuthAPI()),
		data.Set("customize_mfa_in_postlogin_action", tenant.GetCustomizeMFAInPostLoginAction()),
		data.Set("pushed_authorization_requests_supported", tenant.GetPushedAuthorizationRequestsSupported()),
		data.Set("mtls", flattenMTLSConfiguration(tenant.GetMTLS())),
		data.Set("error_page", flattenErrorPageConfiguration(tenant.GetErrorPage())),
		data.Set("default_token_quota", flattenDefaultTokenQuota(tenant.GetDefaultTokenQuota())),
	)

	if tenant.GetIdleSessionLifetime() == 0 {
		result = multierror.Append(result, data.Set("idle_session_lifetime", idleSessionLifetimeDefault))
	}

	if tenant.GetSessionLifetime() == 0 {
		result = multierror.Append(result, data.Set("session_lifetime", sessionLifetimeDefault))
	}

	if tenant.GetACRValuesSupported() == nil {
		result = multierror.Append(result,
			data.Set("disable_acr_values_supported", true),
			data.Set("acr_values_supported", nil),
		)
	} else {
		result = multierror.Append(result,
			data.Set("acr_values_supported", tenant.GetACRValuesSupported()),
			data.Set("disable_acr_values_supported", false),
		)
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

func flattenTenantOidcLogout(oidcLogout *management.TenantOIDCLogout) []interface{} {
	m := make(map[string]interface{})
	m["rp_logout_end_session_endpoint_discovery"] = oidcLogout.GetOIDCResourceProviderLogoutEndSessionEndpointDiscovery()

	return []interface{}{m}
}

func flattenMTLSConfiguration(mtls *management.TenantMTLSConfiguration) []interface{} {
	m := make(map[string]interface{})
	if mtls == nil {
		m["disable"] = true
	} else {
		m["enable_endpoint_aliases"] = mtls.EnableEndpointAliases
	}

	return []interface{}{m}
}

func flattenErrorPageConfiguration(errorPage *management.TenantErrorPage) []interface{} {
	if errorPage == nil {
		return nil
	}

	m := make(map[string]interface{})

	m["html"] = errorPage.GetHTML()
	m["show_log_link"] = errorPage.GetShowLogLink()
	m["url"] = errorPage.GetURL()

	return []interface{}{m}
}

func flattenDefaultTokenQuota(defaultTokenQuota *management.TenantDefaultTokenQuota) []interface{} {
	if defaultTokenQuota == nil {
		return nil
	}

	m := make(map[string]interface{})

	if defaultTokenQuota.Clients != nil {
		m["clients"] = flattenTokenQuota(defaultTokenQuota.Clients)
	}

	if defaultTokenQuota.Organizations != nil {
		m["organizations"] = flattenTokenQuota(defaultTokenQuota.Organizations)
	}

	return []interface{}{m}
}

func flattenTokenQuota(tokenQuota *management.TokenQuota) []interface{} {
	if tokenQuota == nil || tokenQuota.ClientCredentials == nil {
		return nil
	}

	m := make(map[string]interface{})

	clientCreds := make(map[string]interface{})
	if tokenQuota.ClientCredentials.Enforce != nil {
		clientCreds["enforce"] = *tokenQuota.ClientCredentials.Enforce
	}

	if tokenQuota.ClientCredentials.PerDay != nil {
		clientCreds["per_day"] = *tokenQuota.ClientCredentials.PerDay
	}

	if tokenQuota.ClientCredentials.PerHour != nil {
		clientCreds["per_hour"] = *tokenQuota.ClientCredentials.PerHour
	}

	m["client_credentials"] = []interface{}{clientCreds}

	return []interface{}{m}
}
