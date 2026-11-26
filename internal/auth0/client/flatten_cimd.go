package client

import (
	mgmtv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

func flattenCIMDClient(data *schema.ResourceData, client *mgmtv3.GetClientResponseContent, validation *mgmtv3.CimdValidationResult) error {
	result := multierror.Append(
		data.Set("client_id", client.GetClientID()),
		data.Set("name", client.GetName()),
		data.Set("logo_uri", client.GetLogoURI()),
		data.Set("is_first_party", client.GetIsFirstParty()),
		data.Set("callbacks", client.GetCallbacks()),
		data.Set("external_client_id", client.GetExternalClientID()),
		data.Set("external_metadata_type", string(client.GetExternalMetadataType())),
		data.Set("external_metadata_created_by", string(client.GetExternalMetadataCreatedBy())),
		data.Set("external_metadata_status", string(client.GetExternalMetadataStatus())),
		data.Set("jwks_uri", client.GetJwksURI()),
		data.Set("signing_keys", flattenCIMDSigningKeys(client.SigningKeys)),
		data.Set("description", client.GetDescription()),
		data.Set("agent_id", client.GetAgentID()),
		data.Set("app_type", string(client.GetAppType())),
		data.Set("allowed_origins", client.GetAllowedOrigins()),
		data.Set("web_origins", client.GetWebOrigins()),
		data.Set("grant_types", client.GetGrantTypes()),
		data.Set("oidc_conformant", client.GetOidcConformant()),
		data.Set("organization_usage", string(client.GetOrganizationUsage())),
		data.Set("organization_require_behavior", string(client.GetOrganizationRequireBehavior())),
		data.Set("organization_discovery_methods", internalSchema.ToStrSlice(client.OrganizationDiscoveryMethods...)),
		data.Set("require_proof_of_possession", client.GetRequireProofOfPossession()),
		data.Set("skip_non_verifiable_callback_uri_confirmation_prompt", client.GetSkipNonVerifiableCallbackURIConfirmationPrompt()),
		data.Set("third_party_security_mode", string(client.GetThirdPartySecurityMode())),
		data.Set("redirection_policy", string(client.GetRedirectionPolicy())),
		data.Set("client_metadata", flattenCIMDClientMetadata(client.ClientMetadata)),
		data.Set("jwt_configuration", flattenCIMDJwtConfiguration(client.JwtConfiguration)),
		data.Set("refresh_token", flattenCIMDRefreshToken(client.RefreshToken)),
		data.Set("default_organization", flattenCIMDDefaultOrganization(client.DefaultOrganization)),
		data.Set("token_quota", flattenCIMDTokenQuota(client.TokenQuota)),
		data.Set("fedcm_login", flattenCIMDFedCMLogin(client.FedcmLogin)),
		data.Set("identity_assertion_authorization_grant", flattenCIMDIdentityAssertionAuthorizationGrant(client.IdentityAssertionAuthorizationGrant)),
		data.Set("oidc_support", flattenCIMDOIDCSupport(client.OidcSupport)),
		data.Set("validation", flattenCIMDValidation(validation)),
		data.Set("external_client_id_version", data.Get("external_client_id_version")),
	)

	return result.ErrorOrNil()
}

func flattenCIMDJwtConfiguration(jwt *mgmtv3.ClientJwtConfiguration) []interface{} {
	if jwt == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"lifetime_in_seconds": jwt.GetLifetimeInSeconds(),
			"alg":                 string(jwt.GetAlg()),
			"secret_encoded":      jwt.GetSecretEncoded(),
		},
	}
}

func flattenCIMDRefreshToken(rt *mgmtv3.ClientRefreshTokenConfiguration) []interface{} {
	if rt == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"rotation_type":                string(rt.GetRotationType()),
			"expiration_type":              string(rt.GetExpirationType()),
			"leeway":                       rt.GetLeeway(),
			"token_lifetime":               rt.GetTokenLifetime(),
			"infinite_token_lifetime":      rt.GetInfiniteTokenLifetime(),
			"idle_token_lifetime":          rt.GetIdleTokenLifetime(),
			"infinite_idle_token_lifetime": rt.GetInfiniteIdleTokenLifetime(),
		},
	}
}

func flattenCIMDDefaultOrganization(do *mgmtv3.ClientDefaultOrganization) []interface{} {
	if do == nil {
		return nil
	}

	m := map[string]interface{}{
		"organization_id": do.GetOrganizationID(),
		"flows":           internalSchema.ToStrSlice(do.GetFlows()...),
	}

	return []interface{}{m}
}

func flattenCIMDTokenQuota(tq *mgmtv3.TokenQuota) []interface{} {
	if tq == nil || tq.GetClientCredentials() == nil {
		return nil
	}

	cc := tq.GetClientCredentials()
	clientCreds := map[string]interface{}{
		"enforce":  cc.GetEnforce(),
		"per_hour": cc.GetPerHour(),
		"per_day":  cc.GetPerDay(),
	}

	return []interface{}{
		map[string]interface{}{
			"client_credentials": []interface{}{clientCreds},
		},
	}
}

func flattenCIMDFedCMLogin(fedcmLogin *mgmtv3.FedCmLogin) []interface{} {
	if fedcmLogin == nil {
		return nil
	}

	var google []interface{}
	if fedcmLogin.Google != nil {
		google = []interface{}{
			map[string]interface{}{
				"is_enabled": fedcmLogin.Google.GetIsEnabled(),
			},
		}
	}

	return []interface{}{
		map[string]interface{}{
			"google": google,
		},
	}
}

func flattenCIMDIdentityAssertionAuthorizationGrant(grant *mgmtv3.IdentityAssertionAuthorizationGrant) []interface{} {
	if grant == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"active": grant.GetActive(),
		},
	}
}

// flattenCIMDOIDCSupport reads raw pointers instead of getters, because the API only returns the object
// for OIDC conformant clients. A nil object is flattened to an empty block.
func flattenCIMDOIDCSupport(oidcSupport *mgmtv3.ClientOidcSupport) []interface{} {
	if oidcSupport == nil || oidcSupport.IsAllowed == nil {
		return nil
	}

	flattened := map[string]interface{}{
		"is_allowed": *oidcSupport.IsAllowed,
	}
	if oidcSupport.AllowAllScopes != nil && *oidcSupport.AllowAllScopes {
		flattened["allow_all_scopes"] = true
	}
	if len(oidcSupport.AllowedScopes) > 0 {
		scopes := make([]interface{}, 0, len(oidcSupport.AllowedScopes))
		for _, scope := range oidcSupport.AllowedScopes {
			scopes = append(scopes, string(scope))
		}
		flattened["allowed_scopes"] = scopes
	}

	return []interface{}{flattened}
}

func flattenCIMDClientMetadata(cm *mgmtv3.ClientMetadata) map[string]interface{} {
	if cm == nil {
		return nil
	}
	return *cm
}

func flattenCIMDSigningKeys(keys *mgmtv3.ClientSigningKeys) []interface{} {
	if keys == nil {
		return nil
	}

	result := make([]interface{}, 0, len(*keys))
	for _, key := range *keys {
		if key == nil {
			continue
		}
		result = append(result, map[string]interface{}{
			"pkcs7":   key.GetPkcs7(),
			"cert":    key.GetCert(),
			"subject": key.GetSubject(),
		})
	}

	return result
}

func flattenCIMDValidation(v *mgmtv3.CimdValidationResult) []interface{} {
	if v == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"valid":      v.GetValid(),
			"violations": v.GetViolations(),
			"warnings":   v.GetWarnings(),
		},
	}
}
