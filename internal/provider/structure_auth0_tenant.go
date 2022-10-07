package provider

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func flattenTenantChangePassword(changePassword *management.TenantChangePassword) []interface{} {
	if changePassword == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["enabled"] = changePassword.Enabled
	m["html"] = changePassword.HTML

	return []interface{}{m}
}

func flattenTenantGuardianMFAPage(mfa *management.TenantGuardianMFAPage) []interface{} {
	if mfa == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["enabled"] = mfa.Enabled
	m["html"] = mfa.HTML

	return []interface{}{m}
}

func flattenTenantErrorPage(errorPage *management.TenantErrorPage) []interface{} {
	if errorPage == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["html"] = errorPage.HTML
	m["show_log_link"] = errorPage.ShowLogLink
	m["url"] = errorPage.URL

	return []interface{}{m}
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
	m["universal_login"] = flags.UniversalLogin
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

	return []interface{}{m}
}

func flattenTenantUniversalLogin(universalLogin *management.TenantUniversalLogin) []interface{} {
	if universalLogin == nil {
		return nil
	}
	if universalLogin.Colors == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["colors"] = []interface{}{
		map[string]interface{}{
			"primary":         universalLogin.Colors.Primary,
			"page_background": universalLogin.Colors.PageBackground,
		},
	}

	return []interface{}{m}
}

func flattenTenantSessionCookie(sessionCookie *management.TenantSessionCookie) []interface{} {
	m := make(map[string]interface{})
	m["mode"] = sessionCookie.GetMode()

	return []interface{}{m}
}

func expandTenantChangePassword(config cty.Value) *management.TenantChangePassword {
	var changePassword management.TenantChangePassword

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		changePassword.Enabled = value.Bool(d.GetAttr("enabled"))
		changePassword.HTML = value.String(d.GetAttr("html"))
		return stop
	})

	return &changePassword
}

func expandTenantGuardianMFAPage(config cty.Value) *management.TenantGuardianMFAPage {
	var mfa management.TenantGuardianMFAPage

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		mfa.Enabled = value.Bool(d.GetAttr("enabled"))
		mfa.HTML = value.String(d.GetAttr("html"))
		return stop
	})

	return &mfa
}

func expandTenantErrorPage(config cty.Value) *management.TenantErrorPage {
	var errorPage management.TenantErrorPage

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		errorPage.HTML = value.String(d.GetAttr("html"))
		errorPage.ShowLogLink = value.Bool(d.GetAttr("show_log_link"))
		errorPage.URL = value.String(d.GetAttr("url"))
		return stop
	})

	return &errorPage
}

func expandTenantFlags(config cty.Value) *management.TenantFlags {
	var tenantFlags *management.TenantFlags

	config.ForEachElement(func(_ cty.Value, flags cty.Value) (stop bool) {
		tenantFlags = &management.TenantFlags{
			EnableClientConnections:            value.Bool(flags.GetAttr("enable_client_connections")),
			EnableAPIsSection:                  value.Bool(flags.GetAttr("enable_apis_section")),
			EnablePipeline2:                    value.Bool(flags.GetAttr("enable_pipeline2")),
			EnableDynamicClientRegistration:    value.Bool(flags.GetAttr("enable_dynamic_client_registration")),
			EnableCustomDomainInEmails:         value.Bool(flags.GetAttr("enable_custom_domain_in_emails")),
			UniversalLogin:                     value.Bool(flags.GetAttr("universal_login")),
			EnableLegacyLogsSearchV2:           value.Bool(flags.GetAttr("enable_legacy_logs_search_v2")),
			DisableClickjackProtectionHeaders:  value.Bool(flags.GetAttr("disable_clickjack_protection_headers")),
			EnablePublicSignupUserExistsError:  value.Bool(flags.GetAttr("enable_public_signup_user_exists_error")),
			UseScopeDescriptionsForConsent:     value.Bool(flags.GetAttr("use_scope_descriptions_for_consent")),
			AllowLegacyDelegationGrantTypes:    value.Bool(flags.GetAttr("allow_legacy_delegation_grant_types")),
			AllowLegacyROGrantTypes:            value.Bool(flags.GetAttr("allow_legacy_ro_grant_types")),
			AllowLegacyTokenInfoEndpoint:       value.Bool(flags.GetAttr("allow_legacy_tokeninfo_endpoint")),
			EnableLegacyProfile:                value.Bool(flags.GetAttr("enable_legacy_profile")),
			EnableIDTokenAPI2:                  value.Bool(flags.GetAttr("enable_idtoken_api2")),
			NoDisclosureEnterpriseConnections:  value.Bool(flags.GetAttr("no_disclose_enterprise_connections")),
			DisableManagementAPISMSObfuscation: value.Bool(flags.GetAttr("disable_management_api_sms_obfuscation")),
			EnableADFSWAADEmailVerification:    value.Bool(flags.GetAttr("enable_adfs_waad_email_verification")),
			RevokeRefreshTokenGrant:            value.Bool(flags.GetAttr("revoke_refresh_token_grant")),
			DashboardLogStreams:                value.Bool(flags.GetAttr("dashboard_log_streams_next")),
			DashboardInsightsView:              value.Bool(flags.GetAttr("dashboard_insights_view")),
			DisableFieldsMapFix:                value.Bool(flags.GetAttr("disable_fields_map_fix")),
		}

		return stop
	})

	return tenantFlags
}

func expandTenantUniversalLogin(config cty.Value) *management.TenantUniversalLogin {
	var universalLogin management.TenantUniversalLogin

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		colors := d.GetAttr("colors")

		colors.ForEachElement(func(_ cty.Value, color cty.Value) (stop bool) {
			universalLogin.Colors = &management.TenantUniversalLoginColors{
				Primary:        value.String(color.GetAttr("primary")),
				PageBackground: value.String(color.GetAttr("page_background")),
			}
			return stop
		})
		return stop
	})

	return &universalLogin
}

func expandTenantSessionCookie(config cty.Value) *management.TenantSessionCookie {
	var sessionCookie management.TenantSessionCookie

	if config.LengthInt() == 0 {
		return nil
	}

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		sessionCookie.Mode = value.String(d.GetAttr("mode"))
		return stop
	})

	return &sessionCookie
}
