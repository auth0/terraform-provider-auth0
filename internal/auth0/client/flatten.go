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

func flattenClientAddons(addons *management.ClientAddons) []interface{} {
	if addons == nil {
		return nil
	}

	m := map[string]interface{}{
		"aws":                    nil,
		"azure_blob":             nil,
		"azure_sb":               nil,
		"rms":                    nil,
		"mscrm":                  nil,
		"slack":                  nil,
		"sentry":                 nil,
		"echosign":               nil,
		"egnyte":                 nil,
		"firebase":               nil,
		"office365":              nil,
		"salesforce":             nil,
		"salesforce_api":         nil,
		"salesforce_sandbox_api": nil,
		"layer":                  nil,
		"sap_api":                nil,
		"sharepoint":             nil,
		"springcm":               nil,
		"wams":                   nil,
		"zendesk":                nil,
		"zoom":                   nil,
		"sso_integration":        nil,
		"samlp":                  nil,
	}

	if addons.GetAWS() != nil {
		m["aws"] = []interface{}{
			map[string]interface{}{
				"principal":           addons.GetAWS().GetPrincipal(),
				"role":                addons.GetAWS().GetRole(),
				"lifetime_in_seconds": addons.GetAWS().GetLifetimeInSeconds(),
			},
		}
	}

	if addons.GetAzureBlob() != nil {
		m["azure_blob"] = []interface{}{
			map[string]interface{}{
				"account_name":       addons.GetAzureBlob().GetAccountName(),
				"storage_access_key": addons.GetAzureBlob().GetStorageAccessKey(),
				"container_name":     addons.GetAzureBlob().GetContainerName(),
				"blob_name":          addons.GetAzureBlob().GetBlobName(),
				"expiration":         addons.GetAzureBlob().GetExpiration(),
				"signed_identifier":  addons.GetAzureBlob().GetSignedIdentifier(),
				"blob_read":          addons.GetAzureBlob().GetBlobRead(),
				"blob_write":         addons.GetAzureBlob().GetBlobWrite(),
				"blob_delete":        addons.GetAzureBlob().GetBlobDelete(),
				"container_read":     addons.GetAzureBlob().GetContainerRead(),
				"container_write":    addons.GetAzureBlob().GetContainerWrite(),
				"container_delete":   addons.GetAzureBlob().GetContainerDelete(),
				"container_list":     addons.GetAzureBlob().GetContainerList(),
			},
		}
	}

	if addons.GetAzureSB() != nil {
		m["azure_sb"] = []interface{}{
			map[string]interface{}{
				"namespace":    addons.GetAzureSB().GetNamespace(),
				"sas_key_name": addons.GetAzureSB().GetSASKeyName(),
				"sas_key":      addons.GetAzureSB().GetSASKey(),
				"entity_path":  addons.GetAzureSB().GetEntityPath(),
				"expiration":   addons.GetAzureSB().GetExpiration(),
			},
		}
	}

	if addons.GetRMS() != nil {
		m["rms"] = []interface{}{
			map[string]interface{}{
				"url": addons.GetRMS().GetURL(),
			},
		}
	}

	if addons.GetMSCRM() != nil {
		m["mscrm"] = []interface{}{
			map[string]interface{}{
				"url": addons.GetMSCRM().GetURL(),
			},
		}
	}

	if addons.GetSlack() != nil {
		m["slack"] = []interface{}{
			map[string]interface{}{
				"team": addons.GetSlack().GetTeam(),
			},
		}
	}

	if addons.GetSentry() != nil {
		m["sentry"] = []interface{}{
			map[string]interface{}{
				"org_slug": addons.GetSentry().GetOrgSlug(),
				"base_url": addons.GetSentry().GetBaseURL(),
			},
		}
	}

	if addons.GetEchoSign() != nil {
		m["echosign"] = []interface{}{
			map[string]interface{}{
				"domain": addons.GetEchoSign().GetDomain(),
			},
		}
	}

	if addons.GetEgnyte() != nil {
		m["egnyte"] = []interface{}{
			map[string]interface{}{
				"domain": addons.GetEgnyte().GetDomain(),
			},
		}
	}

	if addons.GetFirebase() != nil {
		m["firebase"] = []interface{}{
			map[string]interface{}{
				"secret":              addons.GetFirebase().GetSecret(),
				"private_key_id":      addons.GetFirebase().GetPrivateKeyID(),
				"private_key":         addons.GetFirebase().GetPrivateKey(),
				"client_email":        addons.GetFirebase().GetClientEmail(),
				"lifetime_in_seconds": addons.GetFirebase().GetLifetimeInSeconds(),
			},
		}
	}

	if addons.GetNewRelic() != nil {
		m["newrelic"] = []interface{}{
			map[string]interface{}{
				"account": addons.GetNewRelic().GetAccount(),
			},
		}
	}

	if addons.GetOffice365() != nil {
		m["office365"] = []interface{}{
			map[string]interface{}{
				"domain":     addons.GetOffice365().GetDomain(),
				"connection": addons.GetOffice365().GetConnection(),
			},
		}
	}

	if addons.GetSalesforce() != nil {
		m["salesforce"] = []interface{}{
			map[string]interface{}{
				"entity_id": addons.GetSalesforce().GetEntityID(),
			},
		}
	}

	if addons.GetSalesforceAPI() != nil {
		m["salesforce_api"] = []interface{}{
			map[string]interface{}{
				"client_id":             addons.GetSalesforceAPI().GetClientID(),
				"principal":             addons.GetSalesforceAPI().GetPrincipal(),
				"community_name":        addons.GetSalesforceAPI().GetCommunityName(),
				"community_url_section": addons.GetSalesforceAPI().GetCommunityURLSection(),
			},
		}
	}

	if addons.GetSalesforceSandboxAPI() != nil {
		m["salesforce_sandbox_api"] = []interface{}{
			map[string]interface{}{
				"client_id":             addons.GetSalesforceSandboxAPI().GetClientID(),
				"principal":             addons.GetSalesforceSandboxAPI().GetPrincipal(),
				"community_name":        addons.GetSalesforceSandboxAPI().GetCommunityName(),
				"community_url_section": addons.GetSalesforceSandboxAPI().GetCommunityURLSection(),
			},
		}
	}

	if addons.GetLayer() != nil {
		m["layer"] = []interface{}{
			map[string]interface{}{
				"provider_id": addons.GetLayer().GetProviderID(),
				"key_id":      addons.GetLayer().GetKeyID(),
				"private_key": addons.GetLayer().GetPrivateKey(),
				"principal":   addons.GetLayer().GetPrincipal(),
				"expiration":  addons.GetLayer().GetExpiration(),
			},
		}
	}

	if addons.GetSAPAPI() != nil {
		m["sap_api"] = []interface{}{
			map[string]interface{}{
				"client_id":              addons.GetSAPAPI().GetClientID(),
				"username_attribute":     addons.GetSAPAPI().GetUsernameAttribute(),
				"token_endpoint_url":     addons.GetSAPAPI().GetTokenEndpointURL(),
				"scope":                  addons.GetSAPAPI().GetScope(),
				"service_password":       addons.GetSAPAPI().GetServicePassword(),
				"name_identifier_format": addons.GetSAPAPI().GetNameIdentifierFormat(),
			},
		}
	}

	if addons.GetSharePoint() != nil {
		m["sharepoint"] = []interface{}{
			map[string]interface{}{
				"url":          addons.GetSharePoint().GetURL(),
				"external_url": addons.GetSharePoint().GetExternalURL(),
			},
		}
	}

	if addons.GetSpringCM() != nil {
		m["springcm"] = []interface{}{
			map[string]interface{}{
				"acs_url": addons.GetSpringCM().GetACSURL(),
			},
		}
	}

	if addons.GetWAMS() != nil {
		m["wams"] = []interface{}{
			map[string]interface{}{
				"master_key": addons.GetWAMS().GetMasterkey(),
			},
		}
	}

	if addons.GetZendesk() != nil {
		m["zendesk"] = []interface{}{
			map[string]interface{}{
				"account_name": addons.GetZendesk().GetAccountName(),
			},
		}
	}

	if addons.GetZoom() != nil {
		m["zoom"] = []interface{}{
			map[string]interface{}{
				"account": addons.GetZoom().GetAccount(),
			},
		}
	}

	if addons.GetSSOIntegration() != nil {
		m["sso_integration"] = []interface{}{
			map[string]interface{}{
				"name":    addons.GetSSOIntegration().GetName(),
				"version": addons.GetSSOIntegration().GetVersion(),
			},
		}
	}

	if addons.GetSAML2() != nil && addons.GetSAML2().String() != "{}" {
		var logout interface{}

		if addons.GetSAML2().GetLogout() != nil {
			logout = []interface{}{
				map[string]interface{}{
					"callback":    addons.GetSAML2().GetLogout().GetCallback(),
					"slo_enabled": addons.GetSAML2().GetLogout().GetSLOEnabled(),
				},
			}
		}

		m["samlp"] = []interface{}{
			map[string]interface{}{
				"mappings":                           addons.GetSAML2().GetMappings(),
				"audience":                           addons.GetSAML2().GetAudience(),
				"recipient":                          addons.GetSAML2().GetRecipient(),
				"create_upn_claim":                   addons.GetSAML2().GetCreateUPNClaim(),
				"map_unknown_claims_as_is":           addons.GetSAML2().GetMapUnknownClaimsAsIs(),
				"passthrough_claims_with_no_mapping": addons.GetSAML2().GetPassthroughClaimsWithNoMapping(),
				"map_identities":                     addons.GetSAML2().GetMapIdentities(),
				"signature_algorithm":                addons.GetSAML2().GetSignatureAlgorithm(),
				"digest_algorithm":                   addons.GetSAML2().GetDigestAlgorithm(),
				"issuer":                             addons.GetSAML2().GetIssuer(),
				"destination":                        addons.GetSAML2().GetDestination(),
				"lifetime_in_seconds":                addons.GetSAML2().GetLifetimeInSeconds(),
				"sign_response":                      addons.GetSAML2().GetSignResponse(),
				"name_identifier_format":             addons.GetSAML2().GetNameIdentifierFormat(),
				"name_identifier_probes":             addons.GetSAML2().GetNameIdentifierProbes(),
				"authn_context_class_ref":            addons.GetSAML2().GetAuthnContextClassRef(),
				"typed_attributes":                   addons.GetSAML2().GetTypedAttributes(),
				"include_attribute_name_format":      addons.GetSAML2().GetIncludeAttributeNameFormat(),
				"binding":                            addons.GetSAML2().GetBinding(),
				"signing_cert":                       addons.GetSAML2().GetSigningCert(),
				"logout":                             logout,
			},
		}
	}

	return []interface{}{m}
}
