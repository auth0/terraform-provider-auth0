package client

import (
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
		ClientAliases:                  value.Strings(config.GetAttr("client_aliases")),
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
		OIDCBackchannelLogout:          expandOIDCBackchannelLogout(d),
		ClientMetadata:                 expandClientMetadata(d),
		RefreshToken:                   expandClientRefreshToken(d),
		JWTConfiguration:               expandClientJWTConfiguration(d),
		//Addons:                         expandClientAddons(d), TODO: DXCDT-441.
		NativeSocialLogin: expandClientNativeSocialLogin(d),
		Mobile:            expandClientMobile(d),
	}

	return client
}

func expandOIDCBackchannelLogout(d *schema.ResourceData) *management.OIDCBackchannelLogout {
	raw := d.GetRawConfig().GetAttr("oidc_backchannel_logout_urls")

	logoutUrls := value.Strings(raw)

	if logoutUrls == nil {
		return nil
	}

	return &management.OIDCBackchannelLogout{
		BackChannelLogoutURLs: logoutUrls,
	}
}

func expandClientRefreshToken(d *schema.ResourceData) *management.ClientRefreshToken {
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

	if support == (management.ClientNativeSocialLoginSupportEnabled{}) {
		return nil
	}

	return &support
}

func expandClientMobile(d *schema.ResourceData) *management.ClientMobile {
	mobileConfig := d.GetRawConfig().GetAttr("mobile")
	if mobileConfig.IsNull() {
		return nil
	}

	var mobile management.ClientMobile

	mobileConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		mobile.Android = expandClientMobileAndroid(config.GetAttr("android"))
		mobile.IOS = expandClientMobileIOS(config.GetAttr("ios"))
		return stop
	})

	if mobile == (management.ClientMobile{}) {
		return nil
	}

	return &mobile
}

func expandClientMobileAndroid(androidConfig cty.Value) *management.ClientMobileAndroid {
	if androidConfig.IsNull() {
		return nil
	}

	var android management.ClientMobileAndroid

	androidConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		android.AppPackageName = value.String(config.GetAttr("app_package_name"))
		android.KeyHashes = value.Strings(config.GetAttr("sha256_cert_fingerprints"))
		return stop
	})

	if android == (management.ClientMobileAndroid{}) {
		return nil
	}

	return &android
}

func expandClientMobileIOS(iosConfig cty.Value) *management.ClientMobileIOS {
	if iosConfig.IsNull() {
		return nil
	}

	var ios management.ClientMobileIOS

	iosConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		ios.TeamID = value.String(config.GetAttr("team_id"))
		ios.AppID = value.String(config.GetAttr("app_bundle_identifier"))
		return stop
	})

	if ios == (management.ClientMobileIOS{}) {
		return nil
	}

	return &ios
}

func expandClientMetadata(d *schema.ResourceData) *map[string]interface{} {
	if !d.HasChange("client_metadata") {
		return nil
	}

	oldMetadata, newMetadata := d.GetChange("client_metadata")
	oldMetadataMap := oldMetadata.(map[string]interface{})
	newMetadataMap := newMetadata.(map[string]interface{})

	for key := range oldMetadataMap {
		if _, ok := newMetadataMap[key]; !ok {
			newMetadataMap[key] = nil
		}
	}

	return &newMetadataMap
}

//	if !d.HasChange("addons") {
//		return nil
//	}
//
//	addons := make(map[string]interface{})
//	var allowedAddons = []string{
//		"aws", "azure_blob", "azure_sb", "rms", "mscrm", "slack", "sentry",
//		"box", "cloudbees", "concur", "dropbox", "echosign", "egnyte",
//		"firebase", "newrelic", "office365", "salesforce", "salesforce_api",
//		"salesforce_sandbox_api", "layer", "sap_api", "sharepoint",
//		"springcm", "wams", "wsfed", "zendesk", "zoom",
//	}
//	for _, name := range allowedAddons {
//		if _, ok := d.GetOk("addons.0." + name); ok {
//			addons[name] = mapFromState(d.Get("addons.0." + name).(map[string]interface{}))
//		}
//	}
//
//	addonsConfig := d.GetRawConfig().GetAttr("addons")
//	if addonsConfig.IsNull() {
//		return addons
//	}
//
//	addonsConfig.ForEachElement(func(_ cty.Value, addonsConfig cty.Value) (stop bool) {
//		samlpConfig := addonsConfig.GetAttr("samlp")
//		if samlpConfig.IsNull() {
//			return stop
//		}
//
//		samlp := make(map[string]interface{})
//
//		samlpConfig.ForEachElement(func(_ cty.Value, samlpConfig cty.Value) (stop bool) {
//			if issuer := value.String(samlpConfig.GetAttr("issuer")); issuer != nil {
//				samlp["issuer"] = issuer
//			}
//			if audience := value.String(samlpConfig.GetAttr("audience")); audience != nil {
//				samlp["audience"] = audience
//			}
//			if authnContextClassRef := value.String(samlpConfig.GetAttr("authn_context_class_ref")); authnContextClassRef != nil {
//				samlp["authnContextClassRef"] = authnContextClassRef
//			}
//			if binding := value.String(samlpConfig.GetAttr("binding")); binding != nil {
//				samlp["binding"] = binding
//			}
//			if signingCert := value.String(samlpConfig.GetAttr("signing_cert")); signingCert != nil {
//				samlp["signingCert"] = signingCert
//			}
//			if destination := value.String(samlpConfig.GetAttr("destination")); destination != nil {
//				samlp["destination"] = destination
//			}
//
//			digestAlgorithm := value.String(samlpConfig.GetAttr("digest_algorithm"))
//			samlp["digestAlgorithm"] = digestAlgorithm
//			if digestAlgorithm == nil {
//				samlp["digestAlgorithm"] = "sha1"
//			}
//
//			nameIdentifierFormat := value.String(samlpConfig.GetAttr("name_identifier_format"))
//			samlp["nameIdentifierFormat"] = nameIdentifierFormat
//			if nameIdentifierFormat == nil {
//				samlp["nameIdentifierFormat"] = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
//			}
//
//			if recipient := value.String(samlpConfig.GetAttr("recipient")); recipient != nil {
//				samlp["recipient"] = recipient
//			}
//
//			signatureAlgorithm := value.String(samlpConfig.GetAttr("signature_algorithm"))
//			samlp["signatureAlgorithm"] = signatureAlgorithm
//			if signatureAlgorithm == nil {
//				samlp["signatureAlgorithm"] = "rsa-sha1"
//			}
//
//			createUpnClaim := value.Bool(samlpConfig.GetAttr("create_upn_claim"))
//			samlp["createUpnClaim"] = createUpnClaim
//			if createUpnClaim == nil {
//				samlp["createUpnClaim"] = true
//			}
//
//			includeAttributeNameFormat := value.Bool(samlpConfig.GetAttr("include_attribute_name_format"))
//			samlp["includeAttributeNameFormat"] = includeAttributeNameFormat
//			if includeAttributeNameFormat == nil {
//				samlp["includeAttributeNameFormat"] = true
//			}
//
//			mapIdentities := value.Bool(samlpConfig.GetAttr("map_identities"))
//			samlp["mapIdentities"] = mapIdentities
//			if mapIdentities == nil {
//				samlp["mapIdentities"] = true
//			}
//
//			mapUnknownClaimsAsIs := value.Bool(samlpConfig.GetAttr("map_unknown_claims_as_is"))
//			samlp["mapUnknownClaimsAsIs"] = mapUnknownClaimsAsIs
//			if mapUnknownClaimsAsIs == nil {
//				samlp["mapUnknownClaimsAsIs"] = false
//			}
//
//			passthroughClaimsWithNoMapping := value.Bool(samlpConfig.GetAttr("passthrough_claims_with_no_mapping"))
//			samlp["passthroughClaimsWithNoMapping"] = passthroughClaimsWithNoMapping
//			if passthroughClaimsWithNoMapping == nil {
//				samlp["passthroughClaimsWithNoMapping"] = true
//			}
//
//			if signResponse := value.Bool(samlpConfig.GetAttr("sign_response")); signResponse != nil {
//				samlp["signResponse"] = signResponse
//			}
//
//			typedAttributes := value.Bool(samlpConfig.GetAttr("typed_attributes"))
//			samlp["typedAttributes"] = typedAttributes
//			if typedAttributes == nil {
//				samlp["typedAttributes"] = true
//			}
//
//			lifetimeInSeconds := value.Int(samlpConfig.GetAttr("lifetime_in_seconds"))
//			samlp["lifetimeInSeconds"] = lifetimeInSeconds
//			if lifetimeInSeconds == nil {
//				samlp["lifetimeInSeconds"] = 3600
//			}
//
//			if mappings := value.MapOfStrings(samlpConfig.GetAttr("mappings")); mappings != nil {
//				samlp["mappings"] = mappings
//			}
//			if nameIdentifierProbes := value.Strings(samlpConfig.GetAttr("name_identifier_probes")); nameIdentifierProbes != nil {
//				samlp["nameIdentifierProbes"] = nameIdentifierProbes
//			}
//			if logout := mapFromState(d.Get("addons.0.samlp.0.logout").(map[string]interface{})); len(logout) != 0 {
//				samlp["logout"] = logout
//			}
//
//			return stop
//		})
//
//		if len(samlp) > 0 {
//			addons["samlp"] = samlp
//		}
//
//		return stop
//	})
//
//	return addons
//}
//
// func mapFromState(input map[string]interface{}) map[string]interface{} {
//	output := make(map[string]interface{})
//
//	for key, val := range input {
//		switch v := val.(type) {
//		case string:
//			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
//				output[key] = i
//			} else if f, err := strconv.ParseFloat(v, 64); err == nil {
//				output[key] = f
//			} else if b, err := strconv.ParseBool(v); err == nil {
//				output[key] = b
//			} else {
//				output[key] = v
//			}
//		case map[string]interface{}:
//			output[key] = mapFromState(v)
//		case []interface{}:
//			output[key] = v
//		default:
//			output[key] = v
//		}
//	}
//
//	return output
// }........

func clientHasChange(c *management.Client) bool {
	return c.String() != "{}"
}
