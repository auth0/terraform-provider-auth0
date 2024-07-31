package tenant

import (
	"encoding/json"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandTenant(data *schema.ResourceData) (interface{}, error) {
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
		AllowOrgNameInAuthAPI:                value.Bool(config.GetAttr("allow_organization_name_in_authentication_api")),
		CustomizeMFAInPostLoginAction:        value.Bool(config.GetAttr("customize_mfa_in_postlogin_action")),
		PushedAuthorizationRequestsSupported: value.Bool(config.GetAttr("pushed_authorization_requests_supported")),
	}

	if data.IsNewResource() || data.HasChange("idle_session_lifetime") {
		tenant.IdleSessionLifetime = &idleSessionLifetime
	}
	mtls := config.GetAttr("mtls")

	disableACRValuesSupported := config.GetAttr("disable_acr_values_supported")
	if tenantJSON, err := json.Marshal(tenant); err != nil {
		return nil, err
	} else if !disableACRValuesSupported.IsNull() && disableACRValuesSupported.True() {
		if isMTLSConfigurationDisabled(mtls) {
			nilableTenant := tenantNilACRValuesSupportedMTLS{}
			if err = json.Unmarshal(tenantJSON, &nilableTenant); err != nil {
				return nil, err
			}
			nilableTenant.ACRValuesSupported = nil
			nilableTenant.MTLS = nil

			return nilableTenant, nil
		}
		nilableTenant := tenantNilACRValuesSupported{}
		if err = json.Unmarshal(tenantJSON, &nilableTenant); err != nil {
			return nil, err
		}
		nilableTenant.ACRValuesSupported = nil
		nilableTenant.MTLS = expandMTLSConfiguration(mtls)
		return nilableTenant, nil
	} else if isMTLSConfigurationDisabled(mtls) {
		nilableTenant := tenantNilMTLS{}
		if err = json.Unmarshal(tenantJSON, &nilableTenant); err != nil {
			return nil, err
		}
		nilableTenant.ACRValuesSupported = value.Strings(config.GetAttr("acr_values_supported"))
		nilableTenant.MTLS = nil
		return nilableTenant, nil
	}
	tenant.ACRValuesSupported = value.Strings(config.GetAttr("acr_values_supported"))
	tenant.MTLS = expandMTLSConfiguration(mtls)

	return tenant, nil
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

func isMTLSConfigurationDisabled(config cty.Value) bool {
	return !config.IsNull() && config.LengthInt() > 0 && config.AsValueSlice()[0].GetAttr("disable").True()
}

func expandMTLSConfiguration(config cty.Value) *management.TenantMTLSConfiguration {
	var mtls management.TenantMTLSConfiguration
	if config.IsNull() || config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		disabled := cfg.GetAttr("disable")
		if !disabled.IsNull() && disabled.True() {
			// Force it to exit the ForEachElement and return nil.
			stop = true
		} else {
			mtls.EnableEndpointAliases = value.Bool(cfg.GetAttr("enable_endpoint_aliases"))
		}
		return stop
	}) {
		return nil
	}
	if mtls == (management.TenantMTLSConfiguration{}) {
		return nil
	}
	return &mtls
}
