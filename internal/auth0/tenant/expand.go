package tenant

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandTenant(data *schema.ResourceData) *management.Tenant {
	config := data.GetRawConfig()

	sessionLifetime := data.Get("session_lifetime").(float64)          // Handling separately to preserve default values not honored by `d.GetRawConfig()`.
	idleSessionLifetime := data.Get("idle_session_lifetime").(float64) // Handling separately to preserve default values not honored by `d.GetRawConfig()`.

	tenant := management.Tenant{
		DefaultAudience:                      value.String(config.GetAttr("default_audience")),
		DefaultDirectory:                     value.String(config.GetAttr("default_directory")),
		DefaultRedirectionURI:                value.String(config.GetAttr("default_redirection_uri")),
		FriendlyName:                         value.String(config.GetAttr("friendly_name")),
		PictureURL:                           value.String(config.GetAttr("picture_url")),
		SupportEmail:                         value.String(config.GetAttr("support_email")),
		SupportURL:                           value.String(config.GetAttr("support_url")),
		AllowedLogoutURLs:                    value.Strings(config.GetAttr("allowed_logout_urls")),
		SessionLifetime:                      &sessionLifetime,
		SandboxVersion:                       value.String(config.GetAttr("sandbox_version")),
		EnabledLocales:                       value.Strings(config.GetAttr("enabled_locales")),
		Flags:                                expandTenantFlags(config.GetAttr("flags")),
		SessionCookie:                        expandTenantSessionCookie(config.GetAttr("session_cookie")),
		Sessions:                             expandTenantSessions(config.GetAttr("sessions")),
		OIDCLogout:                           expandTenantOIDCLogout(config.GetAttr("oidc_logout")),
		AllowOrgNameInAuthAPI:                value.Bool(config.GetAttr("allow_organization_name_in_authentication_api")),
		CustomizeMFAInPostLoginAction:        value.Bool(config.GetAttr("customize_mfa_in_postlogin_action")),
		PushedAuthorizationRequestsSupported: value.Bool(config.GetAttr("pushed_authorization_requests_supported")),
		ACRValuesSupported:                   expandACRValuesSupported(data),
		MTLS:                                 expandMTLSConfiguration(data),
	}

	if data.IsNewResource() || data.HasChange("idle_session_lifetime") {
		tenant.IdleSessionLifetime = &idleSessionLifetime
	}

	return &tenant
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
			EnableSSO:                          value.Bool(flags.GetAttr("enable_sso")),
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
			MFAShowFactorListOnEnrollment:      value.Bool(flags.GetAttr("mfa_show_factor_list_on_enrollment")),
			RemoveAlgFromJWKS:                  value.Bool(flags.GetAttr("remove_alg_from_jwks")),
		}

		return stop
	})

	return tenantFlags
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

func expandTenantSessions(config cty.Value) *management.TenantSessions {
	var sessions management.TenantSessions

	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		sessions.OIDCLogoutPromptEnabled = value.Bool(cfg.GetAttr("oidc_logout_prompt_enabled"))
		return stop
	})

	if sessions == (management.TenantSessions{}) {
		return nil
	}

	return &sessions
}

func expandTenantOIDCLogout(config cty.Value) *management.TenantOIDCLogout {
	var oidcLogout management.TenantOIDCLogout

	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		oidcLogout.OIDCResourceProviderLogoutEndSessionEndpointDiscovery = value.Bool(cfg.GetAttr("rp_logout_end_session_endpoint_discovery"))
		return stop
	})

	if oidcLogout == (management.TenantOIDCLogout{}) {
		return nil
	}

	return &oidcLogout
}

func isACRValuesSupportedNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("disable_acr_values_supported") && !data.HasChange("acr_values_supported") {
		return false
	}
	disable := data.GetRawConfig().GetAttr("disable_acr_values_supported")
	return !disable.IsNull() && disable.True()
}

func expandACRValuesSupported(data *schema.ResourceData) *[]string {
	if !data.IsNewResource() && !data.HasChange("disable_acr_values_supported") && !data.HasChange("acr_values_supported") {
		return nil
	} else if isACRValuesSupportedNull(data) {
		return nil
	}

	return value.Strings(data.GetRawConfig().GetAttr("acr_values_supported"))
}

func isMTLSConfigurationNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("mtls") {
		return false
	}
	empty := true

	config := data.GetRawConfig().GetAttr("mtls")
	if config.IsNull() || config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		disable := cfg.GetAttr("disable")
		if !disable.IsNull() && disable.True() {
			stop = true
		} else {
			empty = false
		}
		return stop
	}) {
		// We forced an early return because it was disabled.
		return true
	}

	return empty
}

func expandMTLSConfiguration(data *schema.ResourceData) *management.TenantMTLSConfiguration {
	if !data.IsNewResource() && !data.HasChange("mtls") {
		return nil
	} else if isMTLSConfigurationNull(data) {
		return nil
	}
	var mtls management.TenantMTLSConfiguration

	config := data.GetRawConfig().GetAttr("mtls")
	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		mtls.EnableEndpointAliases = value.Bool(cfg.GetAttr("enable_endpoint_aliases"))
		return stop
	})

	if mtls == (management.TenantMTLSConfiguration{}) {
		return nil
	}

	return &mtls
}
