package tenant

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandTenant(d *schema.ResourceData) *management.Tenant {
	config := d.GetRawConfig()

	sessionLifetime := d.Get("session_lifetime").(float64)          // Handling separately to preserve default values not honored by `d.GetRawConfig()`
	idleSessionLifetime := d.Get("idle_session_lifetime").(float64) // Handling separately to preserve default values not honored by `d.GetRawConfig()`

	tenant := &management.Tenant{
		DefaultAudience:       value.String(config.GetAttr("default_audience")),
		DefaultDirectory:      value.String(config.GetAttr("default_directory")),
		DefaultRedirectionURI: value.String(config.GetAttr("default_redirection_uri")),
		FriendlyName:          value.String(config.GetAttr("friendly_name")),
		PictureURL:            value.String(config.GetAttr("picture_url")),
		SupportEmail:          value.String(config.GetAttr("support_email")),
		SupportURL:            value.String(config.GetAttr("support_url")),
		AllowedLogoutURLs:     value.Strings(config.GetAttr("allowed_logout_urls")),
		SessionLifetime:       &sessionLifetime,
		SandboxVersion:        value.String(config.GetAttr("sandbox_version")),
		EnabledLocales:        value.Strings(config.GetAttr("enabled_locales")),
		ChangePassword:        expandTenantChangePassword(config.GetAttr("change_password")),
		GuardianMFAPage:       expandTenantGuardianMFAPage(config.GetAttr("guardian_mfa_page")),
		ErrorPage:             expandTenantErrorPage(config.GetAttr("error_page")),
		Flags:                 expandTenantFlags(config.GetAttr("flags")),
		UniversalLogin:        expandTenantUniversalLogin(config.GetAttr("universal_login")),
		SessionCookie:         expandTenantSessionCookie(config.GetAttr("session_cookie")),
	}

	if d.IsNewResource() || d.HasChange("idle_session_lifetime") {
		tenant.IdleSessionLifetime = &idleSessionLifetime
	}

	return tenant
}

func expandTenantChangePassword(config cty.Value) *management.TenantChangePassword {
	var changePassword management.TenantChangePassword

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		changePassword.Enabled = value.Bool(d.GetAttr("enabled"))
		changePassword.HTML = value.String(d.GetAttr("html"))
		return stop
	})

	if changePassword == (management.TenantChangePassword{}) {
		return nil
	}

	return &changePassword
}

func expandTenantGuardianMFAPage(config cty.Value) *management.TenantGuardianMFAPage {
	var mfa management.TenantGuardianMFAPage

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		mfa.Enabled = value.Bool(d.GetAttr("enabled"))
		mfa.HTML = value.String(d.GetAttr("html"))
		return stop
	})

	if mfa == (management.TenantGuardianMFAPage{}) {
		return nil
	}

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

	if errorPage == (management.TenantErrorPage{}) {
		return nil
	}

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

	if universalLogin == (management.TenantUniversalLogin{}) {
		return nil
	}

	return &universalLogin
}

func expandTenantSessionCookie(config cty.Value) *management.TenantSessionCookie {
	var sessionCookie management.TenantSessionCookie

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		sessionCookie.Mode = value.String(d.GetAttr("mode"))
		return stop
	})

	if sessionCookie == (management.TenantSessionCookie{}) {
		return nil
	}

	return &sessionCookie
}
