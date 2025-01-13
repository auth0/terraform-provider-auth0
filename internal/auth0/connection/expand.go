package connection

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

var expandConnectionOptionsMap = map[string]expandConnectionOptionsFunc{
	// Database Connection.
	management.ConnectionStrategyAuth0: expandConnectionOptionsAuth0,

	// Social Connections.
	management.ConnectionStrategyGoogleOAuth2:        expandConnectionOptionsGoogleOAuth2,
	management.ConnectionStrategyOAuth2:              expandConnectionOptionsOAuth2,
	management.ConnectionStrategyDropbox:             expandConnectionOptionsOAuth2,
	management.ConnectionStrategyBitBucket:           expandConnectionOptionsOAuth2,
	management.ConnectionStrategyPaypal:              expandConnectionOptionsOAuth2,
	management.ConnectionStrategyTwitter:             expandConnectionOptionsOAuth2,
	management.ConnectionStrategyAmazon:              expandConnectionOptionsOAuth2,
	management.ConnectionStrategyYahoo:               expandConnectionOptionsOAuth2,
	management.ConnectionStrategyBox:                 expandConnectionOptionsOAuth2,
	management.ConnectionStrategyWordpress:           expandConnectionOptionsOAuth2,
	management.ConnectionStrategyShopify:             expandConnectionOptionsOAuth2,
	management.ConnectionStrategyLine:                expandConnectionOptionsOAuth2,
	management.ConnectionStrategyCustom:              expandConnectionOptionsOAuth2,
	management.ConnectionStrategyFacebook:            expandConnectionOptionsFacebook,
	management.ConnectionStrategyApple:               expandConnectionOptionsApple,
	management.ConnectionStrategyLinkedin:            expandConnectionOptionsLinkedin,
	management.ConnectionStrategyGitHub:              expandConnectionOptionsGitHub,
	management.ConnectionStrategyWindowsLive:         expandConnectionOptionsWindowsLive,
	management.ConnectionStrategySalesforce:          expandConnectionOptionsSalesforce,
	management.ConnectionStrategySalesforceCommunity: expandConnectionOptionsSalesforce,
	management.ConnectionStrategySalesforceSandbox:   expandConnectionOptionsSalesforce,

	// Passwordless Connections.
	management.ConnectionStrategySMS:   expandConnectionOptionsSMS,
	management.ConnectionStrategyEmail: expandConnectionOptionsEmail,

	// Enterprise Connections.
	management.ConnectionStrategyOIDC:         expandConnectionOptionsOIDC,
	management.ConnectionStrategyGoogleApps:   expandConnectionOptionsGoogleApps,
	management.ConnectionStrategyOkta:         expandConnectionOptionsOkta,
	management.ConnectionStrategyAD:           expandConnectionOptionsAD,
	management.ConnectionStrategyAzureAD:      expandConnectionOptionsAzureAD,
	management.ConnectionStrategySAML:         expandConnectionOptionsSAML,
	management.ConnectionStrategyADFS:         expandConnectionOptionsADFS,
	management.ConnectionStrategyPingFederate: expandConnectionOptionsPingFederate,
}

type expandConnectionOptionsFunc func(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics)

type scoper interface {
	Scopes() []string
	SetScopes(enable bool, scopes ...string)
}

func expandConnection(
	ctx context.Context,
	data *schema.ResourceData,
	api *management.Management,
) (*management.Connection, diag.Diagnostics) {
	config := data.GetRawConfig()

	connection := &management.Connection{
		DisplayName:        value.String(config.GetAttr("display_name")),
		IsDomainConnection: value.Bool(config.GetAttr("is_domain_connection")),
		Metadata:           value.MapOfStrings(config.GetAttr("metadata")),
	}

	strategy := data.Get("strategy").(string)

	if data.IsNewResource() {
		connection.Name = value.String(config.GetAttr("name"))
		connection.Strategy = &strategy
	}

	if data.HasChange("realms") {
		connection.Realms = value.Strings(config.GetAttr("realms"))
	}

	var diagnostics diag.Diagnostics
	connection.Options, diagnostics = expandConnectionOptions(data, strategy)

	if connectionIsEnterprise(strategy) {
		connection.ShowAsButton = value.Bool(config.GetAttr("show_as_button"))

		if !data.IsNewResource() && connection.Options != nil {
			err := passThroughUnconfigurableConnectionOptions(ctx, api, data.Id(), strategy, connection)
			if err != nil {
				return nil, diag.FromErr(err)
			}
		}
	}

	// Prevent erasing database configuration secrets.
	if !data.IsNewResource() && strategy == management.ConnectionStrategyAuth0 && connection.Options != nil {
		apiConn, err := api.Connection.Read(ctx, data.Id())
		if err != nil {
			return nil, diag.FromErr(err)
		}

		diagnostics = append(
			diagnostics,
			checkForUnmanagedConfigurationSecrets(
				connection.Options.(*management.ConnectionOptions).GetConfiguration(),
				apiConn.Options.(*management.ConnectionOptions).GetConfiguration(),
			)...,
		)
	}

	return connection, diagnostics
}

func connectionIsEnterprise(strategy string) bool {
	switch strategy {
	case management.ConnectionStrategyGoogleApps,
		management.ConnectionStrategyOIDC,
		management.ConnectionStrategyOkta,
		management.ConnectionStrategyAD,
		management.ConnectionStrategyAzureAD,
		management.ConnectionStrategySAML,
		management.ConnectionStrategyADFS,
		management.ConnectionStrategyPingFederate:
		return true
	default:
		return false
	}
}

func expandConnectionOptions(data *schema.ResourceData, strategy string) (interface{}, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	var connectionOptions interface{}

	data.GetRawConfig().GetAttr("options").ForEachElement(func(_ cty.Value, optionsConfig cty.Value) (stop bool) {
		connectionOptionsFunc, ok := expandConnectionOptionsMap[strategy]
		if !ok {
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

			return true
		}

		connectionOptions, diagnostics = connectionOptionsFunc(data, optionsConfig)

		return true
	})

	return connectionOptions, diagnostics
}

func expandConnectionOptionsGitHub(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
	options := &management.ConnectionOptionsGitHub{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsAttributes(config cty.Value) *management.ConnectionOptionsAttributes {
	var coa *management.ConnectionOptionsAttributes
	config.ForEachElement(
		func(_ cty.Value, attributes cty.Value) (stop bool) {
			coa = &management.ConnectionOptionsAttributes{
				Email:       expandConnectionOptionsEmailAttribute(attributes),
				Username:    expandConnectionOptionsUsernameAttribute(attributes),
				PhoneNumber: expandConnectionOptionsPhoneNumberAttribute(attributes),
			}
			return stop
		})
	return coa
}

func expandConnectionOptionsEmailAttribute(config cty.Value) *management.ConnectionOptionsEmailAttribute {
	var coea *management.ConnectionOptionsEmailAttribute
	config.GetAttr("email").ForEachElement(
		func(_ cty.Value, email cty.Value) (stop bool) {
			coea = &management.ConnectionOptionsEmailAttribute{
				Identifier:         expandConnectionOptionsAttributeIdentifier(email),
				ProfileRequired:    value.Bool(email.GetAttr("profile_required")),
				VerificationMethod: (*management.ConnectionOptionsEmailAttributeVerificationMethod)(value.String(email.GetAttr("verification_method"))),
				Signup:             expandConnectionOptionsAttributeSignup(email),
			}
			return stop
		})
	return coea
}

func expandConnectionOptionsUsernameAttribute(config cty.Value) *management.ConnectionOptionsUsernameAttribute {
	var coua *management.ConnectionOptionsUsernameAttribute
	config.GetAttr("username").ForEachElement(
		func(_ cty.Value, username cty.Value) (stop bool) {
			coua = &management.ConnectionOptionsUsernameAttribute{
				Identifier:      expandConnectionOptionsAttributeIdentifier(username),
				ProfileRequired: value.Bool(username.GetAttr("profile_required")),
				Signup:          expandConnectionOptionsAttributeUsernameSignup(username),
				Validation:      expandConnectionOptionsAttributeValidation(username),
			}
			return stop
		})
	return coua
}

func expandConnectionOptionsPhoneNumberAttribute(config cty.Value) *management.ConnectionOptionsPhoneNumberAttribute {
	var copa *management.ConnectionOptionsPhoneNumberAttribute
	config.GetAttr("phone_number").ForEachElement(
		func(_ cty.Value, phoneNumber cty.Value) (stop bool) {
			copa = &management.ConnectionOptionsPhoneNumberAttribute{
				Identifier:      expandConnectionOptionsAttributeIdentifier(phoneNumber),
				ProfileRequired: value.Bool(phoneNumber.GetAttr("profile_required")),
				Signup:          expandConnectionOptionsAttributeSignup(phoneNumber),
			}
			return stop
		})
	return copa
}

func expandConnectionOptionsAttributeIdentifier(config cty.Value) *management.ConnectionOptionsAttributeIdentifier {
	var coai *management.ConnectionOptionsAttributeIdentifier
	config.GetAttr("identifier").ForEachElement(
		func(_ cty.Value, identifier cty.Value) (stop bool) {
			coai = &management.ConnectionOptionsAttributeIdentifier{
				Active: value.Bool(identifier.GetAttr("active")),
			}
			return stop
		})
	return coai
}

func expandConnectionOptionsAttributeUsernameSignup(config cty.Value) *management.ConnectionOptionsAttributeSignup {
	var coas *management.ConnectionOptionsAttributeSignup
	config.GetAttr("signup").ForEachElement(
		func(_ cty.Value, signup cty.Value) (stop bool) {
			coas = &management.ConnectionOptionsAttributeSignup{
				Status: value.String(signup.GetAttr("status")),
			}
			return stop
		})
	return coas
}

func expandConnectionOptionsAttributeSignup(config cty.Value) *management.ConnectionOptionsAttributeSignup {
	var coas *management.ConnectionOptionsAttributeSignup
	config.GetAttr("signup").ForEachElement(
		func(_ cty.Value, signup cty.Value) (stop bool) {
			coas = &management.ConnectionOptionsAttributeSignup{
				Status:       value.String(signup.GetAttr("status")),
				Verification: expandConnectionOptionsAttributeVerification(signup),
			}
			return stop
		})
	return coas
}

func expandConnectionOptionsAttributeVerification(config cty.Value) *management.ConnectionOptionsAttributeVerification {
	var coav *management.ConnectionOptionsAttributeVerification
	config.GetAttr("verification").ForEachElement(
		func(_ cty.Value, verification cty.Value) (stop bool) {
			coav = &management.ConnectionOptionsAttributeVerification{
				Active: value.Bool(verification.GetAttr("active")),
			}
			return stop
		})
	return coav
}

func expandConnectionOptionsAttributeValidation(config cty.Value) *management.ConnectionOptionsAttributeValidation {
	var coav *management.ConnectionOptionsAttributeValidation
	config.GetAttr("validation").ForEachElement(
		func(_ cty.Value, validation cty.Value) (stop bool) {
			coav = &management.ConnectionOptionsAttributeValidation{
				MinLength:    value.Int(validation.GetAttr("min_length")),
				MaxLength:    value.Int(validation.GetAttr("max_length")),
				AllowedTypes: expandConnectionOptionsAttributeAllowedTypes(validation),
			}
			return stop
		})
	return coav
}

func expandConnectionOptionsAttributeAllowedTypes(config cty.Value) *management.ConnectionOptionsAttributeAllowedTypes {
	var coaat *management.ConnectionOptionsAttributeAllowedTypes
	config.GetAttr("allowed_types").ForEachElement(
		func(_ cty.Value, allowedTypes cty.Value) (stop bool) {
			coaat = &management.ConnectionOptionsAttributeAllowedTypes{
				Email:       value.Bool(allowedTypes.GetAttr("email")),
				PhoneNumber: value.Bool(allowedTypes.GetAttr("phone_number")),
			}
			return stop
		})
	return coaat
}

func expandConnectionOptionsAuth0(_ *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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
		Precedence:                       value.Strings(config.GetAttr("precedence")),
		Attributes:                       expandConnectionOptionsAttributes(config.GetAttr("attributes")),
		StrategyVersion:                  value.Int(config.GetAttr("strategy_version")),
	}

	config.GetAttr("validation").ForEachElement(
		func(_ cty.Value, validation cty.Value) (stop bool) {
			validationOption := make(map[string]interface{})

			validation.GetAttr("username").ForEachElement(
				func(_ cty.Value, username cty.Value) (stop bool) {
					usernameValidation := make(map[string]*int)

					if usernameMinLength := value.Int(username.GetAttr("min")); usernameMinLength != nil {
						usernameValidation["min"] = usernameMinLength
					}
					if usernameMaxLength := value.Int(username.GetAttr("max")); usernameMaxLength != nil {
						usernameValidation["max"] = usernameMaxLength
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

func expandConnectionOptionsGoogleOAuth2(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
	options := &management.ConnectionOptionsGoogleOAuth2{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		AllowedAudiences:   value.Strings(config.GetAttr("allowed_audiences")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsGoogleApps(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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

	if data.IsNewResource() {
		options.MapUserIDtoID = value.Bool(config.GetAttr("map_user_id_to_id"))
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsOAuth2(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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
		StrategyVersion:    value.Int(config.GetAttr("strategy_version")),
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsFacebook(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
	options := &management.ConnectionOptionsFacebook{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsApple(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
	options := &management.ConnectionOptionsApple{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		TeamID:             value.String(config.GetAttr("team_id")),
		KeyID:              value.String(config.GetAttr("key_id")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsLinkedin(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
	options := &management.ConnectionOptionsLinkedin{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		StrategyVersion:    value.Int(config.GetAttr("strategy_version")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsSalesforce(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
	options := &management.ConnectionOptionsSalesforce{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		CommunityBaseURL:   value.String(config.GetAttr("community_base_url")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsWindowsLive(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
	options := &management.ConnectionOptionsWindowsLive{
		ClientID:           value.String(config.GetAttr("client_id")),
		ClientSecret:       value.String(config.GetAttr("client_secret")),
		StrategyVersion:    value.Int(config.GetAttr("strategy_version")),
		SetUserAttributes:  value.String(config.GetAttr("set_user_root_attributes")),
		NonPersistentAttrs: value.Strings(config.GetAttr("non_persistent_attrs")),
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsSMS(_ *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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

func expandConnectionOptionsEmail(_ *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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

func expandConnectionOptionsAD(_ *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
	options := &management.ConnectionOptionsAD{
		DomainAliases:                    value.Strings(config.GetAttr("domain_aliases")),
		TenantDomain:                     value.String(config.GetAttr("tenant_domain")),
		LogoURL:                          value.String(config.GetAttr("icon_url")),
		IPs:                              value.Strings(config.GetAttr("ips")),
		CertAuth:                         value.Bool(config.GetAttr("use_cert_auth")),
		Kerberos:                         value.Bool(config.GetAttr("use_kerberos")),
		DisableCache:                     value.Bool(config.GetAttr("disable_cache")),
		NonPersistentAttrs:               value.Strings(config.GetAttr("non_persistent_attrs")),
		BruteForceProtection:             value.Bool(config.GetAttr("brute_force_protection")),
		DisableSelfServiceChangePassword: value.Bool(config.GetAttr("disable_self_service_change_password")),
		StrategyVersion:                  value.Int(config.GetAttr("strategy_version")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
	}

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsAzureAD(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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
		StrategyVersion:     value.Int(config.GetAttr("strategy_version")),
		UserIDAttribute:     value.String(config.GetAttr("user_id_attribute")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
	}

	expandConnectionOptionsScopes(data, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsOIDC(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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

	config.GetAttr("connection_settings").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		options.ConnectionSettings = &management.ConnectionOptionsOIDCConnectionSettings{
			PKCE: value.String(config.GetAttr("pkce")),
		}

		return true
	})

	var err error
	config.GetAttr("attribute_map").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		options.AttributeMap = &management.ConnectionOptionsOIDCAttributeMap{
			UserInfoScope: value.String(config.GetAttr("userinfo_scope")),
			MappingMode:   value.String(config.GetAttr("mapping_mode")),
		}

		options.AttributeMap.Attributes, err = value.MapFromJSON(config.GetAttr("attributes"))

		return true
	})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	expandConnectionOptionsScopes(data, options)

	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsOkta(data *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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

	config.GetAttr("connection_settings").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		options.ConnectionSettings = &management.ConnectionOptionsOIDCConnectionSettings{
			PKCE: value.String(config.GetAttr("pkce")),
		}

		return true
	})

	var err error
	config.GetAttr("attribute_map").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		options.AttributeMap = &management.ConnectionOptionsOIDCAttributeMap{
			UserInfoScope: value.String(config.GetAttr("userinfo_scope")),
			MappingMode:   value.String(config.GetAttr("mapping_mode")),
		}

		options.AttributeMap.Attributes, err = value.MapFromJSON(config.GetAttr("attributes"))

		return true
	})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	expandConnectionOptionsScopes(data, options)

	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsSAML(_ *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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
		StrategyVersion:    value.Int(config.GetAttr("strategy_version")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
	}

	config.GetAttr("idp_initiated").ForEachElement(func(_ cty.Value, idp cty.Value) (stop bool) {
		options.IdpInitiated = &management.ConnectionOptionsSAMLIdpInitiated{
			Enabled:              value.Bool(idp.GetAttr("enabled")),
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

	config.GetAttr("decryption_key").ForEachElement(func(_ cty.Value, key cty.Value) (stop bool) {
		options.DecryptionKey = &management.ConnectionOptionsSAMLDecryptionKey{
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

func expandConnectionOptionsADFS(_ *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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
		StrategyVersion:    value.Int(config.GetAttr("strategy_version")),
	}

	options.SetUserAttributes = value.String(config.GetAttr("set_user_root_attributes"))
	if options.GetSetUserAttributes() == "on_each_login" {
		options.SetUserAttributes = nil // This needs to be omitted to have the toggle enabled in the UI.
	}

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))

	return options, diag.FromErr(err)
}

func expandConnectionOptionsPingFederate(_ *schema.ResourceData, config cty.Value) (interface{}, diag.Diagnostics) {
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
			Enabled:              value.Bool(idp.GetAttr("enabled")),
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

func expandConnectionOptionsScopes(data *schema.ResourceData, options scoper) {
	_, scopesToDisable := value.Difference(data, "options.0.scopes")
	for _, scope := range scopesToDisable {
		options.SetScopes(false, scope.(string))
	}

	scopesList := data.Get("options.0.scopes").(*schema.Set).List()
	for _, scope := range scopesList {
		options.SetScopes(true, scope.(string))
	}
}

// passThroughUnconfigurableConnectionOptions ensures that read-only connection options
// set by external services do not get removed from the connection resource.
//
// This is necessary because the "/api/v2/connections/{id}" endpoint does not follow usual
// PATCH behavior, the 'options' property is entirely replaced by the payload object.
func passThroughUnconfigurableConnectionOptions(
	ctx context.Context,
	api *management.Management,
	connectionID string,
	strategy string,
	connection *management.Connection,
) error {
	var err error

	switch strategy {
	case management.ConnectionStrategyAD:
		err = passThroughUnconfigurableConnectionOptionsAD(ctx, api, connectionID, connection)
	case management.ConnectionStrategyAzureAD:
		err = passThroughUnconfigurableConnectionOptionsAzureAD(ctx, api, connectionID, connection)
	case management.ConnectionStrategySAML:
		err = passThroughUnconfigurableConnectionOptionsSAML(ctx, api, connectionID, connection)
	case management.ConnectionStrategyADFS:
		err = passThroughUnconfigurableConnectionOptionsADFS(ctx, api, connectionID, connection)
	case management.ConnectionStrategyPingFederate:
		err = passThroughUnconfigurableConnectionOptionsPingFederate(ctx, api, connectionID, connection)
	case management.ConnectionStrategyGoogleApps:
		err = passThroughUnconfigurableConnectionOptionsGoogleApps(ctx, api, connectionID, connection)
	}

	return err
}

func passThroughUnconfigurableConnectionOptionsAD(
	ctx context.Context,
	api *management.Management,
	connectionID string,
	connection *management.Connection,
) error {
	existingConnection, err := api.Connection.Read(ctx, connectionID)
	if err != nil {
		return err
	}

	if existingConnection.Options == nil {
		return nil
	}

	existingOptions := existingConnection.Options.(*management.ConnectionOptionsAD)

	expandedOptions := connection.Options.(*management.ConnectionOptionsAD)
	expandedOptions.Thumbprints = existingOptions.Thumbprints
	expandedOptions.Certs = existingOptions.Certs
	expandedOptions.AgentIP = existingOptions.AgentIP
	expandedOptions.AgentVersion = existingOptions.AgentVersion
	expandedOptions.AgentMode = existingOptions.AgentMode

	connection.Options = expandedOptions

	return nil
}

func passThroughUnconfigurableConnectionOptionsAzureAD(
	ctx context.Context,
	api *management.Management,
	connectionID string,
	connection *management.Connection,
) error {
	existingConnection, err := api.Connection.Read(ctx, connectionID)
	if err != nil {
		return err
	}

	if existingConnection.Options == nil {
		return nil
	}

	existingOptions := existingConnection.Options.(*management.ConnectionOptionsAzureAD)

	expandedOptions := connection.Options.(*management.ConnectionOptionsAzureAD)
	expandedOptions.Thumbprints = existingOptions.Thumbprints
	expandedOptions.AppDomain = existingOptions.AppDomain
	expandedOptions.CertRolloverNotification = existingOptions.CertRolloverNotification
	expandedOptions.Granted = existingOptions.Granted
	expandedOptions.TenantID = existingOptions.TenantID

	connection.Options = expandedOptions

	return nil
}

func passThroughUnconfigurableConnectionOptionsADFS(
	ctx context.Context,
	api *management.Management,
	connectionID string,
	connection *management.Connection,
) error {
	existingConnection, err := api.Connection.Read(ctx, connectionID)
	if err != nil {
		return err
	}

	if existingConnection.Options == nil {
		return nil
	}

	existingOptions := existingConnection.Options.(*management.ConnectionOptionsADFS)

	expandedOptions := connection.Options.(*management.ConnectionOptionsADFS)
	expandedOptions.Thumbprints = existingOptions.Thumbprints
	expandedOptions.CertRolloverNotification = existingOptions.CertRolloverNotification
	expandedOptions.EntityID = existingOptions.EntityID
	expandedOptions.PreviousThumbprints = existingOptions.PreviousThumbprints

	connection.Options = expandedOptions

	return nil
}

func passThroughUnconfigurableConnectionOptionsSAML(
	ctx context.Context,
	api *management.Management,
	connectionID string,
	connection *management.Connection,
) error {
	existingConnection, err := api.Connection.Read(ctx, connectionID)
	if err != nil {
		return err
	}

	if existingConnection.Options == nil {
		return nil
	}

	existingOptions := existingConnection.Options.(*management.ConnectionOptionsSAML)

	expandedOptions := connection.Options.(*management.ConnectionOptionsSAML)
	expandedOptions.Thumbprints = existingOptions.Thumbprints
	expandedOptions.BindingMethod = existingOptions.BindingMethod
	expandedOptions.CertRolloverNotification = existingOptions.CertRolloverNotification
	expandedOptions.AgentIP = existingOptions.AgentIP
	expandedOptions.AgentVersion = existingOptions.AgentVersion
	expandedOptions.AgentMode = existingOptions.AgentMode
	expandedOptions.ExtGroups = existingOptions.ExtGroups
	expandedOptions.ExtProfile = existingOptions.ExtProfile

	connection.Options = expandedOptions

	return nil
}

func passThroughUnconfigurableConnectionOptionsPingFederate(
	ctx context.Context,
	api *management.Management,
	connectionID string,
	connection *management.Connection,
) error {
	existingConnection, err := api.Connection.Read(ctx, connectionID)
	if err != nil {
		return err
	}

	if existingConnection.Options == nil {
		return nil
	}

	existingOptions := existingConnection.Options.(*management.ConnectionOptionsPingFederate)

	expandedOptions := connection.Options.(*management.ConnectionOptionsPingFederate)
	expandedOptions.APIEnableUsers = existingOptions.APIEnableUsers
	expandedOptions.SignOutEndpoint = existingOptions.SignOutEndpoint
	expandedOptions.Subject = existingOptions.Subject
	expandedOptions.DisableSignout = existingOptions.DisableSignout
	expandedOptions.UserIDAttribute = existingOptions.UserIDAttribute
	expandedOptions.Debug = existingOptions.Debug
	expandedOptions.ProtocolBinding = existingOptions.ProtocolBinding
	expandedOptions.RequestTemplate = existingOptions.RequestTemplate
	expandedOptions.Thumbprints = existingOptions.Thumbprints
	expandedOptions.BindingMethod = existingOptions.BindingMethod
	expandedOptions.Expires = existingOptions.Expires
	expandedOptions.MetadataURL = existingOptions.MetadataURL
	expandedOptions.FieldsMap = existingOptions.FieldsMap
	expandedOptions.MetadataXML = existingOptions.MetadataXML
	expandedOptions.EntityID = existingOptions.EntityID
	expandedOptions.CertRolloverNotification = existingOptions.CertRolloverNotification
	expandedOptions.SigningKey = existingOptions.SigningKey
	expandedOptions.DecryptionKey = existingOptions.DecryptionKey
	expandedOptions.AgentIP = existingOptions.AgentIP
	expandedOptions.AgentVersion = existingOptions.AgentVersion
	expandedOptions.AgentMode = existingOptions.AgentMode
	expandedOptions.ExtGroups = existingOptions.ExtGroups
	expandedOptions.ExtProfile = existingOptions.ExtProfile

	connection.Options = expandedOptions

	return nil
}

func passThroughUnconfigurableConnectionOptionsGoogleApps(
	ctx context.Context,
	api *management.Management,
	connectionID string,
	connection *management.Connection,
) error {
	existingConnection, err := api.Connection.Read(ctx, connectionID)
	if err != nil {
		return err
	}

	if existingConnection.Options == nil {
		return nil
	}

	existingOptions := existingConnection.Options.(*management.ConnectionOptionsGoogleApps)

	expandedOptions := connection.Options.(*management.ConnectionOptionsGoogleApps)
	expandedOptions.AdminAccessToken = existingOptions.AdminAccessToken
	expandedOptions.AdminRefreshToken = existingOptions.AdminRefreshToken
	expandedOptions.AdminAccessTokenExpiresIn = existingOptions.AdminAccessTokenExpiresIn
	expandedOptions.HandleLoginFromSocial = existingOptions.HandleLoginFromSocial

	connection.Options = expandedOptions

	return nil
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

func expandSCIMConfigurationMapping(data *schema.ResourceData) *[]management.SCIMConfigurationMapping {
	srcMapping := data.Get("mapping").(*schema.Set)
	mapping := make([]management.SCIMConfigurationMapping, 0, srcMapping.Len())
	for _, item := range srcMapping.List() {
		srcMap := item.(map[string]interface{})
		mapping = append(mapping, management.SCIMConfigurationMapping{
			Auth0: auth0.String(srcMap["auth0"].(string)),
			SCIM:  auth0.String(srcMap["scim"].(string)),
		})
	}

	return &mapping
}

func expandSCIMConfiguration(data *schema.ResourceData) *management.SCIMConfiguration {
	cfg := data.GetRawConfig()
	scimConfiguration := &management.SCIMConfiguration{}
	if !cfg.GetAttr("user_id_attribute").IsNull() {
		scimConfiguration.UserIDAttribute = auth0.String(data.Get("user_id_attribute").(string))
	}
	if !cfg.GetAttr("mapping").IsNull() && cfg.GetAttr("mapping").AsValueSet().Length() > 0 {
		scimConfiguration.Mapping = expandSCIMConfigurationMapping(data)
	}

	if scimConfiguration.Mapping != nil || scimConfiguration.UserIDAttribute != nil {
		return scimConfiguration
	}

	return nil
}
