package provider

import (
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

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
	case *management.ConnectionOptionsAD:
		m, diags = flattenConnectionOptionsAD(connectionOptions)
	case *management.ConnectionOptionsAzureAD:
		m, diags = flattenConnectionOptionsAzureAD(connectionOptions)
	case *management.ConnectionOptionsADFS:
		m, diags = flattenConnectionOptionsADFS(connectionOptions)
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
		"password_policy":                options.GetPasswordPolicy(),
		"enabled_database_customization": options.GetEnabledDatabaseCustomization(),
		"brute_force_protection":         options.GetBruteForceProtection(),
		"import_mode":                    options.GetImportMode(),
		"disable_signup":                 options.GetDisableSignup(),
		"requires_username":              options.GetRequiresUsername(),
		"custom_scripts":                 options.GetCustomScripts(),
		"configuration":                  dbSecretConfig, // Values do not get read back.
		"non_persistent_attrs":           options.GetNonPersistentAttrs(),
		"set_user_root_attributes":       options.GetSetUserAttributes(),
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

	diags := checkForUnmanagedConfigurationSecrets(
		dbSecretConfig.(map[string]interface{}),
		options.GetConfiguration(),
	)

	return m, diags
}

// checkForUnmanagedConfigurationSecrets is used to assess keys diff because values are sent back encrypted.
func checkForUnmanagedConfigurationSecrets(configFromTF map[string]interface{}, configFromAPI map[string]string) diag.Diagnostics {
	var warnings diag.Diagnostics

	for key := range configFromAPI {
		if _, ok := configFromTF[key]; !ok {
			warnings = append(warnings, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unmanaged Configuration Secret",
				Detail: fmt.Sprintf("Detected a configuration secret not managed though terraform: %q. "+
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
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"domain":                   options.GetDomain(),
		"tenant_domain":            options.GetTenantDomain(),
		"api_enable_users":         options.GetEnableUsersAPI(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"domain_aliases":           options.GetDomainAliases(),
		"icon_url":                 options.GetLogoURL(),
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
		"tenant_domain":            options.GetTenantDomain(),
		"domain_aliases":           options.GetDomainAliases(),
		"icon_url":                 options.GetLogoURL(),
		"ips":                      options.GetIPs(),
		"use_cert_auth":            options.GetCertAuth(),
		"use_kerberos":             options.GetKerberos(),
		"disable_cache":            options.GetDisableCache(),
		"brute_force_protection":   options.GetBruteForceProtection(),
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
		"set_user_root_attributes":               options.GetSetUserAttributes(),
		"non_persistent_attrs":                   options.GetNonPersistentAttrs(),
		"should_trust_email_verified_connection": options.GetTrustEmailVerified(),
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
		"tenant_domain":            options.GetTenantDomain(),
		"domain_aliases":           options.GetDomainAliases(),
		"icon_url":                 options.GetLogoURL(),
		"adfs_server":              options.GetADFSServer(),
		"api_enable_users":         options.GetEnableUsersAPI(),
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

func flattenConnectionOptionsSAML(
	d *schema.ResourceData,
	options *management.ConnectionOptionsSAML,
) (interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"signing_cert":             options.GetSigningCert(),
		"protocol_binding":         options.GetProtocolBinding(),
		"debug":                    options.GetDebug(),
		"tenant_domain":            options.GetTenantDomain(),
		"domain_aliases":           options.GetDomainAliases(),
		"sign_in_endpoint":         options.GetSignInEndpoint(),
		"sign_out_endpoint":        options.GetSignOutEndpoint(),
		"disable_sign_out":         options.GetDisableSignOut(),
		"signature_algorithm":      options.GetSignatureAlgorithm(),
		"digest_algorithm":         options.GetDigestAglorithm(),
		"sign_saml_request":        options.GetSignSAMLRequest(),
		"icon_url":                 options.GetLogoURL(),
		"request_template":         options.GetRequestTemplate(),
		"user_id_attribute":        options.GetUserIDAttribute(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"entity_id":                options.GetEntityID(),
		"metadata_url":             options.GetMetadataURL(),
		"metadata_xml":             d.Get("options.0.metadata_xml").(string), // Does not get read back.
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

func expandConnection(d *schema.ResourceData) (*management.Connection, diag.Diagnostics) {
	config := d.GetRawConfig()

	connection := &management.Connection{
		DisplayName:        value.String(config.GetAttr("display_name")),
		IsDomainConnection: value.Bool(config.GetAttr("is_domain_connection")),
		EnabledClients:     value.Strings(config.GetAttr("enabled_clients")),
		Metadata:           value.MapOfStrings(config.GetAttr("metadata")),
	}

	if d.IsNewResource() {
		connection.Name = value.String(config.GetAttr("name"))
		connection.Strategy = value.String(config.GetAttr("strategy"))
	}

	if d.IsNewResource() || d.HasChange("realms") {
		connection.Realms = value.Strings(config.GetAttr("realms"))
	}

	var diagnostics diag.Diagnostics
	strategy := d.Get("strategy").(string)
	showAsButton := value.Bool(config.GetAttr("show_as_button"))

	config.GetAttr("options").ForEachElement(func(_ cty.Value, options cty.Value) (stop bool) {
		switch strategy {
		case management.ConnectionStrategyAuth0:
			connection.Options, diagnostics = expandConnectionOptionsAuth0(options)
		case management.ConnectionStrategyGoogleOAuth2:
			connection.Options, diagnostics = expandConnectionOptionsGoogleOAuth2(d, options)
		case management.ConnectionStrategyGoogleApps:
			connection.ShowAsButton = showAsButton
			connection.Options, diagnostics = expandConnectionOptionsGoogleApps(d, options)
		case management.ConnectionStrategyOAuth2,
			management.ConnectionStrategyDropbox,
			management.ConnectionStrategyBitBucket,
			management.ConnectionStrategyPaypal,
			management.ConnectionStrategyTwitter,
			management.ConnectionStrategyAmazon,
			management.ConnectionStrategyYahoo,
			management.ConnectionStrategyBox,
			management.ConnectionStrategyWordpress,
			management.ConnectionStrategyDiscord,
			management.ConnectionStrategyImgur,
			management.ConnectionStrategySpotify,
			management.ConnectionStrategyShopify,
			management.ConnectionStrategyFigma,
			management.ConnectionStrategySlack,
			management.ConnectionStrategyDigitalOcean,
			management.ConnectionStrategyTwitch,
			management.ConnectionStrategyVimeo,
			management.ConnectionStrategyCustom:
			connection.Options, diagnostics = expandConnectionOptionsOAuth2(d, options)
		case management.ConnectionStrategyFacebook:
			connection.Options, diagnostics = expandConnectionOptionsFacebook(d, options)
		case management.ConnectionStrategyApple:
			connection.Options, diagnostics = expandConnectionOptionsApple(d, options)
		case management.ConnectionStrategyLinkedin:
			connection.Options, diagnostics = expandConnectionOptionsLinkedin(d, options)
		case management.ConnectionStrategyGitHub:
			connection.Options, diagnostics = expandConnectionOptionsGitHub(d, options)
		case management.ConnectionStrategyWindowsLive:
			connection.Options, diagnostics = expandConnectionOptionsWindowsLive(d, options)
		case management.ConnectionStrategySalesforce,
			management.ConnectionStrategySalesforceCommunity,
			management.ConnectionStrategySalesforceSandbox:
			connection.Options, diagnostics = expandConnectionOptionsSalesforce(d, options)
		case management.ConnectionStrategySMS:
			connection.Options, diagnostics = expandConnectionOptionsSMS(options)
		case management.ConnectionStrategyOIDC:
			connection.ShowAsButton = showAsButton
			connection.Options, diagnostics = expandConnectionOptionsOIDC(d, options)
		case management.ConnectionStrategyAD:
			connection.ShowAsButton = showAsButton
			connection.Options, diagnostics = expandConnectionOptionsAD(options)
		case management.ConnectionStrategyAzureAD:
			connection.ShowAsButton = showAsButton
			connection.Options, diagnostics = expandConnectionOptionsAzureAD(d, options)
		case management.ConnectionStrategyEmail:
			connection.Options, diagnostics = expandConnectionOptionsEmail(options)
		case management.ConnectionStrategySAML:
			connection.ShowAsButton = showAsButton
			connection.Options, diagnostics = expandConnectionOptionsSAML(options)
		case management.ConnectionStrategyADFS:
			connection.ShowAsButton = showAsButton
			connection.Options, diagnostics = expandConnectionOptionsADFS(options)
		default:
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unsupported Connection Strategy",
				Detail: fmt.Sprintf(
					"Raise an issue at %s in order to have the following connection strategy supported: %q",
					"https://github.com/auth0/terraform-provider-auth0/issues/new",
					strategy,
				),
				AttributePath: cty.Path{cty.GetAttrStep{Name: "strategy"}},
			})
		}

		return stop
	})

	return connection, diagnostics
}

func expandConnectionOptionsGitHub(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsGitHub, diag.Diagnostics) {
	options := &management.ConnectionOptionsGitHub{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsAuth0(config cty.Value) (*management.ConnectionOptions, diag.Diagnostics) {
	options := &management.ConnectionOptions{
		PasswordPolicy:               value.String(config.GetAttr("password_policy")),
		NonPersistentAttrs:           value.Strings(config.GetAttr("non_persistent_attrs")),
		SetUserAttributes:            value.String(config.GetAttr("set_user_root_attributes")),
		EnabledDatabaseCustomization: value.Bool(config.GetAttr("enabled_database_customization")),
		BruteForceProtection:         value.Bool(config.GetAttr("brute_force_protection")),
		ImportMode:                   value.Bool(config.GetAttr("import_mode")),
		DisableSignup:                value.Bool(config.GetAttr("disable_signup")),
		RequiresUsername:             value.Bool(config.GetAttr("requires_username")),
		CustomScripts:                value.MapOfStrings(config.GetAttr("custom_scripts")),
		Configuration:                value.MapOfStrings(config.GetAttr("configuration")),
	}

	config.GetAttr("validation").ForEachElement(
		func(_ cty.Value, validation cty.Value) (stop bool) {
			validationOption := make(map[string]interface{})

			validation.GetAttr("username").ForEachElement(
				func(_ cty.Value, username cty.Value) (stop bool) {
					usernameValidation := make(map[string]*int)

					if min := value.Int(username.GetAttr("min")); min != nil {
						usernameValidation["min"] = min
					}
					if max := value.Int(username.GetAttr("max")); max != nil {
						usernameValidation["max"] = max
					}

					if len(usernameValidation) > 0 {
						validationOption["username"] = usernameValidation
					}

					return stop
				},
			)

			if len(validationOption) > 0 {
				options.Validation = validationOption
			}

			return stop
		},
	)

	config.GetAttr("password_history").ForEachElement(
		func(_ cty.Value, passwordHistory cty.Value) (stop bool) {
			passwordHistoryOption := make(map[string]interface{})

			if enable := value.Bool(passwordHistory.GetAttr("enable")); enable != nil {
				passwordHistoryOption["enable"] = enable
			}

			if size := value.Int(passwordHistory.GetAttr("size")); size != nil && *size != 0 {
				passwordHistoryOption["size"] = size
			}

			if len(passwordHistoryOption) > 0 {
				options.PasswordHistory = passwordHistoryOption
			}

			return stop
		},
	)

	config.GetAttr("password_no_personal_info").ForEachElement(
		func(_ cty.Value, passwordNoPersonalInfo cty.Value) (stop bool) {
			if enable := value.Bool(passwordNoPersonalInfo.GetAttr("enable")); enable != nil {
				options.PasswordNoPersonalInfo = map[string]interface{}{
					"enable": enable,
				}
			}

			return stop
		},
	)

	config.GetAttr("password_dictionary").ForEachElement(
		func(_ cty.Value, passwordDictionary cty.Value) (stop bool) {
			passwordDictionaryOption := make(map[string]interface{})

			if enable := value.Bool(passwordDictionary.GetAttr("enable")); enable != nil {
				passwordDictionaryOption["enable"] = enable
			}
			if dictionary := value.Strings(passwordDictionary.GetAttr("dictionary")); dictionary != nil {
				passwordDictionaryOption["dictionary"] = dictionary
			}

			if len(passwordDictionaryOption) > 0 {
				options.PasswordDictionary = passwordDictionaryOption
			}

			return stop
		},
	)

	config.GetAttr("password_complexity_options").ForEachElement(
		func(_ cty.Value, passwordComplexity cty.Value) (stop bool) {
			if minLength := value.Int(passwordComplexity.GetAttr("min_length")); minLength != nil {
				options.PasswordComplexityOptions = map[string]interface{}{
					"min_length": minLength,
				}
			}

			return stop
		},
	)

	config.GetAttr("mfa").ForEachElement(
		func(_ cty.Value, mfa cty.Value) (stop bool) {
			mfaOption := make(map[string]interface{})

			if active := value.Bool(mfa.GetAttr("active")); active != nil {
				mfaOption["active"] = active
			}
			if returnEnrollSettings := value.Bool(mfa.GetAttr("return_enroll_settings")); returnEnrollSettings != nil {
				mfaOption["return_enroll_settings"] = returnEnrollSettings
			}

			if len(mfaOption) > 0 {
				options.MFA = mfaOption
			}

			return stop
		},
	)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsGoogleOAuth2(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsGoogleOAuth2, diag.Diagnostics) {
	options := &management.ConnectionOptionsGoogleOAuth2{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		AllowedAudiences:   value.Strings(config.GetAttr("allowed_audiences")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsGoogleApps(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsGoogleApps, diag.Diagnostics) {
	options := &management.ConnectionOptionsGoogleApps{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		Domain:             value.String(config.GetAttr("domain")),
		TenantDomain:       value.String(config.GetAttr("tenant_domain")),
		EnableUsersAPI:     value.Bool(config.GetAttr("api_enable_users")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
		DomainAliases:      value.Strings(config.GetAttr("domain_aliases")),
		LogoURL:            value.String(config.GetAttr("icon_url")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsOAuth2(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsOAuth2, diag.Diagnostics) {
	options := &management.ConnectionOptionsOAuth2{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		AuthorizationURL:   value.String(config.GetAttr("authorization_endpoint")),
		TokenURL:           value.String(config.GetAttr("token_endpoint")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
		LogoURL:            value.String(config.GetAttr("icon_url")),
		PKCEEnabled:        value.Bool(config.GetAttr("pkce_enabled")),
		Scripts:            value.MapOfStrings(config.GetAttr("scripts")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsFacebook(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsFacebook, diag.Diagnostics) {
	options := &management.ConnectionOptionsFacebook{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsApple(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsApple, diag.Diagnostics) {
	options := &management.ConnectionOptionsApple{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		TeamID:             value.String(config.GetAttr("team_id")),
		KeyID:              value.String(config.GetAttr("key_id")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsLinkedin(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsLinkedin, diag.Diagnostics) {
	options := &management.ConnectionOptionsLinkedin{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		StrategyVersion:    value.Int(config.GetAttr("strategy_version")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsSalesforce(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsSalesforce, diag.Diagnostics) {
	options := &management.ConnectionOptionsSalesforce{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		CommunityBaseURL:   value.String(config.GetAttr("community_base_url")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsWindowsLive(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsWindowsLive, diag.Diagnostics) {
	options := &management.ConnectionOptionsWindowsLive{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		StrategyVersion:    value.Int(config.GetAttr("strategy_version")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsSMS(config cty.Value) (*management.ConnectionOptionsSMS, diag.Diagnostics) {
	options := &management.ConnectionOptionsSMS{
		Name:                 value.String(config.GetAttr("name")),
		From:                 value.String(config.GetAttr("from")),
		Syntax:               value.String(config.GetAttr("syntax")),
		Template:             value.String(config.GetAttr("template")),
		TwilioSID:            value.String(config.GetAttr("twilio_sid")),
		TwilioToken:          value.String(config.GetAttr("twilio_token")),
		MessagingServiceSID:  value.String(config.GetAttr("messaging_service_sid")),
		Provider:             value.String(config.GetAttr("provider")),
		GatewayURL:           value.String(config.GetAttr("gateway_url")),
		ForwardRequestInfo:   value.Bool(config.GetAttr("forward_request_info")),
		DisableSignup:        value.Bool(config.GetAttr("disable_signup")),
		BruteForceProtection: value.Bool(config.GetAttr("brute_force_protection")),
	}

	config.GetAttr("totp").ForEachElement(func(_ cty.Value, totp cty.Value) (stop bool) {
		options.OTP = &management.ConnectionOptionsOTP{
			TimeStep: value.Int(totp.GetAttr("time_step")),
			Length:   value.Int(totp.GetAttr("length")),
		}

		return stop
	})

	config.GetAttr("gateway_authentication").ForEachElement(func(_ cty.Value, auth cty.Value) (stop bool) {
		options.GatewayAuthentication = &management.ConnectionGatewayAuthentication{
			Method:              value.String(auth.GetAttr("method")),
			Subject:             value.String(auth.GetAttr("subject")),
			Audience:            value.String(auth.GetAttr("audience")),
			Secret:              value.String(auth.GetAttr("secret")),
			SecretBase64Encoded: value.Bool(auth.GetAttr("secret_base64_encoded")),
		}

		return stop
	})

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsEmail(config cty.Value) (*management.ConnectionOptionsEmail, diag.Diagnostics) {
	options := &management.ConnectionOptionsEmail{
		Name:          value.String(config.GetAttr("name")),
		DisableSignup: value.Bool(config.GetAttr("disable_signup")),
		Email: &management.ConnectionOptionsEmailSettings{
			Syntax:  value.String(config.GetAttr("syntax")),
			From:    value.String(config.GetAttr("from")),
			Subject: value.String(config.GetAttr("subject")),
			Body:    value.String(config.GetAttr("template")),
		},
		BruteForceProtection: value.Bool(config.GetAttr("brute_force_protection")),
		SetUserAttributes:    value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs:   value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	config.GetAttr("totp").ForEachElement(func(_ cty.Value, totp cty.Value) (stop bool) {
		options.OTP = &management.ConnectionOptionsOTP{
			TimeStep: value.Int(totp.GetAttr("time_step")),
			Length:   value.Int(totp.GetAttr("length")),
		}

		return stop
	})

	if authParamsMap := value.MapOfStrings(config.GetAttr("auth_params")); authParamsMap != nil {
		options.AuthParams = authParamsMap
	}

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsAD(config cty.Value) (*management.ConnectionOptionsAD, diag.Diagnostics) {
	options := &management.ConnectionOptionsAD{
		DomainAliases:        value.Strings(config.GetAttr("domain_aliases")),
		TenantDomain:         value.String(config.GetAttr("tenant_domain")),
		LogoURL:              value.String(config.GetAttr("icon_url")),
		IPs:                  value.Strings(config.GetAttr("ips")),
		CertAuth:             value.Bool(config.GetAttr("use_cert_auth")),
		Kerberos:             value.Bool(config.GetAttr("use_kerberos")),
		DisableCache:         value.Bool(config.GetAttr("disable_cache")),
		SetUserAttributes:    value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs:   value.Strings(config.GetAttr("non_persistent_attrs")),
		BruteForceProtection: value.Bool(config.GetAttr("brute_force_protection")),
	}

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsAzureAD(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsAzureAD, diag.Diagnostics) {
	options := &management.ConnectionOptionsAzureAD{
		ClientID:            value.String(config.GetAttr("client_id")),
		ClientSecret:        value.String(config.GetAttr("client_secret")),
		AppID:               value.String(config.GetAttr("app_id")),
		Domain:              value.String(config.GetAttr("domain")),
		DomainAliases:       value.Strings(config.GetAttr("domain_aliases")),
		TenantDomain:        value.String(config.GetAttr("tenant_domain")),
		MaxGroupsToRetrieve: value.String(config.GetAttr("max_groups_to_retrieve")),
		UseWSFederation:     value.Bool(config.GetAttr("use_wsfed")),
		WAADProtocol:        value.String(config.GetAttr("waad_protocol")),
		UseCommonEndpoint:   value.Bool(config.GetAttr("waad_common_endpoint")),
		EnableUsersAPI:      value.Bool(config.GetAttr("api_enable_users")),
		LogoURL:             value.String(config.GetAttr("icon_url")),
		IdentityAPI:         value.String(config.GetAttr("identity_api")),
		SetUserAttributes:   value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs:  value.Strings(config.GetAttr("non_persistent_attrs")),
		TrustEmailVerified:  value.String(config.GetAttr("should_trust_email_verified_connection")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsOIDC(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsOIDC, diag.Diagnostics) {
	options := &management.ConnectionOptionsOIDC{
		ClientID:              value.String(config.GetAttr("client_id")),
		ClientSecret:          value.String(config.GetAttr("client_secret")),
		TenantDomain:          value.String(config.GetAttr("tenant_domain")),
		DomainAliases:         value.Strings(config.GetAttr("domain_aliases")),
		LogoURL:               value.String(config.GetAttr("icon_url")),
		DiscoveryURL:          value.String(config.GetAttr("discovery_url")),
		AuthorizationEndpoint: value.String(config.GetAttr("authorization_endpoint")),
		Issuer:                value.String(config.GetAttr("issuer")),
		JWKSURI:               value.String(config.GetAttr("jwks_uri")),
		Type:                  value.String(config.GetAttr("type")),
		UserInfoEndpoint:      value.String(config.GetAttr("userinfo_endpoint")),
		TokenEndpoint:         value.String(config.GetAttr("token_endpoint")),
		SetUserAttributes:     value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs:    value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsSAML(config cty.Value) (*management.ConnectionOptionsSAML, diag.Diagnostics) {
	options := &management.ConnectionOptionsSAML{
		Debug:              value.Bool(config.GetAttr("debug")),
		SigningCert:        value.String(config.GetAttr("signing_cert")),
		ProtocolBinding:    value.String(config.GetAttr("protocol_binding")),
		TenantDomain:       value.String(config.GetAttr("tenant_domain")),
		DomainAliases:      value.Strings(config.GetAttr("domain_aliases")),
		SignInEndpoint:     value.String(config.GetAttr("sign_in_endpoint")),
		SignOutEndpoint:    value.String(config.GetAttr("sign_out_endpoint")),
		DisableSignOut:     value.Bool(config.GetAttr("disable_sign_out")),
		SignatureAlgorithm: value.String(config.GetAttr("signature_algorithm")),
		DigestAglorithm:    value.String(config.GetAttr("digest_algorithm")),
		SignSAMLRequest:    value.Bool(config.GetAttr("sign_saml_request")),
		RequestTemplate:    value.String(config.GetAttr("request_template")),
		UserIDAttribute:    value.String(config.GetAttr("user_id_attribute")),
		LogoURL:            value.String(config.GetAttr("icon_url")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
		EntityID:           value.String(config.GetAttr("entity_id")),
		MetadataXML:        value.String(config.GetAttr("metadata_xml")),
		MetadataURL:        value.String(config.GetAttr("metadata_url")),
	}

	config.GetAttr("idp_initiated").ForEachElement(func(_ cty.Value, idp cty.Value) (stop bool) {
		options.IdpInitiated = &management.ConnectionOptionsSAMLIdpInitiated{
			ClientID:             value.String(idp.GetAttr("client_id")),
			ClientProtocol:       value.String(idp.GetAttr("client_protocol")),
			ClientAuthorizeQuery: value.String(idp.GetAttr("client_authorize_query")),
		}

		return stop
	})

	config.GetAttr("signing_key").ForEachElement(func(_ cty.Value, key cty.Value) (stop bool) {
		options.SigningKey = &management.ConnectionOptionsSAMLSigningKey{
			Cert: value.String(key.GetAttr("cert")),
			Key:  value.String(key.GetAttr("key")),
		}

		return stop
	})

	var err error

	options.FieldsMap, err = value.MapFromJSON(config.GetAttr("fields_map"))
	diagnostics := diag.FromErr(err)

	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))
	diagnostics = append(diagnostics, diag.FromErr(err)...)

	return options, diagnostics
}

func expandConnectionOptionsADFS(config cty.Value) (*management.ConnectionOptionsADFS, diag.Diagnostics) {
	options := &management.ConnectionOptionsADFS{
		TenantDomain:       value.String(config.GetAttr("tenant_domain")),
		DomainAliases:      value.Strings(config.GetAttr("domain_aliases")),
		LogoURL:            value.String(config.GetAttr("icon_url")),
		ADFSServer:         value.String(config.GetAttr("adfs_server")),
		EnableUsersAPI:     value.Bool(config.GetAttr("api_enable_users")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

type scoper interface {
	Scopes() []string
	SetScopes(enable bool, scopes ...string)
}

func expandConnectionOptionsScopes(d *schema.ResourceData, s scoper) {
	scopesList := Set(d, "options.0.scopes").List()
	_, scopesToDisable := Diff(d, "options.0.scopes")
	for _, scope := range scopesList {
		s.SetScopes(true, scope.(string))
	}
	for _, scope := range scopesToDisable.List() {
		s.SetScopes(false, scope.(string))
	}
}
