package auth0

import (
	"log"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func flattenConnectionOptions(d ResourceData, options interface{}) ([]interface{}, error) {
	if options == nil {
		return nil, nil
	}

	var m interface{}
	var err error
	switch connectionOptions := options.(type) {
	case *management.ConnectionOptions:
		m, err = flattenConnectionOptionsAuth0(d, connectionOptions)
	case *management.ConnectionOptionsGoogleOAuth2:
		m, err = flattenConnectionOptionsGoogleOAuth2(connectionOptions)
	case *management.ConnectionOptionsGoogleApps:
		m, err = flattenConnectionOptionsGoogleApps(connectionOptions)
	case *management.ConnectionOptionsOAuth2:
		m, err = flattenConnectionOptionsOAuth2(connectionOptions)
	case *management.ConnectionOptionsFacebook:
		m, err = flattenConnectionOptionsFacebook(connectionOptions)
	case *management.ConnectionOptionsApple:
		m, err = flattenConnectionOptionsApple(connectionOptions)
	case *management.ConnectionOptionsLinkedin:
		m, err = flattenConnectionOptionsLinkedin(connectionOptions)
	case *management.ConnectionOptionsGitHub:
		m, err = flattenConnectionOptionsGitHub(connectionOptions)
	case *management.ConnectionOptionsWindowsLive:
		m, err = flattenConnectionOptionsWindowsLive(connectionOptions)
	case *management.ConnectionOptionsSalesforce:
		m, err = flattenConnectionOptionsSalesforce(connectionOptions)
	case *management.ConnectionOptionsEmail:
		m, err = flattenConnectionOptionsEmail(connectionOptions)
	case *management.ConnectionOptionsSMS:
		m, err = flattenConnectionOptionsSMS(connectionOptions)
	case *management.ConnectionOptionsOIDC:
		m, err = flattenConnectionOptionsOIDC(connectionOptions)
	case *management.ConnectionOptionsAD:
		m, err = flattenConnectionOptionsAD(connectionOptions)
	case *management.ConnectionOptionsAzureAD:
		m, err = flattenConnectionOptionsAzureAD(connectionOptions)
	case *management.ConnectionOptionsADFS:
		m, err = flattenConnectionOptionsADFS(connectionOptions)
	case *management.ConnectionOptionsSAML:
		m, err = flattenConnectionOptionsSAML(d, connectionOptions)
	}
	if err != nil {
		return nil, err
	}

	return []interface{}{m}, nil
}

func flattenConnectionOptionsGitHub(options *management.ConnectionOptionsGitHub) (interface{}, error) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"scopes":                   options.Scopes(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsWindowsLive(options *management.ConnectionOptionsWindowsLive) (interface{}, error) {
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
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsAuth0(d ResourceData, options *management.ConnectionOptions) (interface{}, error) {
	m := map[string]interface{}{
		"password_policy":                options.GetPasswordPolicy(),
		"enabled_database_customization": options.GetEnabledDatabaseCustomization(),
		"brute_force_protection":         options.GetBruteForceProtection(),
		"import_mode":                    options.GetImportMode(),
		"disable_signup":                 options.GetDisableSignup(),
		"requires_username":              options.GetRequiresUsername(),
		"custom_scripts":                 options.CustomScripts,
		"configuration":                  Map(d, "options.0.configuration"), // does not get read back
		"non_persistent_attrs":           options.GetNonPersistentAttrs(),
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
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsGoogleOAuth2(options *management.ConnectionOptionsGoogleOAuth2) (interface{}, error) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"allowed_audiences":        options.AllowedAudiences,
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsGoogleApps(options *management.ConnectionOptionsGoogleApps) (interface{}, error) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"domain":                   options.GetDomain(),
		"tenant_domain":            options.GetTenantDomain(),
		"api_enable_users":         options.GetEnableUsersAPI(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"domain_aliases":           options.DomainAliases,
		"icon_url":                 options.GetLogoURL(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsOAuth2(options *management.ConnectionOptionsOAuth2) (interface{}, error) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"scopes":                   options.Scopes(),
		"token_endpoint":           options.GetTokenURL(),
		"authorization_endpoint":   options.GetAuthorizationURL(),
		"scripts":                  options.Scripts,
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"icon_url":                 options.GetLogoURL(),
		"pkce_enabled":             options.GetPKCEEnabled(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsFacebook(options *management.ConnectionOptionsFacebook) (interface{}, error) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsApple(options *management.ConnectionOptionsApple) (interface{}, error) {
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
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsLinkedin(options *management.ConnectionOptionsLinkedin) (interface{}, error) {
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
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsSalesforce(options *management.ConnectionOptionsSalesforce) (interface{}, error) {
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
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsSMS(options *management.ConnectionOptionsSMS) (interface{}, error) {
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
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsOIDC(options *management.ConnectionOptionsOIDC) (interface{}, error) {
	m := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"icon_url":                 options.GetLogoURL(),
		"tenant_domain":            options.GetTenantDomain(),
		"domain_aliases":           options.DomainAliases,
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
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsEmail(options *management.ConnectionOptionsEmail) (interface{}, error) {
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
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsAD(options *management.ConnectionOptionsAD) (interface{}, error) {
	m := map[string]interface{}{
		"tenant_domain":            options.GetTenantDomain(),
		"domain_aliases":           options.DomainAliases,
		"icon_url":                 options.GetLogoURL(),
		"ips":                      options.IPs,
		"use_cert_auth":            options.GetCertAuth(),
		"use_kerberos":             options.GetKerberos(),
		"disable_cache":            options.GetDisableCache(),
		"brute_force_protection":   options.GetBruteForceProtection(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsAzureAD(options *management.ConnectionOptionsAzureAD) (interface{}, error) {
	m := map[string]interface{}{
		"client_id":                              options.GetClientID(),
		"client_secret":                          options.GetClientSecret(),
		"app_id":                                 options.GetAppID(),
		"tenant_domain":                          options.GetTenantDomain(),
		"domain":                                 options.GetDomain(),
		"domain_aliases":                         options.DomainAliases,
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
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsADFS(options *management.ConnectionOptionsADFS) (interface{}, error) {
	m := map[string]interface{}{
		"tenant_domain":            options.GetTenantDomain(),
		"domain_aliases":           options.DomainAliases,
		"icon_url":                 options.GetLogoURL(),
		"adfs_server":              options.GetADFSServer(),
		"api_enable_users":         options.GetEnableUsersAPI(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func flattenConnectionOptionsSAML(d ResourceData, options *management.ConnectionOptionsSAML) (interface{}, error) {
	m := map[string]interface{}{
		"signing_cert":             options.GetSigningCert(),
		"protocol_binding":         options.GetProtocolBinding(),
		"debug":                    options.GetDebug(),
		"tenant_domain":            options.GetTenantDomain(),
		"domain_aliases":           options.DomainAliases,
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
		"metadata_xml":             String(d, "options.0.metadata_xml"), // Does not get read back.
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
		return nil, err
	}
	m["fields_map"] = fieldsMap

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, err
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func expandConnection(d ResourceData) (*management.Connection, error) {
	connection := &management.Connection{
		Name:               String(d, "name", IsNewResource()),
		DisplayName:        String(d, "display_name"),
		Strategy:           String(d, "strategy", IsNewResource()),
		IsDomainConnection: Bool(d, "is_domain_connection"),
		EnabledClients:     Set(d, "enabled_clients").List(),
		Realms:             Slice(d, "realms", IsNewResource(), HasChange()),
	}

	if metadataKeyMap := Map(d, "metadata"); metadataKeyMap != nil {
		connection.Metadata = map[string]string{}
		for key, value := range metadataKeyMap {
			connection.Metadata[key] = value.(string)
		}
	}

	strategy := d.Get("strategy").(string)
	switch strategy {
	case management.ConnectionStrategyGoogleApps,
		management.ConnectionStrategyOIDC,
		management.ConnectionStrategyAD,
		management.ConnectionStrategyAzureAD,
		management.ConnectionStrategySAML,
		management.ConnectionStrategyADFS:
		connection.ShowAsButton = Bool(d, "show_as_button")
	}

	var err error
	List(d, "options").Elem(func(d ResourceData) {
		switch strategy {
		case management.ConnectionStrategyAuth0:
			connection.Options, err = expandConnectionOptionsAuth0(d)
		case management.ConnectionStrategyGoogleOAuth2:
			connection.Options, err = expandConnectionOptionsGoogleOAuth2(d)
		case management.ConnectionStrategyGoogleApps:
			connection.Options, err = expandConnectionOptionsGoogleApps(d)
		case management.ConnectionStrategyOAuth2:
			connection.Options, err = expandConnectionOptionsOAuth2(d)
		case management.ConnectionStrategyFacebook:
			connection.Options, err = expandConnectionOptionsFacebook(d)
		case management.ConnectionStrategyApple:
			connection.Options, err = expandConnectionOptionsApple(d)
		case management.ConnectionStrategyLinkedin:
			connection.Options, err = expandConnectionOptionsLinkedin(d)
		case management.ConnectionStrategyGitHub:
			connection.Options, err = expandConnectionOptionsGitHub(d)
		case management.ConnectionStrategyWindowsLive:
			connection.Options, err = expandConnectionOptionsWindowsLive(d)
		case management.ConnectionStrategySalesforce,
			management.ConnectionStrategySalesforceCommunity,
			management.ConnectionStrategySalesforceSandbox:
			connection.Options, err = expandConnectionOptionsSalesforce(d)
		case management.ConnectionStrategySMS:
			connection.Options, err = expandConnectionOptionsSMS(d)
		case management.ConnectionStrategyOIDC:
			connection.Options, err = expandConnectionOptionsOIDC(d)
		case management.ConnectionStrategyAD:
			connection.Options, err = expandConnectionOptionsAD(d)
		case management.ConnectionStrategyAzureAD:
			connection.Options, err = expandConnectionOptionsAzureAD(d)
		case management.ConnectionStrategyEmail:
			connection.Options, err = expandConnectionOptionsEmail(d)
		case management.ConnectionStrategySAML:
			connection.Options, err = expandConnectionOptionsSAML(d)
		case management.ConnectionStrategyADFS:
			connection.Options, err = expandConnectionOptionsADFS(d)
		default:
			log.Printf("[WARN]: Unsupported connection strategy %s", strategy)
			log.Printf("[WARN]: Raise an issue with the auth0 provider in order to support it:")
			log.Printf("[WARN]: 	https://github.com/auth0/terraform-provider-auth0/issues/new")
		}
	})
	if err != nil {
		return nil, err
	}

	return connection, nil
}

func expandConnectionOptionsGitHub(d ResourceData) (*management.ConnectionOptionsGitHub, error) {
	options := &management.ConnectionOptionsGitHub{
		ClientID:           String(d, "client_id"),
		ClientSecret:       String(d, "client_secret"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsAuth0(d ResourceData) (*management.ConnectionOptions, error) {
	options := &management.ConnectionOptions{
		PasswordPolicy:     String(d, "password_policy"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	List(d, "validation").Elem(func(d ResourceData) {
		options.Validation = make(map[string]interface{})
		List(d, "username").Elem(func(d ResourceData) {
			usernameValidation := make(map[string]*int)
			usernameValidation["min"] = Int(d, "min")
			usernameValidation["max"] = Int(d, "max")
			options.Validation["username"] = usernameValidation
		})
	})

	List(d, "password_history").Elem(func(d ResourceData) {
		options.PasswordHistory = make(map[string]interface{})
		options.PasswordHistory["enable"] = Bool(d, "enable")

		if size, ok := d.GetOk("size"); ok {
			options.PasswordHistory["size"] = auth0.Int(size.(int))
		}
	})

	List(d, "password_no_personal_info").Elem(func(d ResourceData) {
		options.PasswordNoPersonalInfo = make(map[string]interface{})
		options.PasswordNoPersonalInfo["enable"] = Bool(d, "enable")
	})

	List(d, "password_dictionary").Elem(func(d ResourceData) {
		options.PasswordDictionary = make(map[string]interface{})
		options.PasswordDictionary["enable"] = Bool(d, "enable")
		options.PasswordDictionary["dictionary"] = Set(d, "dictionary").List()
	})

	List(d, "password_complexity_options").Elem(func(d ResourceData) {
		options.PasswordComplexityOptions = make(map[string]interface{})
		options.PasswordComplexityOptions["min_length"] = Int(d, "min_length")
	})

	List(d, "mfa").Elem(func(d ResourceData) {
		options.MFA = make(map[string]interface{})
		options.MFA["active"] = Bool(d, "active")
		options.MFA["return_enroll_settings"] = Bool(d, "return_enroll_settings")
	})

	options.EnabledDatabaseCustomization = Bool(d, "enabled_database_customization")
	options.BruteForceProtection = Bool(d, "brute_force_protection")
	options.ImportMode = Bool(d, "import_mode")
	options.DisableSignup = Bool(d, "disable_signup")
	options.RequiresUsername = Bool(d, "requires_username")
	options.CustomScripts = Map(d, "custom_scripts")
	options.Configuration = Map(d, "configuration")

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsGoogleOAuth2(d ResourceData) (*management.ConnectionOptionsGoogleOAuth2, error) {
	options := &management.ConnectionOptionsGoogleOAuth2{
		ClientID:           String(d, "client_id"),
		ClientSecret:       String(d, "client_secret"),
		AllowedAudiences:   Set(d, "allowed_audiences").List(),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsGoogleApps(d ResourceData) (*management.ConnectionOptionsGoogleApps, error) {
	options := &management.ConnectionOptionsGoogleApps{
		ClientID:           String(d, "client_id"),
		ClientSecret:       String(d, "client_secret"),
		Domain:             String(d, "domain"),
		TenantDomain:       String(d, "tenant_domain"),
		EnableUsersAPI:     Bool(d, "api_enable_users"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
		DomainAliases:      Set(d, "domain_aliases").List(),
		LogoURL:            String(d, "icon_url"),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsOAuth2(d ResourceData) (*management.ConnectionOptionsOAuth2, error) {
	options := &management.ConnectionOptionsOAuth2{
		ClientID:           String(d, "client_id"),
		ClientSecret:       String(d, "client_secret"),
		AuthorizationURL:   String(d, "authorization_endpoint"),
		TokenURL:           String(d, "token_endpoint"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
		LogoURL:            String(d, "icon_url"),
		PKCEEnabled:        Bool(d, "pkce_enabled"),
	}
	options.Scripts = Map(d, "scripts")

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsFacebook(d ResourceData) (*management.ConnectionOptionsFacebook, error) {
	options := &management.ConnectionOptionsFacebook{
		ClientID:           String(d, "client_id"),
		ClientSecret:       String(d, "client_secret"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsApple(d ResourceData) (*management.ConnectionOptionsApple, error) {
	options := &management.ConnectionOptionsApple{
		ClientID:           String(d, "client_id"),
		ClientSecret:       String(d, "client_secret"),
		TeamID:             String(d, "team_id"),
		KeyID:              String(d, "key_id"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsLinkedin(d ResourceData) (*management.ConnectionOptionsLinkedin, error) {
	options := &management.ConnectionOptionsLinkedin{
		ClientID:           String(d, "client_id"),
		ClientSecret:       String(d, "client_secret"),
		StrategyVersion:    Int(d, "strategy_version"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsSalesforce(d ResourceData) (*management.ConnectionOptionsSalesforce, error) {
	options := &management.ConnectionOptionsSalesforce{
		ClientID:           String(d, "client_id"),
		ClientSecret:       String(d, "client_secret"),
		CommunityBaseURL:   String(d, "community_base_url"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsWindowsLive(d ResourceData) (*management.ConnectionOptionsWindowsLive, error) {
	options := &management.ConnectionOptionsWindowsLive{
		ClientID:           String(d, "client_id"),
		ClientSecret:       String(d, "client_secret"),
		StrategyVersion:    Int(d, "strategy_version"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsSMS(d ResourceData) (*management.ConnectionOptionsSMS, error) {
	options := &management.ConnectionOptionsSMS{
		Name:                 String(d, "name"),
		From:                 String(d, "from"),
		Syntax:               String(d, "syntax"),
		Template:             String(d, "template"),
		TwilioSID:            String(d, "twilio_sid"),
		TwilioToken:          String(d, "twilio_token"),
		MessagingServiceSID:  String(d, "messaging_service_sid"),
		Provider:             String(d, "provider"),
		GatewayURL:           String(d, "gateway_url"),
		ForwardRequestInfo:   Bool(d, "forward_request_info"),
		DisableSignup:        Bool(d, "disable_signup"),
		BruteForceProtection: Bool(d, "brute_force_protection"),
	}

	List(d, "totp").Elem(func(d ResourceData) {
		options.OTP = &management.ConnectionOptionsOTP{
			TimeStep: Int(d, "time_step"),
			Length:   Int(d, "length"),
		}
	})

	List(d, "gateway_authentication").Elem(func(d ResourceData) {
		options.GatewayAuthentication = &management.ConnectionGatewayAuthentication{
			Method:              String(d, "method"),
			Subject:             String(d, "subject"),
			Audience:            String(d, "audience"),
			Secret:              String(d, "secret"),
			SecretBase64Encoded: Bool(d, "secret_base64_encoded"),
		}
	})

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsEmail(d ResourceData) (*management.ConnectionOptionsEmail, error) {
	options := &management.ConnectionOptionsEmail{
		Name:          String(d, "name"),
		DisableSignup: Bool(d, "disable_signup"),
		Email: &management.ConnectionOptionsEmailSettings{
			Syntax:  String(d, "syntax"),
			From:    String(d, "from"),
			Subject: String(d, "subject"),
			Body:    String(d, "template"),
		},
		BruteForceProtection: Bool(d, "brute_force_protection"),
		SetUserAttributes:    String(d, "set_user_root_attributes"),
		NonPersistentAttrs:   castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	List(d, "totp").Elem(func(d ResourceData) {
		options.OTP = &management.ConnectionOptionsOTP{
			TimeStep: Int(d, "time_step"),
			Length:   Int(d, "length"),
		}
	})

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsAD(d ResourceData) (*management.ConnectionOptionsAD, error) {
	options := &management.ConnectionOptionsAD{
		DomainAliases:        Set(d, "domain_aliases").List(),
		TenantDomain:         String(d, "tenant_domain"),
		LogoURL:              String(d, "icon_url"),
		IPs:                  Set(d, "ips").List(),
		CertAuth:             Bool(d, "use_cert_auth"),
		Kerberos:             Bool(d, "use_kerberos"),
		DisableCache:         Bool(d, "disable_cache"),
		SetUserAttributes:    String(d, "set_user_root_attributes"),
		NonPersistentAttrs:   castToListOfStrings(Set(d, "non_persistent_attrs").List()),
		BruteForceProtection: Bool(d, "brute_force_protection"),
	}

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsAzureAD(d ResourceData) (*management.ConnectionOptionsAzureAD, error) {
	options := &management.ConnectionOptionsAzureAD{
		ClientID:            String(d, "client_id"),
		ClientSecret:        String(d, "client_secret"),
		AppID:               String(d, "app_id"),
		Domain:              String(d, "domain"),
		DomainAliases:       Set(d, "domain_aliases").List(),
		TenantDomain:        String(d, "tenant_domain"),
		MaxGroupsToRetrieve: String(d, "max_groups_to_retrieve"),
		UseWSFederation:     Bool(d, "use_wsfed"),
		WAADProtocol:        String(d, "waad_protocol"),
		UseCommonEndpoint:   Bool(d, "waad_common_endpoint"),
		EnableUsersAPI:      Bool(d, "api_enable_users"),
		LogoURL:             String(d, "icon_url"),
		IdentityAPI:         String(d, "identity_api"),
		SetUserAttributes:   String(d, "set_user_root_attributes"),
		NonPersistentAttrs:  castToListOfStrings(Set(d, "non_persistent_attrs").List()),
		TrustEmailVerified:  String(d, "should_trust_email_verified_connection"),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsOIDC(d ResourceData) (*management.ConnectionOptionsOIDC, error) {
	options := &management.ConnectionOptionsOIDC{
		ClientID:              String(d, "client_id"),
		ClientSecret:          String(d, "client_secret"),
		TenantDomain:          String(d, "tenant_domain"),
		DomainAliases:         Set(d, "domain_aliases").List(),
		LogoURL:               String(d, "icon_url"),
		DiscoveryURL:          String(d, "discovery_url"),
		AuthorizationEndpoint: String(d, "authorization_endpoint"),
		Issuer:                String(d, "issuer"),
		JWKSURI:               String(d, "jwks_uri"),
		Type:                  String(d, "type"),
		UserInfoEndpoint:      String(d, "userinfo_endpoint"),
		TokenEndpoint:         String(d, "token_endpoint"),
		SetUserAttributes:     String(d, "set_user_root_attributes"),
		NonPersistentAttrs:    castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsSAML(d ResourceData) (*management.ConnectionOptionsSAML, error) {
	options := &management.ConnectionOptionsSAML{
		Debug:              Bool(d, "debug"),
		SigningCert:        String(d, "signing_cert"),
		ProtocolBinding:    String(d, "protocol_binding"),
		TenantDomain:       String(d, "tenant_domain"),
		DomainAliases:      Set(d, "domain_aliases").List(),
		SignInEndpoint:     String(d, "sign_in_endpoint"),
		SignOutEndpoint:    String(d, "sign_out_endpoint"),
		DisableSignOut:     Bool(d, "disable_sign_out"),
		SignatureAlgorithm: String(d, "signature_algorithm"),
		DigestAglorithm:    String(d, "digest_algorithm"),
		SignSAMLRequest:    Bool(d, "sign_saml_request"),
		RequestTemplate:    String(d, "request_template"),
		UserIDAttribute:    String(d, "user_id_attribute"),
		LogoURL:            String(d, "icon_url"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
		EntityID:           String(d, "entity_id"),
		MetadataXML:        String(d, "metadata_xml"),
		MetadataURL:        String(d, "metadata_url"),
	}

	List(d, "idp_initiated").Elem(func(d ResourceData) {
		options.IdpInitiated = &management.ConnectionOptionsSAMLIdpInitiated{
			ClientID:             String(d, "client_id"),
			ClientProtocol:       String(d, "client_protocol"),
			ClientAuthorizeQuery: String(d, "client_authorize_query"),
		}
	})

	List(d, "signing_key").Elem(func(d ResourceData) {
		options.SigningKey = &management.ConnectionOptionsSAMLSigningKey{
			Cert: String(d, "cert"),
			Key:  String(d, "key"),
		}
	})

	var err error
	options.FieldsMap, err = JSON(d, "fields_map")
	if err != nil {
		return nil, err
	}

	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

func expandConnectionOptionsADFS(d ResourceData) (*management.ConnectionOptionsADFS, error) {
	options := &management.ConnectionOptionsADFS{
		TenantDomain:       String(d, "tenant_domain"),
		DomainAliases:      Set(d, "domain_aliases").List(),
		LogoURL:            String(d, "icon_url"),
		ADFSServer:         String(d, "adfs_server"),
		EnableUsersAPI:     Bool(d, "api_enable_users"),
		SetUserAttributes:  String(d, "set_user_root_attributes"),
		NonPersistentAttrs: castToListOfStrings(Set(d, "non_persistent_attrs").List()),
	}

	var err error
	options.UpstreamParams, err = JSON(d, "upstream_params")
	if err != nil {
		return nil, err
	}

	return options, nil
}

type scoper interface {
	Scopes() []string
	SetScopes(enable bool, scopes ...string)
}

func expandConnectionOptionsScopes(d ResourceData, s scoper) {
	scopesList := Set(d, "scopes").List()
	_, scopesDiff := Diff(d, "scopes")
	for _, scope := range scopesList {
		s.SetScopes(true, scope.(string))
	}
	for _, scope := range scopesDiff.List() {
		s.SetScopes(false, scope.(string))
	}
}

func castToListOfStrings(interfaces []interface{}) *[]string {
	var strings []string
	for _, v := range interfaces {
		strings = append(strings, v.(string))
	}
	return &strings
}
