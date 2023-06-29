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
		"aws":        nil,
		"azure_blob": nil,
		"azure_sb":   nil,
		"rms":        nil,
		"mscrm":      nil,
		"slack":      nil,
		"sentry":     nil,
		"echosign":   nil,
		"egnyte":     nil,
		"firebase":   nil,
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

	return []interface{}{m}
}
