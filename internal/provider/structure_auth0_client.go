package provider

import (
	"strconv"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandClient(d *schema.ResourceData) *management.Client {
	config := d.GetRawConfig()

	client := &management.Client{
		Name:                           value.String(config.GetAttr("name")),
		Description:                    value.String(config.GetAttr("description")),
		AppType:                        value.String(config.GetAttr("app_type")),
		LogoURI:                        value.String(config.GetAttr("logo_uri")),
		IsFirstParty:                   value.Bool(config.GetAttr("is_first_party")),
		IsTokenEndpointIPHeaderTrusted: value.Bool(config.GetAttr("is_token_endpoint_ip_header_trusted")),
		OIDCConformant:                 value.Bool(config.GetAttr("oidc_conformant")),
		Callbacks:                      value.Strings(config.GetAttr("callbacks")),
		AllowedLogoutURLs:              value.Strings(config.GetAttr("allowed_logout_urls")),
		AllowedOrigins:                 value.Strings(config.GetAttr("allowed_origins")),
		AllowedClients:                 value.Strings(config.GetAttr("allowed_clients")),
		GrantTypes:                     value.Strings(config.GetAttr("grant_types")),
		OrganizationUsage:              value.String(config.GetAttr("organization_usage")),
		OrganizationRequireBehavior:    value.String(config.GetAttr("organization_require_behavior")),
		WebOrigins:                     value.Strings(config.GetAttr("web_origins")),
		SSO:                            value.Bool(config.GetAttr("sso")),
		SSODisabled:                    value.Bool(config.GetAttr("sso_disabled")),
		CrossOriginAuth:                value.Bool(config.GetAttr("cross_origin_auth")),
		CrossOriginLocation:            value.String(config.GetAttr("cross_origin_loc")),
		CustomLoginPageOn:              value.Bool(config.GetAttr("custom_login_page_on")),
		CustomLoginPage:                value.String(config.GetAttr("custom_login_page")),
		FormTemplate:                   value.String(config.GetAttr("form_template")),
		TokenEndpointAuthMethod:        value.String(config.GetAttr("token_endpoint_auth_method")),
		InitiateLoginURI:               value.String(config.GetAttr("initiate_login_uri")),
		EncryptionKey:                  value.MapOfStrings(config.GetAttr("encryption_key")),
		ClientMetadata:                 value.MapOfStrings(config.GetAttr("client_metadata")),
		RefreshToken:                   expandClientRefreshToken(d),
		JWTConfiguration:               expandClientJWTConfiguration(d),
		Addons:                         expandClientAddons(d),
		NativeSocialLogin:              expandClientNativeSocialLogin(d),
		Mobile:                         expandClientMobile(d),
	}

	return client
}

func expandClientRefreshToken(d *schema.ResourceData) *management.ClientRefreshToken {
	if !d.IsNewResource() || !d.HasChange("refresh_token") {
		return nil
	}

	refreshTokenConfig := d.GetRawConfig().GetAttr("refresh_token")
	if refreshTokenConfig.IsNull() {
		return nil
	}

	var refreshToken management.ClientRefreshToken

	refreshTokenConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		refreshToken.RotationType = value.String(config.GetAttr("rotation_type"))
		refreshToken.ExpirationType = value.String(config.GetAttr("expiration_type"))
		refreshToken.Leeway = value.Int(config.GetAttr("leeway"))
		refreshToken.TokenLifetime = value.Int(config.GetAttr("token_lifetime"))
		refreshToken.InfiniteTokenLifetime = value.Bool(config.GetAttr("infinite_token_lifetime"))
		refreshToken.InfiniteIdleTokenLifetime = value.Bool(config.GetAttr("infinite_idle_token_lifetime"))
		refreshToken.IdleTokenLifetime = value.Int(config.GetAttr("idle_token_lifetime"))
		return stop
	})

	if refreshToken == (management.ClientRefreshToken{}) {
		return nil
	}

	return &refreshToken
}

func expandClientJWTConfiguration(d *schema.ResourceData) *management.ClientJWTConfiguration {
	jwtConfig := d.GetRawConfig().GetAttr("jwt_configuration")
	if jwtConfig.IsNull() {
		return nil
	}

	var jwt management.ClientJWTConfiguration

	jwtConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		jwt.LifetimeInSeconds = value.Int(config.GetAttr("lifetime_in_seconds"))
		jwt.Algorithm = value.String(config.GetAttr("alg"))
		jwt.Scopes = value.MapOfStrings(config.GetAttr("scopes"))

		if d.IsNewResource() {
			jwt.SecretEncoded = value.Bool(config.GetAttr("secret_encoded"))
		}

		return stop
	})

	if jwt == (management.ClientJWTConfiguration{}) {
		return nil
	}

	return &jwt
}

func expandClientNativeSocialLogin(d *schema.ResourceData) *management.ClientNativeSocialLogin {
	nativeSocialLoginConfig := d.GetRawConfig().GetAttr("native_social_login")
	if nativeSocialLoginConfig.IsNull() {
		return nil
	}

	var nativeSocialLogin management.ClientNativeSocialLogin

	nativeSocialLoginConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		nativeSocialLogin.Apple = expandClientNativeSocialLoginSupportEnabled(config.GetAttr("apple"))
		nativeSocialLogin.Facebook = expandClientNativeSocialLoginSupportEnabled(config.GetAttr("facebook"))
		return stop
	})

	if nativeSocialLogin == (management.ClientNativeSocialLogin{}) {
		return nil
	}

	return &nativeSocialLogin
}

func expandClientNativeSocialLoginSupportEnabled(config cty.Value) *management.ClientNativeSocialLoginSupportEnabled {
	if config.IsNull() {
		return nil
	}

	var support management.ClientNativeSocialLoginSupportEnabled

	config.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		support.Enabled = value.Bool(config.GetAttr("enabled"))
		return stop
	})

	return &support
}

func expandClientMobile(d *schema.ResourceData) map[string]interface{} {
	mobileConfig := d.GetRawConfig().GetAttr("mobile")
	if mobileConfig.IsNull() {
		return nil
	}

	mobile := make(map[string]interface{})

	mobileConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		androidConfig := config.GetAttr("android")
		if !androidConfig.IsNull() {
			config.GetAttr("android").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
				android := make(map[string]interface{})

				if appPackageName := value.String(config.GetAttr("app_package_name")); appPackageName != nil {
					android["app_package_name"] = appPackageName
				}
				if cert := value.Strings(config.GetAttr("sha256_cert_fingerprints")); cert != nil {
					android["sha256_cert_fingerprints"] = cert
				}

				if len(android) > 0 {
					mobile["android"] = android
				}

				return stop
			})
		}

		iosConfig := config.GetAttr("ios")
		if !iosConfig.IsNull() {
			config.GetAttr("ios").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
				ios := make(map[string]interface{})

				if teamID := value.String(config.GetAttr("team_id")); teamID != nil {
					ios["team_id"] = teamID
				}
				if appBundleIdentifier := value.String(config.GetAttr("app_bundle_identifier")); appBundleIdentifier != nil {
					ios["app_bundle_identifier"] = appBundleIdentifier
				}

				if len(ios) > 0 {
					mobile["ios"] = ios
				}

				return stop
			})
		}

		return stop
	})

	if len(mobile) > 0 {
		return mobile
	}

	return nil
}

func expandClientAddons(d *schema.ResourceData) map[string]interface{} {
	if !d.HasChange("addons") {
		return nil
	}

	addons := make(map[string]interface{})
	var allowedAddons = []string{
		"aws", "azure_blob", "azure_sb", "rms", "mscrm", "slack", "sentry",
		"box", "cloudbees", "concur", "dropbox", "echosign", "egnyte",
		"firebase", "newrelic", "office365", "salesforce", "salesforce_api",
		"salesforce_sandbox_api", "layer", "sap_api", "sharepoint",
		"springcm", "wams", "wsfed", "zendesk", "zoom",
	}
	for _, name := range allowedAddons {
		if _, ok := d.GetOk("addons.0." + name); ok {
			addons[name] = mapFromState(d.Get("addons.0." + name).(map[string]interface{}))
		}
	}

	samlpConfig := d.GetRawConfig().
		GetAttr("addons").Index(cty.NumberIntVal(0)).
		GetAttr("samlp").Index(cty.NumberIntVal(0))
	samlp := make(map[string]interface{})

	if audience := value.String(samlpConfig.GetAttr("audience")); audience != nil {
		samlp["audience"] = audience
	}
	if authnContextClassRef := value.String(samlpConfig.GetAttr("authn_context_class_ref")); authnContextClassRef != nil {
		samlp["authnContextClassRef"] = authnContextClassRef
	}
	if binding := value.String(samlpConfig.GetAttr("binding")); binding != nil {
		samlp["binding"] = binding
	}
	if signingCert := value.String(samlpConfig.GetAttr("signing_cert")); signingCert != nil {
		samlp["signingCert"] = signingCert
	}
	if destination := value.String(samlpConfig.GetAttr("destination")); destination != nil {
		samlp["destination"] = destination
	}
	if digestAlgorithm := value.String(samlpConfig.GetAttr("digest_algorithm")); digestAlgorithm != nil {
		samlp["digestAlgorithm"] = digestAlgorithm
	}
	if nameIdentifierFormat := value.String(samlpConfig.GetAttr("name_identifier_format")); nameIdentifierFormat != nil {
		samlp["nameIdentifierFormat"] = nameIdentifierFormat
	}
	if recipient := value.String(samlpConfig.GetAttr("recipient")); recipient != nil {
		samlp["recipient"] = recipient
	}
	if signatureAlgorithm := value.String(samlpConfig.GetAttr("signature_algorithm")); signatureAlgorithm != nil {
		samlp["signatureAlgorithm"] = signatureAlgorithm
	}
	if createUpnClaim := value.Bool(samlpConfig.GetAttr("create_upn_claim")); createUpnClaim != nil {
		samlp["createUpnClaim"] = createUpnClaim
	}
	if includeAttributeNameFormat := value.Bool(samlpConfig.GetAttr("include_attribute_name_format")); includeAttributeNameFormat != nil {
		samlp["includeAttributeNameFormat"] = includeAttributeNameFormat
	}
	if mapIdentities := value.Bool(samlpConfig.GetAttr("map_identities")); mapIdentities != nil {
		samlp["mapIdentities"] = mapIdentities
	}
	if mapUnknownClaimsAsIs := value.Bool(samlpConfig.GetAttr("map_unknown_claims_as_is")); mapUnknownClaimsAsIs != nil {
		samlp["mapUnknownClaimsAsIs"] = mapUnknownClaimsAsIs
	}
	if passthroughClaimsWithNoMapping := value.Bool(samlpConfig.GetAttr("passthrough_claims_with_no_mapping")); passthroughClaimsWithNoMapping != nil {
		samlp["passthroughClaimsWithNoMapping"] = passthroughClaimsWithNoMapping
	}
	if signResponse := value.Bool(samlpConfig.GetAttr("sign_response")); signResponse != nil {
		samlp["signResponse"] = signResponse
	}
	if typedAttributes := value.Bool(samlpConfig.GetAttr("typed_attributes")); typedAttributes != nil {
		samlp["typedAttributes"] = typedAttributes
	}
	if lifetimeInSeconds := value.Int(samlpConfig.GetAttr("lifetime_in_seconds")); lifetimeInSeconds != nil {
		samlp["lifetimeInSeconds"] = lifetimeInSeconds
	}
	if mappings := value.MapOfStrings(samlpConfig.GetAttr("mappings")); mappings != nil {
		samlp["mappings"] = mappings
	}
	if nameIdentifierProbes := value.Strings(samlpConfig.GetAttr("name_identifier_probes")); nameIdentifierProbes != nil {
		samlp["nameIdentifierProbes"] = nameIdentifierProbes
	}
	if logout := mapFromState(d.Get("addons.0.samlp.0.logout").(map[string]interface{})); logout != nil {
		samlp["logout"] = logout
	}

	if len(samlp) > 0 {
		addons["samlp"] = samlp
	}

	if len(addons) > 0 {
		return addons
	}

	return nil
}

func mapFromState(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for key, val := range input {
		switch v := val.(type) {
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				output[key] = i
			} else if f, err := strconv.ParseFloat(v, 64); err == nil {
				output[key] = f
			} else if b, err := strconv.ParseBool(v); err == nil {
				output[key] = b
			} else {
				output[key] = v
			}
		case map[string]interface{}:
			output[key] = mapFromState(v)
		case []interface{}:
			output[key] = v
		default:
			output[key] = v
		}
	}

	return output
}

func flattenCustomSocialConfiguration(customSocial *management.ClientNativeSocialLogin) []interface{} {
	if customSocial == nil {
		return nil
	}

	m := make(map[string]interface{})

	m["apple"] = []interface{}{
		map[string]interface{}{
			"enabled": customSocial.GetApple().GetEnabled(),
		},
	}
	m["facebook"] = []interface{}{
		map[string]interface{}{
			"enabled": customSocial.GetFacebook().GetEnabled(),
		},
	}

	return []interface{}{m}
}

func flattenClientJwtConfiguration(jwt *management.ClientJWTConfiguration) []interface{} {
	if jwt == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"lifetime_in_seconds": jwt.GetLifetimeInSeconds(),
			"secret_encoded":      jwt.GetSecretEncoded(),
			"scopes":              jwt.GetScopes(),
			"alg":                 jwt.GetAlgorithm(),
		},
	}
}

func flattenClientRefreshTokenConfiguration(refreshToken *management.ClientRefreshToken) []interface{} {
	if refreshToken == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"rotation_type":                refreshToken.GetRotationType(),
			"expiration_type":              refreshToken.GetExpirationType(),
			"leeway":                       refreshToken.GetLeeway(),
			"token_lifetime":               refreshToken.GetTokenLifetime(),
			"infinite_token_lifetime":      refreshToken.GetInfiniteTokenLifetime(),
			"infinite_idle_token_lifetime": refreshToken.GetInfiniteIdleTokenLifetime(),
			"idle_token_lifetime":          refreshToken.GetIdleTokenLifetime(),
		},
	}
}

func flattenClientAddons(addons map[string]interface{}) []interface{} {
	if addons == nil {
		return nil
	}

	m := make(map[string]interface{})

	if v, ok := addons["samlp"]; ok {
		samlp := v.(map[string]interface{})

		samlpMap := map[string]interface{}{
			"audience":                           samlp["audience"],
			"recipient":                          samlp["recipient"],
			"mappings":                           samlp["mappings"],
			"create_upn_claim":                   samlp["createUpnClaim"],
			"passthrough_claims_with_no_mapping": samlp["passthroughClaimsWithNoMapping"],
			"map_unknown_claims_as_is":           samlp["mapUnknownClaimsAsIs"],
			"map_identities":                     samlp["mapIdentities"],
			"signature_algorithm":                samlp["signatureAlgorithm"],
			"digest_algorithm":                   samlp["digestAlgorithm"],
			"destination":                        samlp["destination"],
			"lifetime_in_seconds":                samlp["lifetimeInSeconds"],
			"sign_response":                      samlp["signResponse"],
			"name_identifier_format":             samlp["nameIdentifierFormat"],
			"name_identifier_probes":             samlp["nameIdentifierProbes"],
			"authn_context_class_ref":            samlp["authnContextClassRef"],
			"typed_attributes":                   samlp["typedAttributes"],
			"include_attribute_name_format":      samlp["includeAttributeNameFormat"],
			"binding":                            samlp["binding"],
			"signing_cert":                       samlp["signingCert"],
		}

		if logout, ok := samlp["logout"].(map[string]interface{}); ok {
			samlpMap["logout"] = mapToState(logout)
		}

		m["samlp"] = []interface{}{samlpMap}
	}

	for _, name := range []string{
		"aws", "azure_blob", "azure_sb", "rms", "mscrm", "slack", "sentry",
		"box", "cloudbees", "concur", "dropbox", "echosign", "egnyte",
		"firebase", "newrelic", "office365", "salesforce", "salesforce_api",
		"salesforce_sandbox_api", "layer", "sap_api", "sharepoint",
		"springcm", "wams", "wsfed", "zendesk", "zoom",
	} {
		if v, ok := addons[name]; ok {
			if addonType, ok := v.(map[string]interface{}); ok {
				m[name] = mapToState(addonType)
			}
		}
	}

	return []interface{}{m}
}

func flattenClientMobile(mobile map[string]interface{}) []interface{} {
	if mobile == nil {
		return nil
	}

	m := make(map[string]interface{})

	if value, ok := mobile["android"]; ok {
		android := value.(map[string]interface{})

		m["android"] = []interface{}{
			map[string]interface{}{
				"app_package_name":         android["app_package_name"],
				"sha256_cert_fingerprints": android["sha256_cert_fingerprints"],
			},
		}
	}

	if value, ok := mobile["ios"]; ok {
		ios := value.(map[string]interface{})

		m["ios"] = []interface{}{
			map[string]interface{}{
				"team_id":               ios["team_id"],
				"app_bundle_identifier": ios["app_bundle_identifier"],
			},
		}
	}

	return []interface{}{m}
}

func mapToState(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for key, v := range input {
		switch val := v.(type) {
		case bool:
			if val {
				output[key] = "true"
			} else {
				output[key] = "false"
			}
		case float64:
			output[key] = strconv.Itoa(int(val))
		case int:
			output[key] = strconv.Itoa(val)
		default:
			output[key] = val
		}
	}

	return output
}
