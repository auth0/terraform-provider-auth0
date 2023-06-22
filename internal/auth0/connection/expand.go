package connection

import (
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandConnection(d *schema.ResourceData, api *management.Management) (*management.Connection, diag.Diagnostics) {
	config := d.GetRawConfig()

	connection := &management.Connection{
		DisplayName:        value.String(config.GetAttr("display_name")),
		IsDomainConnection: value.Bool(config.GetAttr("is_domain_connection")),
		Metadata:           value.MapOfStrings(config.GetAttr("metadata")),
	}

	if d.IsNewResource() {
		connection.Name = value.String(config.GetAttr("name"))
		connection.Strategy = value.String(config.GetAttr("strategy"))
	}

	if d.HasChange("realms") {
		connection.Realms = value.Strings(config.GetAttr("realms"))
	}

	if d.HasChange("enabled_clients") {
		connection.EnabledClients = value.Strings(config.GetAttr("enabled_clients"))
	}

	var diagnostics diag.Diagnostics
	strategy := d.Get("strategy").(string)
	showAsButton := value.Bool(config.GetAttr("show_as_button"))

	config.GetAttr("options").ForEachElement(func(_ cty.Value, options cty.Value) (stop bool) {
		switch strategy {
		case management.ConnectionStrategyAuth0:
			connection.Options, diagnostics = expandConnectionOptionsAuth0(d, options, api)
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
		case management.ConnectionStrategyOkta:
			connection.ShowAsButton = showAsButton
			connection.Options, diagnostics = expandConnectionOptionsOkta(d, options)
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
		case management.ConnectionStrategyPingFederate:
			connection.ShowAsButton = showAsButton
			connection.Options, diagnostics = expandConnectionOptionsPingFederate(options)
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

func expandConnectionOptionsAuth0(
	d *schema.ResourceData,
	config cty.Value,
	api *management.Management,
) (*management.ConnectionOptions, diag.Diagnostics) {
	options := &management.ConnectionOptions{
		PasswordPolicy:                   value.String(config.GetAttr("password_policy")),
		NonPersistentAttrs:               value.Strings(config.GetAttr("non_persistent_attrs")),
		SetUserAttributes:                value.String(config.GetAttr("set_user_root_attributes")),
		EnableScriptContext:              value.Bool(config.GetAttr("enable_script_context")),
		EnabledDatabaseCustomization:     value.Bool(config.GetAttr("enabled_database_customization")),
		BruteForceProtection:             value.Bool(config.GetAttr("brute_force_protection")),
		ImportMode:                       value.Bool(config.GetAttr("import_mode")),
		DisableSignup:                    value.Bool(config.GetAttr("disable_signup")),
		DisableSelfServiceChangePassword: value.Bool(config.GetAttr("disable_self_service_change_password")),
		RequiresUsername:                 value.Bool(config.GetAttr("requires_username")),
		CustomScripts:                    value.MapOfStrings(config.GetAttr("custom_scripts")),
		Configuration:                    value.MapOfStrings(config.GetAttr("configuration")),
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
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if !d.IsNewResource() {
		apiConn, err := api.Connection.Read(d.Id())
		if err != nil {
			return nil, diag.FromErr(err)
		}

		diags := checkForUnmanagedConfigurationSecrets(
			options.GetConfiguration(),
			apiConn.Options.(*management.ConnectionOptions).GetConfiguration(),
		)

		if diags.HasError() {
			return nil, diags
		}
	}

	return options, nil
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
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
		DomainAliases:      value.Strings(config.GetAttr("domain_aliases")),
		LogoURL:            value.String(config.GetAttr("icon_url")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
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
		NonPersistentAttrs:   value.Strings(config.GetAttr("non_persistent_attrs")),
		BruteForceProtection: value.Bool(config.GetAttr("brute_force_protection")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
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
		NonPersistentAttrs:  value.Strings(config.GetAttr("non_persistent_attrs")),
		TrustEmailVerified:  value.String(config.GetAttr("should_trust_email_verified_connection")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
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

func expandConnectionOptionsOkta(
	d *schema.ResourceData,
	config cty.Value,
) (*management.ConnectionOptionsOkta, diag.Diagnostics) {
	options := &management.ConnectionOptionsOkta{
		ClientID:              value.String(config.GetAttr("client_id")),
		ClientSecret:          value.String(config.GetAttr("client_secret")),
		Domain:                value.String(config.GetAttr("domain")),
		DomainAliases:         value.Strings(config.GetAttr("domain_aliases")),
		AuthorizationEndpoint: value.String(config.GetAttr("authorization_endpoint")),
		Issuer:                value.String(config.GetAttr("issuer")),
		JWKSURI:               value.String(config.GetAttr("jwks_uri")),
		UserInfoEndpoint:      value.String(config.GetAttr("userinfo_endpoint")),
		TokenEndpoint:         value.String(config.GetAttr("token_endpoint")),
		NonPersistentAttrs:    value.Strings(config.GetAttr("non_persistent_attrs")),
		SetUserAttributes:     value.String(config.GetAttr("set_user_root_attributes")),
		LogoURL:               value.String(config.GetAttr("icon_url")),
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
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
		EntityID:           value.String(config.GetAttr("entity_id")),
		MetadataXML:        value.String(config.GetAttr("metadata_xml")),
		MetadataURL:        value.String(config.GetAttr("metadata_url")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
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
		FedMetadataXML:     value.String(config.GetAttr("fed_metadata_xml")),
		SignInEndpoint:     value.String(config.GetAttr("sign_in_endpoint")),
		EnableUsersAPI:     value.Bool(config.GetAttr("api_enable_users")),
		TrustEmailVerified: value.String(config.GetAttr("should_trust_email_verified_connection")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
	}

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsPingFederate(
	config cty.Value,
) (*management.ConnectionOptionsPingFederate, diag.Diagnostics) {
	options := &management.ConnectionOptionsPingFederate{
		SigningCert:         value.String(config.GetAttr("signing_cert")),
		LogoURL:             value.String(config.GetAttr("icon_url")),
		TenantDomain:        value.String(config.GetAttr("tenant_domain")),
		DomainAliases:       value.Strings(config.GetAttr("domain_aliases")),
		SignInEndpoint:      value.String(config.GetAttr("sign_in_endpoint")),
		DigestAlgorithm:     value.String(config.GetAttr("digest_algorithm")),
		SignSAMLRequest:     value.Bool(config.GetAttr("sign_saml_request")),
		SignatureAlgorithm:  value.String(config.GetAttr("signature_algorithm")),
		PingFederateBaseURL: value.String(config.GetAttr("ping_federate_base_url")),
		NonPersistentAttrs:  value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
	}

	config.GetAttr("idp_initiated").ForEachElement(func(_ cty.Value, idp cty.Value) (stop bool) {
		options.IdpInitiated = &management.ConnectionOptionsSAMLIdpInitiated{
			ClientID:             value.String(idp.GetAttr("client_id")),
			ClientProtocol:       value.String(idp.GetAttr("client_protocol")),
			ClientAuthorizeQuery: value.String(idp.GetAttr("client_authorize_query")),
		}

		return stop
	})

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

type scoper interface {
	Scopes() []string
	SetScopes(enable bool, scopes ...string)
}

func expandConnectionOptionsScopes(d *schema.ResourceData, s scoper) {
	scopesList := d.Get("options.0.scopes").(*schema.Set).List()

	_, scopesToDisable := value.Difference(d, "options.0.scopes")

	for _, scope := range scopesToDisable {
		s.SetScopes(false, scope.(string))
	}

	for _, scope := range scopesList {
		s.SetScopes(true, scope.(string))
	}
}
