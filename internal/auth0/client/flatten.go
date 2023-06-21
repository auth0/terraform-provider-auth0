package client

import (
	"github.com/auth0/go-auth0/management"
)

func flattenCustomSocialConfiguration(customSocial *management.ClientNativeSocialLogin) []interface{} {
	if customSocial == nil {
		return nil
	}

	m := map[string]interface{}{
		"apple": []interface{}{
			map[string]interface{}{
				"enabled": customSocial.GetApple().GetEnabled(),
			},
		},
		"facebook": []interface{}{
			map[string]interface{}{
				"enabled": customSocial.GetFacebook().GetEnabled(),
			},
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

//	if addons == nil {
//		return nil
//	}
//
//	m := make(map[string]interface{})
//
//	if v, ok := addons["samlp"]; ok {
//		samlp := v.(map[string]interface{})
//
//		samlpMap := map[string]interface{}{
//			"issuer":                             samlp["issuer"],
//			"audience":                           samlp["audience"],
//			"recipient":                          samlp["recipient"],
//			"mappings":                           samlp["mappings"],
//			"create_upn_claim":                   samlp["createUpnClaim"],
//			"passthrough_claims_with_no_mapping": samlp["passthroughClaimsWithNoMapping"],
//			"map_unknown_claims_as_is":           samlp["mapUnknownClaimsAsIs"],
//			"map_identities":                     samlp["mapIdentities"],
//			"signature_algorithm":                samlp["signatureAlgorithm"],
//			"digest_algorithm":                   samlp["digestAlgorithm"],
//			"destination":                        samlp["destination"],
//			"lifetime_in_seconds":                samlp["lifetimeInSeconds"],
//			"sign_response":                      samlp["signResponse"],
//			"name_identifier_format":             samlp["nameIdentifierFormat"],
//			"name_identifier_probes":             samlp["nameIdentifierProbes"],
//			"authn_context_class_ref":            samlp["authnContextClassRef"],
//			"typed_attributes":                   samlp["typedAttributes"],
//			"include_attribute_name_format":      samlp["includeAttributeNameFormat"],
//			"binding":                            samlp["binding"],
//			"signing_cert":                       samlp["signingCert"],
//		}
//
//		if logout, ok := samlp["logout"].(map[string]interface{}); ok {
//			samlpMap["logout"] = mapToState(logout)
//		}
//
//		m["samlp"] = []interface{}{samlpMap}
//	}
//
//	for _, name := range []string{
//		"aws", "azure_blob", "azure_sb", "rms", "mscrm", "slack", "sentry",
//		"box", "cloudbees", "concur", "dropbox", "echosign", "egnyte",
//		"firebase", "newrelic", "office365", "salesforce", "salesforce_api",
//		"salesforce_sandbox_api", "layer", "sap_api", "sharepoint",
//		"springcm", "wams", "wsfed", "zendesk", "zoom",
//	} {
//		if v, ok := addons[name]; ok {
//			if addonType, ok := v.(map[string]interface{}); ok {
//				m[name] = mapToState(addonType)
//			}
//		}
//	}
//
//	return []interface{}{m}
// }.

func flattenClientMobile(mobile *management.ClientMobile) []interface{} {
	if mobile == nil {
		return nil
	}

	m := map[string]interface{}{
		"android": nil,
		"ios":     nil,
	}

	if mobile.GetAndroid() != nil {
		m["android"] = []interface{}{
			map[string]interface{}{
				"app_package_name":         mobile.GetAndroid().GetAppPackageName(),
				"sha256_cert_fingerprints": mobile.GetAndroid().GetKeyHashes(),
			},
		}
	}

	if mobile.GetIOS() != nil {
		m["ios"] = []interface{}{
			map[string]interface{}{
				"team_id":               mobile.GetIOS().GetTeamID(),
				"app_bundle_identifier": mobile.GetIOS().GetAppID(),
			},
		}
	}

	return []interface{}{m}
}

//	output := make(map[string]interface{})
//
//	for key, v := range input {
//		switch val := v.(type) {
//		case bool:
//			if val {
//				output[key] = "true"
//			} else {
//				output[key] = "false"
//			}
//		case float64:
//			output[key] = strconv.Itoa(int(val))
//		case int:
//			output[key] = strconv.Itoa(val)
//		default:
//			output[key] = val
//		}
//	}
//
//	return output
// }.
