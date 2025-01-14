package connection

import (
	"errors"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

var errUnsupportedConnectionOptionsType = errors.New("unsupported connection options type")

var flattenConnectionOptionsMap = map[string]flattenConnectionOptionsFunc{
	// Database Connection.
	management.ConnectionStrategyAuth0: flattenConnectionOptionsAuth0,

	// Social Connections.
	management.ConnectionStrategyGoogleOAuth2:        flattenConnectionOptionsGoogleOAuth2,
	management.ConnectionStrategyOAuth2:              flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyDropbox:             flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyBitBucket:           flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyPaypal:              flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyTwitter:             flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyAmazon:              flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyYahoo:               flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyBox:                 flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyWordpress:           flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyShopify:             flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyLine:                flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyCustom:              flattenConnectionOptionsOAuth2,
	management.ConnectionStrategyFacebook:            flattenConnectionOptionsFacebook,
	management.ConnectionStrategyApple:               flattenConnectionOptionsApple,
	management.ConnectionStrategyLinkedin:            flattenConnectionOptionsLinkedin,
	management.ConnectionStrategyGitHub:              flattenConnectionOptionsGitHub,
	management.ConnectionStrategyWindowsLive:         flattenConnectionOptionsWindowsLive,
	management.ConnectionStrategySalesforce:          flattenConnectionOptionsSalesforce,
	management.ConnectionStrategySalesforceCommunity: flattenConnectionOptionsSalesforce,
	management.ConnectionStrategySalesforceSandbox:   flattenConnectionOptionsSalesforce,

	// Passwordless Connections.
	management.ConnectionStrategySMS:   flattenConnectionOptionsSMS,
	management.ConnectionStrategyEmail: flattenConnectionOptionsEmail,

	// Enterprise Connections.
	management.ConnectionStrategyOIDC:         flattenConnectionOptionsOIDC,
	management.ConnectionStrategyGoogleApps:   flattenConnectionOptionsGoogleApps,
	management.ConnectionStrategyOkta:         flattenConnectionOptionsOkta,
	management.ConnectionStrategyAD:           flattenConnectionOptionsAD,
	management.ConnectionStrategyAzureAD:      flattenConnectionOptionsAzureAD,
	management.ConnectionStrategySAML:         flattenConnectionOptionsSAML,
	management.ConnectionStrategyADFS:         flattenConnectionOptionsADFS,
	management.ConnectionStrategyPingFederate: flattenConnectionOptionsPingFederate,
}

type flattenConnectionOptionsFunc func(data *schema.ResourceData, options interface{}) (interface{}, diag.Diagnostics)

func flattenConnection(data *schema.ResourceData, connection *management.Connection) diag.Diagnostics {
	connectionOptions, diags := flattenConnectionOptions(data, connection)
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
	)

	if connectionIsEnterprise(connection.GetStrategy()) {
		result = multierror.Append(result, data.Set("show_as_button", connection.GetShowAsButton()))
	}

	return diag.FromErr(result.ErrorOrNil())
}

func flattenConnectionForDataSource(data *schema.ResourceData, connection *management.Connection) diag.Diagnostics {
	diags := flattenConnection(data, connection)

	err := data.Set("enabled_clients", connection.GetEnabledClients())

	diags = append(diags, diag.FromErr(err)...)

	return diags
}

func flattenConnectionOptions(data *schema.ResourceData, connection *management.Connection) ([]interface{}, diag.Diagnostics) {
	if connection == nil || connection.Options == nil {
		return nil, nil
	}

	connectionOptionsFunc, ok := flattenConnectionOptionsMap[connection.GetStrategy()]
	if !ok {
		return nil, diag.Diagnostics{
			{
				Severity: diag.Error,
				Summary:  "Unsupported Connection Strategy",
				Detail: fmt.Sprintf(
					"Raise an issue at %s in order to have the following connection strategy supported: %q",
					"https://github.com/auth0/terraform-provider-auth0/issues/new",
					connection.GetStrategy(),
				),
				AttributePath: cty.Path{cty.GetAttrStep{Name: "strategy"}},
			},
		}
	}

	connectionOptionsMap, diagnostics := connectionOptionsFunc(data, connection.Options)

	return []interface{}{connectionOptionsMap}, diagnostics
}

func flattenConnectionOptionsGitHub(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsGitHub)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"scopes":                   options.Scopes(),
		"upstream_params":          upstreamParams,
	}

	return optionsMap, nil
}

func flattenConnectionOptionsWindowsLive(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsWindowsLive)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"strategy_version":         options.GetStrategyVersion(),
		"upstream_params":          upstreamParams,
	}

	return optionsMap, nil
}

func flattenAttributes(connAttributes *management.ConnectionOptionsAttributes) interface{} {
	if connAttributes == nil {
		return nil
	}

	return map[string]interface{}{
		"email":        flattenEmailAttribute(connAttributes.Email),
		"username":     flattenUsernameAttribute(connAttributes.Username),
		"phone_number": flattenPhoneNumberAttribute(connAttributes.PhoneNumber),
	}
}

func flattenEmailAttribute(emailAttribute *management.ConnectionOptionsEmailAttribute) []map[string]interface{} {
	if emailAttribute == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"identifier":       flattenIdentifier(emailAttribute.GetIdentifier()),
			"profile_required": emailAttribute.GetProfileRequired(),
			"signup":           flattenSignUp(emailAttribute.GetSignup()),
		},
	}
}

func flattenUsernameAttribute(usernameAttribute *management.ConnectionOptionsUsernameAttribute) []map[string]interface{} {
	if usernameAttribute == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"identifier":       flattenIdentifier(usernameAttribute.GetIdentifier()),
			"profile_required": usernameAttribute.GetProfileRequired(),
			"signup":           flattenUsernameSignUp(usernameAttribute.GetSignup()),
			"validation":       flattenValidation(usernameAttribute.GetValidation()),
		},
	}
}

func flattenPhoneNumberAttribute(phoneNumberAttribute *management.ConnectionOptionsPhoneNumberAttribute) []map[string]interface{} {
	if phoneNumberAttribute == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"identifier":       flattenIdentifier(phoneNumberAttribute.GetIdentifier()),
			"profile_required": phoneNumberAttribute.GetProfileRequired(),
			"signup":           flattenSignUp(phoneNumberAttribute.GetSignup()),
		},
	}
}

func flattenIdentifier(identifier *management.ConnectionOptionsAttributeIdentifier) []map[string]interface{} {
	if identifier == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"active": identifier.GetActive(),
		},
	}
}

func flattenSignUp(signup *management.ConnectionOptionsAttributeSignup) []map[string]interface{} {
	if signup == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"status":       signup.GetStatus(),
			"verification": flattenVerification(signup.GetVerification()),
		},
	}
}

func flattenUsernameSignUp(signup *management.ConnectionOptionsAttributeSignup) []map[string]interface{} {
	if signup == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"status": signup.GetStatus(),
		},
	}
}

func flattenValidation(validation *management.ConnectionOptionsAttributeValidation) []map[string]interface{} {
	if validation == nil {
		return nil
	}
	return []map[string]interface{}{
		{
			"min_length": validation.GetMinLength(),
			"max_length": validation.GetMaxLength(),
			"allowed_types": []map[string]interface{}{
				{
					"email":        validation.GetAllowedTypes().GetEmail(),
					"phone_number": validation.GetAllowedTypes().GetPhoneNumber(),
				},
			},
		},
	}
}

func flattenVerification(verification *management.ConnectionOptionsAttributeVerification) []map[string]interface{} {
	if verification == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"active": verification.GetActive(),
		},
	}
}

func flattenConnectionOptionsAuth0(
	data *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptions)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	dbSecretConfig, ok := data.GetOk("options.0.configuration")
	if !ok {
		dbSecretConfig = make(map[string]interface{})
	}

	optionsMap := map[string]interface{}{
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
		"upstream_params":                      upstreamParams,
		"precedence":                           options.GetPrecedence(),
		"strategy_version":                     options.GetStrategyVersion(),
	}

	if options.Attributes != nil {
		optionsMap["attributes"] = []interface{}{flattenAttributes(options.GetAttributes())}
	}

	if options.PasswordComplexityOptions != nil {
		optionsMap["password_complexity_options"] = []interface{}{options.PasswordComplexityOptions}
	}

	if options.PasswordDictionary != nil {
		optionsMap["password_dictionary"] = []interface{}{options.PasswordDictionary}
	}

	if options.PasswordNoPersonalInfo != nil {
		optionsMap["password_no_personal_info"] = []interface{}{options.PasswordNoPersonalInfo}
	}

	if options.PasswordHistory != nil {
		optionsMap["password_history"] = []interface{}{options.PasswordHistory}
	}

	if options.MFA != nil {
		optionsMap["mfa"] = []interface{}{options.MFA}
	}

	if options.Validation != nil {
		optionsMap["validation"] = []interface{}{
			map[string]interface{}{
				"username": []interface{}{
					options.Validation["username"],
				},
			},
		}
	}

	return optionsMap, nil
}

func flattenConnectionOptionsGoogleOAuth2(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsGoogleOAuth2)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"allowed_audiences":        options.GetAllowedAudiences(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"upstream_params":          upstreamParams,
	}

	return optionsMap, nil
}

func flattenConnectionOptionsGoogleApps(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsGoogleApps)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"domain":                   options.GetDomain(),
		"tenant_domain":            options.GetTenantDomain(),
		"api_enable_users":         options.GetEnableUsersAPI(),
		"scopes":                   options.Scopes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"domain_aliases":           options.GetDomainAliases(),
		"icon_url":                 options.GetLogoURL(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"map_user_id_to_id":        options.GetMapUserIDtoID(),
		"upstream_params":          upstreamParams,
	}

	if options.GetSetUserAttributes() == "" {
		optionsMap["set_user_root_attributes"] = "on_each_login"
	}

	return optionsMap, nil
}

func flattenConnectionOptionsOAuth2(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsOAuth2)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
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
		"strategy_version":         options.GetStrategyVersion(),
		"upstream_params":          upstreamParams,
	}

	return optionsMap, nil
}

func flattenConnectionOptionsFacebook(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsFacebook)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"upstream_params":          upstreamParams,
	}

	return optionsMap, nil
}

func flattenConnectionOptionsApple(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsApple)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"team_id":                  options.GetTeamID(),
		"key_id":                   options.GetKeyID(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"upstream_params":          upstreamParams,
	}

	return optionsMap, nil
}

func flattenConnectionOptionsLinkedin(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsLinkedin)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"strategy_version":         options.GetStrategyVersion(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"upstream_params":          upstreamParams,
	}

	return optionsMap, nil
}

func flattenConnectionOptionsSalesforce(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsSalesforce)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"community_base_url":       options.GetCommunityBaseURL(),
		"scopes":                   options.Scopes(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"upstream_params":          upstreamParams,
	}

	return optionsMap, nil
}

func flattenConnectionOptionsSMS(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsSMS)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
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
		"upstream_params":        upstreamParams,
	}

	if options.OTP != nil {
		optionsMap["totp"] = []interface{}{
			map[string]interface{}{
				"time_step": options.OTP.GetTimeStep(),
				"length":    options.OTP.GetLength(),
			},
		}
	}

	if options.GatewayAuthentication != nil {
		optionsMap["gateway_authentication"] = []interface{}{
			map[string]interface{}{
				"method":                options.GatewayAuthentication.GetMethod(),
				"subject":               options.GatewayAuthentication.GetSubject(),
				"audience":              options.GatewayAuthentication.GetAudience(),
				"secret":                options.GatewayAuthentication.GetSecret(),
				"secret_base64_encoded": options.GatewayAuthentication.GetSecretBase64Encoded(),
			},
		}
	}

	return optionsMap, nil
}

func flattenConnectionOptionsOIDC(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsOIDC)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
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
		"upstream_params":          upstreamParams,
	}

	attributes, err := structure.FlattenJsonToString(options.GetAttributeMap().GetAttributes())
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if options.AttributeMap != nil {
		optionsMap["attribute_map"] = []map[string]interface{}{
			{
				"mapping_mode":   options.GetAttributeMap().GetMappingMode(),
				"userinfo_scope": options.GetAttributeMap().GetUserInfoScope(),
				"attributes":     attributes,
			},
		}
	}

	if options.ConnectionSettings != nil {
		optionsMap["connection_settings"] = []map[string]string{
			{
				"pkce": options.GetConnectionSettings().GetPKCE(),
			},
		}
	}

	return optionsMap, nil
}

func flattenConnectionOptionsOkta(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsOkta)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
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
		"upstream_params":          upstreamParams,
	}

	attributes, err := structure.FlattenJsonToString(options.GetAttributeMap().GetAttributes())
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if options.AttributeMap != nil {
		optionsMap["attribute_map"] = []map[string]interface{}{
			{
				"mapping_mode":   options.GetAttributeMap().GetMappingMode(),
				"userinfo_scope": options.GetAttributeMap().GetUserInfoScope(),
				"attributes":     attributes,
			},
		}
	}

	if options.ConnectionSettings != nil {
		optionsMap["connection_settings"] = []map[string]string{
			{
				"pkce": options.GetConnectionSettings().GetPKCE(),
			},
		}
	}

	return optionsMap, nil
}

func flattenConnectionOptionsEmail(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsEmail)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"name":                     options.GetName(),
		"from":                     options.GetEmail().GetFrom(),
		"syntax":                   options.GetEmail().GetSyntax(),
		"subject":                  options.GetEmail().GetSubject(),
		"template":                 options.GetEmail().GetBody(),
		"disable_signup":           options.GetDisableSignup(),
		"brute_force_protection":   options.GetBruteForceProtection(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"upstream_params":          upstreamParams,
	}

	if options.OTP != nil {
		optionsMap["totp"] = []interface{}{
			map[string]interface{}{
				"time_step": options.OTP.GetTimeStep(),
				"length":    options.OTP.GetLength(),
			},
		}
	}

	if options.AuthParams != nil {
		v, ok := options.AuthParams.(map[string]interface{})
		if !ok {
			return optionsMap, diag.Diagnostics{{
				Severity:      diag.Warning,
				Summary:       "Unable to cast auth_params to map[string]string",
				Detail:        fmt.Sprintf(`Authentication Parameters are required to be a map of strings, the existing value of %v is not compatible. It is recommended to express the existing value as a valid map[string]string. Subsequent terraform applys will clear this configuration to empty map.`, options.AuthParams),
				AttributePath: cty.Path{cty.GetAttrStep{Name: "options.auth_params"}},
			}}
		}

		optionsMap["auth_params"] = v
	}

	return optionsMap, nil
}

func flattenConnectionOptionsAD(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsAD)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"tenant_domain":                        options.GetTenantDomain(),
		"domain_aliases":                       options.GetDomainAliases(),
		"icon_url":                             options.GetLogoURL(),
		"ips":                                  options.GetIPs(),
		"use_cert_auth":                        options.GetCertAuth(),
		"use_kerberos":                         options.GetKerberos(),
		"disable_cache":                        options.GetDisableCache(),
		"brute_force_protection":               options.GetBruteForceProtection(),
		"non_persistent_attrs":                 options.GetNonPersistentAttrs(),
		"set_user_root_attributes":             options.GetSetUserAttributes(),
		"disable_self_service_change_password": options.GetDisableSelfServiceChangePassword(),
		"strategy_version":                     options.GetStrategyVersion(),
		"upstream_params":                      upstreamParams,
	}

	if options.GetSetUserAttributes() == "" {
		optionsMap["set_user_root_attributes"] = "on_each_login"
	}

	return optionsMap, nil
}

func flattenConnectionOptionsAzureAD(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsAzureAD)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
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
		"set_user_root_attributes":               options.GetSetUserAttributes(),
		"strategy_version":                       options.GetStrategyVersion(),
		"user_id_attribute":                      options.GetUserIDAttribute(),
		"upstream_params":                        upstreamParams,
	}

	if options.GetSetUserAttributes() == "" {
		optionsMap["set_user_root_attributes"] = "on_each_login"
	}

	return optionsMap, nil
}

func flattenConnectionOptionsADFS(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsADFS)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"tenant_domain":                          options.GetTenantDomain(),
		"domain_aliases":                         options.GetDomainAliases(),
		"icon_url":                               options.GetLogoURL(),
		"adfs_server":                            options.GetADFSServer(),
		"fed_metadata_xml":                       options.GetFedMetadataXML(),
		"sign_in_endpoint":                       options.GetSignInEndpoint(),
		"api_enable_users":                       options.GetEnableUsersAPI(),
		"should_trust_email_verified_connection": options.GetTrustEmailVerified(),
		"non_persistent_attrs":                   options.GetNonPersistentAttrs(),
		"set_user_root_attributes":               options.GetSetUserAttributes(),
		"strategy_version":                       options.GetStrategyVersion(),
		"upstream_params":                        upstreamParams,
	}

	if options.GetSetUserAttributes() == "" {
		optionsMap["set_user_root_attributes"] = "on_each_login"
	}

	return optionsMap, nil
}

func flattenConnectionOptionsSAML(
	data *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsSAML)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	fieldsMap, err := structure.FlattenJsonToString(options.FieldsMap)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
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
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"entity_id":                options.GetEntityID(),
		"metadata_url":             options.GetMetadataURL(),
		"metadata_xml":             data.Get("options.0.metadata_xml").(string), // Does not get read back.
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"strategy_version":         options.GetStrategyVersion(),
		"fields_map":               fieldsMap,
		"upstream_params":          upstreamParams,
	}

	if options.GetSetUserAttributes() == "" {
		optionsMap["set_user_root_attributes"] = "on_each_login"
	}

	if options.IdpInitiated != nil {
		optionsMap["idp_initiated"] = []interface{}{
			map[string]interface{}{
				"enabled":                options.IdpInitiated.GetEnabled(),
				"client_id":              options.IdpInitiated.GetClientID(),
				"client_protocol":        options.IdpInitiated.GetClientProtocol(),
				"client_authorize_query": options.IdpInitiated.GetClientAuthorizeQuery(),
			},
		}
	}

	if options.SigningKey != nil {
		optionsMap["signing_key"] = []interface{}{
			map[string]interface{}{
				"key":  options.GetSigningKey().GetKey(),
				"cert": options.GetSigningKey().GetCert(),
			},
		}
	}

	if options.DecryptionKey != nil {
		optionsMap["decryption_key"] = []interface{}{
			map[string]interface{}{
				"key":  options.GetDecryptionKey().GetKey(),
				"cert": options.GetDecryptionKey().GetCert(),
			},
		}
	}

	return optionsMap, nil
}

func flattenConnectionOptionsPingFederate(
	_ *schema.ResourceData,
	rawOptions interface{},
) (interface{}, diag.Diagnostics) {
	options, ok := rawOptions.(*management.ConnectionOptionsPingFederate)
	if !ok {
		return nil, diag.FromErr(errUnsupportedConnectionOptionsType)
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	optionsMap := map[string]interface{}{
		"signing_cert":             options.GetSigningCert(),
		"tenant_domain":            options.GetTenantDomain(),
		"domain_aliases":           options.GetDomainAliases(),
		"sign_in_endpoint":         options.GetSignInEndpoint(),
		"signature_algorithm":      options.GetSignatureAlgorithm(),
		"digest_algorithm":         options.GetDigestAlgorithm(),
		"sign_saml_request":        options.GetSignSAMLRequest(),
		"ping_federate_base_url":   options.GetPingFederateBaseURL(),
		"icon_url":                 options.GetLogoURL(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"upstream_params":          upstreamParams,
		"idp_initiated": []map[string]interface{}{
			{
				"enabled":                options.GetIdpInitiated().GetEnabled(),
				"client_id":              options.GetIdpInitiated().GetClientID(),
				"client_protocol":        options.GetIdpInitiated().GetClientProtocol(),
				"client_authorize_query": options.GetIdpInitiated().GetClientAuthorizeQuery(),
			},
		},
	}

	if options.GetSigningCert() == "" {
		optionsMap["signing_cert"] = options.GetCert()
	}

	if options.GetSetUserAttributes() == "" {
		optionsMap["set_user_root_attributes"] = "on_each_login"
	}

	return optionsMap, nil
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

func flattenSCIMMappings(mappings []management.SCIMConfigurationMapping) *[]map[string]string {
	flattenedMappings := make([]map[string]string, 0, len(mappings))
	for _, mapping := range mappings {
		flattenedMappings = append(flattenedMappings, map[string]string{
			"auth0": mapping.GetAuth0(),
			"scim":  mapping.GetSCIM(),
		})
	}

	return &flattenedMappings
}

func flattenSCIMConfiguration(data *schema.ResourceData, scimConfiguration *management.SCIMConfiguration) diag.Diagnostics {
	result := multierror.Append(
		data.Set("connection_id", scimConfiguration.GetConnectionID()),
		data.Set("connection_name", scimConfiguration.GetConnectionName()),
		data.Set("user_id_attribute", scimConfiguration.GetUserIDAttribute()),
		data.Set("mapping", flattenSCIMMappings(scimConfiguration.GetMapping())),
		data.Set("strategy", scimConfiguration.GetStrategy()),
		data.Set("tenant_name", scimConfiguration.GetTenantName()),
	)

	return diag.FromErr(result.ErrorOrNil())
}
