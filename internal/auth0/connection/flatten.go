package connection

import (
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func flattenConnection(data *schema.ResourceData, connection *management.Connection) diag.Diagnostics {
	connectionOptions, diags := flattenConnectionOptions(data, connection.Options)
	if diags.HasError() {
		return diags
	}

	result := multierror.Append(
		data.Set("name", connection.GetName()),
		data.Set("display_name", connection.GetDisplayName()),
		data.Set("is_domain_connection", connection.GetIsDomainConnection()),
		data.Set("strategy", connection.GetStrategy()),
		data.Set("options", connectionOptions),
		data.Set("realms", connection.GetRealms()),
		data.Set("metadata", connection.GetMetadata()),
		data.Set("enabled_clients", connection.GetEnabledClients()),
	)

	switch connection.GetStrategy() {
	case management.ConnectionStrategyGoogleApps,
		management.ConnectionStrategyOIDC,
		management.ConnectionStrategyAD,
		management.ConnectionStrategyAzureAD,
		management.ConnectionStrategySAML,
		management.ConnectionStrategyADFS:
		result = multierror.Append(result, data.Set("show_as_button", connection.GetShowAsButton()))
	}

	return diag.FromErr(result.ErrorOrNil())
}

func flattenConnectionOptions(d *schema.ResourceData, options interface{}) ([]interface{}, diag.Diagnostics) {
	if options == nil {
		return nil, nil
	}

	var m interface{}
	var diags diag.Diagnostics
	switch connectionOptions := options.(type) {
	case *management.ConnectionOptions:
		m, diags = flattenConnectionOptionsAuth0(d, connectionOptions)
	case *management.ConnectionOptionsGoogleOAuth2:
		m, diags = flattenConnectionOptionsGoogleOAuth2(connectionOptions)
	case *management.ConnectionOptionsGoogleApps:
		m, diags = flattenConnectionOptionsGoogleApps(connectionOptions)
	case *management.ConnectionOptionsOAuth2:
		m, diags = flattenConnectionOptionsOAuth2(connectionOptions)
	case *management.ConnectionOptionsFacebook:
		m, diags = flattenConnectionOptionsFacebook(connectionOptions)
	case *management.ConnectionOptionsApple:
		m, diags = flattenConnectionOptionsApple(connectionOptions)
	case *management.ConnectionOptionsLinkedin:
		m, diags = flattenConnectionOptionsLinkedin(connectionOptions)
	case *management.ConnectionOptionsGitHub:
		m, diags = flattenConnectionOptionsGitHub(connectionOptions)
	case *management.ConnectionOptionsWindowsLive:
		m, diags = flattenConnectionOptionsWindowsLive(connectionOptions)
	case *management.ConnectionOptionsSalesforce:
		m, diags = flattenConnectionOptionsSalesforce(connectionOptions)
	case *management.ConnectionOptionsEmail:
		m, diags = flattenConnectionOptionsEmail(connectionOptions)
	case *management.ConnectionOptionsSMS:
		m, diags = flattenConnectionOptionsSMS(connectionOptions)
	case *management.ConnectionOptionsOIDC:
		m, diags = flattenConnectionOptionsOIDC(connectionOptions)
	case *management.ConnectionOptionsOkta:
		m, diags = flattenConnectionOptionsOkta(connectionOptions)
	case *management.ConnectionOptionsAD:
		m, diags = flattenConnectionOptionsAD(connectionOptions)
	case *management.ConnectionOptionsAzureAD:
		m, diags = flattenConnectionOptionsAzureAD(connectionOptions)
	case *management.ConnectionOptionsADFS:
		m, diags = flattenConnectionOptionsADFS(connectionOptions)
	case *management.ConnectionOptionsPingFederate:
		m, diags = flattenConnectionOptionsPingFederate(connectionOptions)
	case *management.ConnectionOptionsSAML:
		m, diags = flattenConnectionOptionsSAML(d, connectionOptions)
	}

	return []interface{}{m}, diags
}

func flattenConnectionOptionsGitHub(options *management.ConnectionOptionsGitHub) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"scopes":                   options.Scopes(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsWindowsLive(options *management.ConnectionOptionsWindowsLive) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"strategy_version":         options.GetStrategyVersion(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsAuth0(
	d *schema.ResourceData,
	options *management.ConnectionOptions,
) (interface{}, diag.Diagnostics) {
	dbSecretConfig, ok := d.GetOk("options.0.configuration")
	if !ok {
		dbSecretConfig = make(map[string]interface{})
	}

	m := map[string]interface{}{
		"password_policy":                      options.GetPasswordPolicy(),
		"enable_script_context":                options.GetEnableScriptContext(),
		"enabled_database_customization":       options.GetEnabledDatabaseCustomization(),
		"brute_force_protection":               options.GetBruteForceProtection(),
		"import_mode":                          options.GetImportMode(),
		"disable_signup":                       options.GetDisableSignup(),
		"disable_self_service_change_password": options.GetDisableSelfServiceChangePassword(),
		"requires_username":                    options.GetRequiresUsername(),
		"custom_scripts":                       options.GetCustomScripts(),
		"configuration":                        dbSecretConfig, // Values do not get read back.
		"non_persistent_attrs":                 options.GetNonPersistentAttrs(),
		"set_user_root_attributes":             options.GetSetUserAttributes(),
	}

	if options.PasswordComplexityOptions != nil {
		m["password_complexity_options"] = []interface{}{options.PasswordComplexityOptions}
	}
	if options.PasswordDictionary != nil {
		m["password_dictionary"] = []interface{}{options.PasswordDictionary}
	}
	if options.PasswordNoPersonalInfo != nil {
		m["password_no_personal_info"] = []interface{}{options.PasswordNoPersonalInfo}
	}
	if options.PasswordHistory != nil {
		m["password_history"] = []interface{}{options.PasswordHistory}
	}
	if options.MFA != nil {
		m["mfa"] = []interface{}{options.MFA}
	}
	if options.Validation != nil {
		m["validation"] = []interface{}{
			map[string]interface{}{
				"username": []interface{}{
					options.Validation["username"],
				},
			},
		}
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

// checkForUnmanagedConfigurationSecrets is used to assess keys diff because values are sent back encrypted.
func checkForUnmanagedConfigurationSecrets(configFromTF, configFromAPI map[string]string) diag.Diagnostics {
	var warnings diag.Diagnostics

	for key := range configFromAPI {
		if _, ok := configFromTF[key]; !ok {
			warnings = append(warnings, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unmanaged Configuration Secret",
				Detail: fmt.Sprintf("Detected a configuration secret not managed through terraform: %q. "+
					"If you proceed, this configuration secret will get deleted. It is required to "+
					"add this configuration secret to your custom database settings to "+
					"prevent unintentionally destructive results.",
					key,
				),
				AttributePath: cty.Path{cty.GetAttrStep{Name: "options.configuration"}},
			})
		}
	}

	return warnings
}

func flattenConnectionOptionsGoogleOAuth2(
	options *management.ConnectionOptionsGoogleOAuth2,
) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"allowed_audiences":        options.GetAllowedAudiences(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsGoogleApps(
	options *management.ConnectionOptionsGoogleApps,
) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":            options.GetClientID(),
		"client_secret":        options.GetClientSecret(),
		"domain":               options.GetDomain(),
		"tenant_domain":        options.GetTenantDomain(),
		"api_enable_users":     options.GetEnableUsersAPI(),
		"scopes":               options.Scopes(),
		"non_persistent_attrs": options.GetNonPersistentAttrs(),
		"domain_aliases":       options.GetDomainAliases(),
		"icon_url":             options.GetLogoURL(),
	}

	m["set_user_root_attributes"] = options.GetSetUserAttributes()
	if options.GetSetUserAttributes() == "" {
		m["set_user_root_attributes"] = "on_each_login"
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsOAuth2(options *management.ConnectionOptionsOAuth2) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"scopes":                   options.Scopes(),
		"token_endpoint":           options.GetTokenURL(),
		"authorization_endpoint":   options.GetAuthorizationURL(),
		"scripts":                  options.GetScripts(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"icon_url":                 options.GetLogoURL(),
		"pkce_enabled":             options.GetPKCEEnabled(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsFacebook(options *management.ConnectionOptionsFacebook) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsApple(options *management.ConnectionOptionsApple) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"team_id":                  options.GetTeamID(),
		"key_id":                   options.GetKeyID(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsLinkedin(options *management.ConnectionOptionsLinkedin) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"strategy_version":         options.GetStrategyVersion(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsSalesforce(options *management.ConnectionOptionsSalesforce) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"community_base_url":       options.GetCommunityBaseURL(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsSMS(options *management.ConnectionOptionsSMS) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"name":                   options.GetName(),
		"from":                   options.GetFrom(),
		"syntax":                 options.GetSyntax(),
		"template":               options.GetTemplate(),
		"twilio_sid":             options.GetTwilioSID(),
		"twilio_token":           options.GetTwilioToken(),
		"messaging_service_sid":  options.GetMessagingServiceSID(),
		"disable_signup":         options.GetDisableSignup(),
		"brute_force_protection": options.GetBruteForceProtection(),
		"provider":               options.GetProvider(),
		"gateway_url":            options.GetGatewayURL(),
		"forward_request_info":   options.GetForwardRequestInfo(),
	}

	if options.OTP != nil {
		m["totp"] = []interface{}{
			map[string]interface{}{
				"time_step": options.OTP.GetTimeStep(),
				"length":    options.OTP.GetLength(),
			},
		}
	}

	if options.GatewayAuthentication != nil {
		m["gateway_authentication"] = []interface{}{
			map[string]interface{}{
				"method":                options.GatewayAuthentication.GetMethod(),
				"subject":               options.GatewayAuthentication.GetSubject(),
				"audience":              options.GatewayAuthentication.GetAudience(),
				"secret":                options.GatewayAuthentication.GetSecret(),
				"secret_base64_encoded": options.GatewayAuthentication.GetSecretBase64Encoded(),
			},
		}
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsOIDC(options *management.ConnectionOptionsOIDC) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"icon_url":                 options.GetLogoURL(),
		"tenant_domain":            options.GetTenantDomain(),
		"domain_aliases":           options.GetDomainAliases(),
		"type":                     options.GetType(),
		"scopes":                   options.Scopes(),
		"issuer":                   options.GetIssuer(),
		"jwks_uri":                 options.GetJWKSURI(),
		"discovery_url":            options.GetDiscoveryURL(),
		"token_endpoint":           options.GetTokenEndpoint(),
		"userinfo_endpoint":        options.GetUserInfoEndpoint(),
		"authorization_endpoint":   options.GetAuthorizationEndpoint(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsOkta(options *management.ConnectionOptionsOkta) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"domain":                   options.GetDomain(),
		"domain_aliases":           options.GetDomainAliases(),
		"scopes":                   options.Scopes(),
		"issuer":                   options.GetIssuer(),
		"jwks_uri":                 options.GetJWKSURI(),
		"token_endpoint":           options.GetTokenEndpoint(),
		"userinfo_endpoint":        options.GetUserInfoEndpoint(),
		"authorization_endpoint":   options.GetAuthorizationEndpoint(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"icon_url":                 options.GetLogoURL(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsEmail(options *management.ConnectionOptionsEmail) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"name":                     options.GetName(),
		"from":                     options.GetEmail().GetFrom(),
		"syntax":                   options.GetEmail().GetSyntax(),
		"subject":                  options.GetEmail().GetSubject(),
		"template":                 options.GetEmail().GetBody(),
		"disable_signup":           options.GetDisableSignup(),
		"brute_force_protection":   options.GetBruteForceProtection(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	if options.OTP != nil {
		m["totp"] = []interface{}{
			map[string]interface{}{
				"time_step": options.OTP.GetTimeStep(),
				"length":    options.OTP.GetLength(),
			},
		}
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	if options.AuthParams != nil {
		v, ok := options.AuthParams.(map[string]interface{})
		if !ok {
			return m, diag.Diagnostics{{
				Severity:      diag.Warning,
				Summary:       "Unable to cast auth_params to map[string]string",
				Detail:        fmt.Sprintf(`Authentication Parameters are required to be a map of strings, the existing value of %v is not compatible. It is recommended to express the existing value as a valid map[string]string. Subsequent terraform applys will clear this configuration to empty map.`, options.AuthParams),
				AttributePath: cty.Path{cty.GetAttrStep{Name: "options.auth_params"}},
			}}
		}
		m["auth_params"] = v
	}

	return m, nil
}

func flattenConnectionOptionsAD(options *management.ConnectionOptionsAD) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"tenant_domain":          options.GetTenantDomain(),
		"domain_aliases":         options.GetDomainAliases(),
		"icon_url":               options.GetLogoURL(),
		"ips":                    options.GetIPs(),
		"use_cert_auth":          options.GetCertAuth(),
		"use_kerberos":           options.GetKerberos(),
		"disable_cache":          options.GetDisableCache(),
		"brute_force_protection": options.GetBruteForceProtection(),
		"non_persistent_attrs":   options.GetNonPersistentAttrs(),
	}

	m["set_user_root_attributes"] = options.GetSetUserAttributes()
	if options.GetSetUserAttributes() == "" {
		m["set_user_root_attributes"] = "on_each_login"
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsAzureAD(options *management.ConnectionOptionsAzureAD) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"client_id":                              options.GetClientID(),
		"client_secret":                          options.GetClientSecret(),
		"app_id":                                 options.GetAppID(),
		"tenant_domain":                          options.GetTenantDomain(),
		"domain":                                 options.GetDomain(),
		"domain_aliases":                         options.GetDomainAliases(),
		"icon_url":                               options.GetLogoURL(),
		"identity_api":                           options.GetIdentityAPI(),
		"waad_protocol":                          options.GetWAADProtocol(),
		"waad_common_endpoint":                   options.GetUseCommonEndpoint(),
		"use_wsfed":                              options.GetUseWSFederation(),
		"api_enable_users":                       options.GetEnableUsersAPI(),
		"max_groups_to_retrieve":                 options.GetMaxGroupsToRetrieve(),
		"scopes":                                 options.Scopes(),
		"non_persistent_attrs":                   options.GetNonPersistentAttrs(),
		"should_trust_email_verified_connection": options.GetTrustEmailVerified(),
	}

	m["set_user_root_attributes"] = options.GetSetUserAttributes()
	if options.GetSetUserAttributes() == "" {
		m["set_user_root_attributes"] = "on_each_login"
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsADFS(options *management.ConnectionOptionsADFS) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"tenant_domain":                          options.GetTenantDomain(),
		"domain_aliases":                         options.GetDomainAliases(),
		"icon_url":                               options.GetLogoURL(),
		"adfs_server":                            options.GetADFSServer(),
		"fed_metadata_xml":                       options.GetFedMetadataXML(),
		"sign_in_endpoint":                       options.GetSignInEndpoint(),
		"api_enable_users":                       options.GetEnableUsersAPI(),
		"should_trust_email_verified_connection": options.GetTrustEmailVerified(),
		"non_persistent_attrs":                   options.GetNonPersistentAttrs(),
	}

	m["set_user_root_attributes"] = options.GetSetUserAttributes()
	if options.GetSetUserAttributes() == "" {
		m["set_user_root_attributes"] = "on_each_login"
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsSAML(
	d *schema.ResourceData,
	options *management.ConnectionOptionsSAML,
) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"signing_cert":         options.GetSigningCert(),
		"protocol_binding":     options.GetProtocolBinding(),
		"debug":                options.GetDebug(),
		"tenant_domain":        options.GetTenantDomain(),
		"domain_aliases":       options.GetDomainAliases(),
		"sign_in_endpoint":     options.GetSignInEndpoint(),
		"sign_out_endpoint":    options.GetSignOutEndpoint(),
		"disable_sign_out":     options.GetDisableSignOut(),
		"signature_algorithm":  options.GetSignatureAlgorithm(),
		"digest_algorithm":     options.GetDigestAglorithm(),
		"sign_saml_request":    options.GetSignSAMLRequest(),
		"icon_url":             options.GetLogoURL(),
		"request_template":     options.GetRequestTemplate(),
		"user_id_attribute":    options.GetUserIDAttribute(),
		"non_persistent_attrs": options.GetNonPersistentAttrs(),
		"entity_id":            options.GetEntityID(),
		"metadata_url":         options.GetMetadataURL(),
		"metadata_xml":         d.Get("options.0.metadata_xml").(string), // Does not get read back.
	}

	m["set_user_root_attributes"] = options.GetSetUserAttributes()
	if options.GetSetUserAttributes() == "" {
		m["set_user_root_attributes"] = "on_each_login"
	}

	if options.IdpInitiated != nil {
		m["idp_initiated"] = []interface{}{
			map[string]interface{}{
				"client_id":              options.IdpInitiated.GetClientID(),
				"client_protocol":        options.IdpInitiated.GetClientProtocol(),
				"client_authorize_query": options.IdpInitiated.GetClientAuthorizeQuery(),
			},
		}
	}

	if options.SigningKey != nil {
		m["signing_key"] = []interface{}{
			map[string]interface{}{
				"key":  options.SigningKey.GetKey(),
				"cert": options.SigningKey.GetCert(),
			},
		}
	}

	fieldsMap, err := structure.FlattenJsonToString(options.FieldsMap)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["fields_map"] = fieldsMap

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsPingFederate(
	options *management.ConnectionOptionsPingFederate,
) (interface{}, diag.Diagnostics) {
	signingCert := options.GetSigningCert()
	if signingCert == "" {
		signingCert = options.GetCert()
	}

	m := map[string]interface{}{
		"signing_cert":           signingCert,
		"tenant_domain":          options.GetTenantDomain(),
		"domain_aliases":         options.GetDomainAliases(),
		"sign_in_endpoint":       options.GetSignInEndpoint(),
		"signature_algorithm":    options.GetSignatureAlgorithm(),
		"digest_algorithm":       options.GetDigestAlgorithm(),
		"sign_saml_request":      options.GetSignSAMLRequest(),
		"ping_federate_base_url": options.GetPingFederateBaseURL(),
		"icon_url":               options.GetLogoURL(),
		"non_persistent_attrs":   options.GetNonPersistentAttrs(),
	}

	m["set_user_root_attributes"] = options.GetSetUserAttributes()
	if options.GetSetUserAttributes() == "" {
		m["set_user_root_attributes"] = "on_each_login"
	}

	m["idp_initiated"] = []interface{}{
		map[string]interface{}{
			"client_id":              options.GetIdpInitiated().GetClientID(),
			"client_protocol":        options.GetIdpInitiated().GetClientProtocol(),
			"client_authorize_query": options.GetIdpInitiated().GetClientAuthorizeQuery(),
		},
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionClient(data *schema.ResourceData, connection *management.Connection) error {
	result := multierror.Append(
		data.Set("name", connection.GetName()),
		data.Set("strategy", connection.GetStrategy()),
	)

	return result.ErrorOrNil()
}

func flattenConnectionClients(data *schema.ResourceData, connection *management.Connection) error {
	result := multierror.Append(
		data.Set("connection_id", connection.GetID()),
		data.Set("name", connection.GetName()),
		data.Set("strategy", connection.GetStrategy()),
		data.Set("enabled_clients", connection.GetEnabledClients()),
	)

	return result.ErrorOrNil()
}
