package client

import (
	mgmtv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenCIMDClient(data *schema.ResourceData, client *mgmtv2.GetClientResponseContent, validation *mgmtv2.CimdValidationResult) error {
	result := multierror.Append(
		data.Set("client_id", client.GetClientID()),
		data.Set("name", client.GetName()),
		data.Set("logo_uri", client.GetLogoURI()),
		data.Set("is_first_party", client.GetIsFirstParty()),
		data.Set("callbacks", client.GetCallbacks()),
		data.Set("external_client_id", client.GetExternalClientID()),
		data.Set("external_metadata_type", string(client.GetExternalMetadataType())),
		data.Set("external_metadata_created_by", string(client.GetExternalMetadataCreatedBy())),
		data.Set("jwks_uri", client.GetJwksURI()),
		data.Set("signing_keys", flattenCIMDSigningKeys(client.SigningKeys)),
		data.Set("description", client.GetDescription()),
		data.Set("app_type", string(client.GetAppType())),
		data.Set("allowed_origins", client.GetAllowedOrigins()),
		data.Set("web_origins", client.GetWebOrigins()),
		data.Set("grant_types", client.GetGrantTypes()),
		data.Set("oidc_conformant", client.GetOidcConformant()),
		data.Set("organization_discovery_methods", enumSliceToStrings(client.OrganizationDiscoveryMethods)),
		data.Set("require_proof_of_possession", client.GetRequireProofOfPossession()),
		data.Set("skip_non_verifiable_callback_uri_confirmation_prompt", client.GetSkipNonVerifiableCallbackURIConfirmationPrompt()),
		data.Set("third_party_security_mode", string(client.GetThirdPartySecurityMode())),
		data.Set("redirection_policy", string(client.GetRedirectionPolicy())),
		data.Set("client_metadata", flattenCIMDClientMetadata(client.ClientMetadata)),
		data.Set("jwt_configuration", flattenCIMDJwtConfiguration(client.JwtConfiguration)),
		data.Set("refresh_token", flattenCIMDRefreshToken(client.RefreshToken)),
		data.Set("default_organization", flattenCIMDDefaultOrganization(client.DefaultOrganization)),
		data.Set("token_quota", flattenCIMDTokenQuota(client.TokenQuota)),
		data.Set("validation", flattenCIMDValidation(validation)),
		data.Set("external_client_id_version", data.Get("external_client_id_version")),
	)

	return result.ErrorOrNil()
}

func flattenCIMDJwtConfiguration(jwt *mgmtv2.ClientJwtConfiguration) []interface{} {
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

func flattenCIMDRefreshToken(rt *mgmtv2.ClientRefreshTokenConfiguration) []interface{} {
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

func flattenCIMDDefaultOrganization(do *mgmtv2.ClientDefaultOrganization) []interface{} {
	if do == nil {
		return nil
	}

	m := map[string]interface{}{
		"organization_id": do.GetOrganizationID(),
		"flows":           enumSliceToStrings(do.GetFlows()),
	}

	return []interface{}{m}
}

func flattenCIMDTokenQuota(tq *mgmtv2.TokenQuota) []interface{} {
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

func flattenCIMDClientMetadata(cm *mgmtv2.ClientMetadata) map[string]interface{} {
	if cm == nil {
		return nil
	}
	return *cm
}

func flattenCIMDSigningKeys(keys *mgmtv2.ClientSigningKeys) []interface{} {
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

func flattenCIMDValidation(v *mgmtv2.CimdValidationResult) []interface{} {
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

func enumSliceToStrings[T ~string](s []T) []string {
	if s == nil {
		return nil
	}
	result := make([]string, len(s))
	for i, v := range s {
		result[i] = string(v)
	}
	return result
}
