package auth0

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
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

func expandTenantChangePassword(d ResourceData) *management.TenantChangePassword {
	var changePassword management.TenantChangePassword

	List(d, "change_password").Elem(func(d ResourceData) {
		changePassword.Enabled = Bool(d, "enabled")
		changePassword.HTML = String(d, "html")
	})

	return &changePassword
}

func expandTenantGuardianMFAPage(d ResourceData) *management.TenantGuardianMFAPage {
	var mfa management.TenantGuardianMFAPage

	List(d, "guardian_mfa_page").Elem(func(d ResourceData) {
		mfa.Enabled = Bool(d, "enabled")
		mfa.HTML = String(d, "html")
	})

	return &mfa
}

func expandTenantErrorPage(d ResourceData) *management.TenantErrorPage {
	var errorPage management.TenantErrorPage

	List(d, "error_page").Elem(func(d ResourceData) {
		errorPage.HTML = String(d, "html")
		errorPage.ShowLogLink = Bool(d, "show_log_link")
		errorPage.URL = String(d, "url")
	})

	return &errorPage
}

func expandTenantFlags(flagsList cty.Value) *management.TenantFlags {
	var tenantFlags *management.TenantFlags

	flagsList.ForEachElement(func(_ cty.Value, flags cty.Value) (stop bool) {
		tenantFlags = &management.TenantFlags{
			EnableClientConnections:           Flag(flags, "enable_client_connections"),
			EnableAPIsSection:                 Flag(flags, "enable_apis_section"),
			EnablePipeline2:                   Flag(flags, "enable_pipeline2"),
			EnableDynamicClientRegistration:   Flag(flags, "enable_dynamic_client_registration"),
			EnableCustomDomainInEmails:        Flag(flags, "enable_custom_domain_in_emails"),
			UniversalLogin:                    Flag(flags, "universal_login"),
			EnableLegacyLogsSearchV2:          Flag(flags, "enable_legacy_logs_search_v2"),
			DisableClickjackProtectionHeaders: Flag(flags, "disable_clickjack_protection_headers"),
			EnablePublicSignupUserExistsError: Flag(flags, "enable_public_signup_user_exists_error"),
			UseScopeDescriptionsForConsent:    Flag(flags, "use_scope_descriptions_for_consent"),
		}

		return stop
	})

	return tenantFlags
}

func expandTenantUniversalLogin(d ResourceData) *management.TenantUniversalLogin {
	var universalLogin management.TenantUniversalLogin

	List(d, "universal_login").Elem(func(d ResourceData) {
		List(d, "colors").Elem(func(d ResourceData) {
			universalLogin.Colors = &management.TenantUniversalLoginColors{
				Primary:        String(d, "primary"),
				PageBackground: String(d, "page_background"),
			}
		})
	})

	return &universalLogin
}
