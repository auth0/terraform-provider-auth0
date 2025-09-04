package resourceserver

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenResourceServer(data *schema.ResourceData, resourceServer *management.ResourceServer) error {
	result := multierror.Append(
		data.Set("name", resourceServer.GetName()),
		data.Set("identifier", resourceServer.GetIdentifier()),
		data.Set("token_lifetime", resourceServer.GetTokenLifetime()),
		data.Set("allow_offline_access", resourceServer.GetAllowOfflineAccess()),
		data.Set("token_lifetime_for_web", resourceServer.GetTokenLifetimeForWeb()),
		data.Set("signing_alg", resourceServer.GetSigningAlgorithm()),
		data.Set("signing_secret", resourceServer.GetSigningSecret()),
		data.Set(
			"skip_consent_for_verifiable_first_party_clients",
			resourceServer.GetSkipConsentForVerifiableFirstPartyClients(),
		),
		data.Set("consent_policy", flattenConsentPolicy(resourceServer.ConsentPolicy)),
		data.Set("authorization_details", flattenAuthorizationDetails(resourceServer.GetAuthorizationDetails())),
		data.Set("token_encryption", flattenTokenEncryption(data, resourceServer.GetTokenEncryption())),
		data.Set("proof_of_possession", flattenProofOfPossession(resourceServer.GetProofOfPossession())),
		data.Set("client_id", resourceServer.GetClientID()),
	)
	if resourceServer.GetName() != auth0ManagementAPIName {
		result = multierror.Append(
			result,
			data.Set("verification_location", resourceServer.GetVerificationLocation()),
			data.Set("enforce_policies", resourceServer.GetEnforcePolicies()),
			data.Set("token_dialect", resourceServer.GetTokenDialect()),
		)
	}

	return result.ErrorOrNil()
}

func flattenResourceServerForDataSource(data *schema.ResourceData, resourceServer *management.ResourceServer) error {
	result := multierror.Append(
		flattenResourceServer(data, resourceServer),
		data.Set("verification_location", resourceServer.GetVerificationLocation()),
		data.Set("enforce_policies", resourceServer.GetEnforcePolicies()),
		data.Set("token_dialect", resourceServer.GetTokenDialect()),
		data.Set("scopes", flattenResourceServerScopesSlice(resourceServer.GetScopes())),
	)

	return result.ErrorOrNil()
}

func flattenResourceServerScopes(data *schema.ResourceData, resourceServer *management.ResourceServer) error {
	result := multierror.Append(
		data.Set("resource_server_identifier", resourceServer.GetIdentifier()),
		data.Set("scopes", flattenResourceServerScopesSlice(resourceServer.GetScopes())),
	)

	return result.ErrorOrNil()
}

func flattenResourceServerScopesSlice(resourceServerScopes []management.ResourceServerScope) []map[string]interface{} {
	scopes := make([]map[string]interface{}, len(resourceServerScopes))

	for index, scope := range resourceServerScopes {
		scopes[index] = map[string]interface{}{
			"name":        scope.GetValue(),
			"description": scope.GetDescription(),
		}
	}

	return scopes
}

func flattenConsentPolicy(consentPolicy *string) string {
	if consentPolicy == nil {
		return "null"
	}
	return *consentPolicy
}

func flattenAuthorizationDetails(authorizationDetails []management.ResourceServerAuthorizationDetails) []map[string]interface{} {
	if authorizationDetails == nil {
		return []map[string]interface{}{
			{
				"disable": true,
			},
		}
	}
	results := make([]map[string]interface{}, len(authorizationDetails))

	for index, item := range authorizationDetails {
		results[index] = map[string]interface{}{
			"type": item.GetType(),
		}
	}

	return results
}

func flattenTokenEncryption(data *schema.ResourceData, tokenEncryption *management.ResourceServerTokenEncryption) []map[string]interface{} {
	if tokenEncryption == nil {
		return []map[string]interface{}{
			{
				"disable": true,
			},
		}
	}
	result := map[string]interface{}{
		"format": tokenEncryption.GetFormat(),
	}
	encryptionKey := tokenEncryption.GetEncryptionKey()
	if encryptionKey == nil {
		result["encryption_key"] = nil
	} else {
		result["encryption_key"] = []map[string]interface{}{
			{
				"name":      encryptionKey.GetName(),
				"algorithm": encryptionKey.GetAlg(),
				"kid":       encryptionKey.GetKid(),
				// This one doesn't get read back, so we have to get it from the state.
				"pem": data.Get("token_encryption.0.encryption_key.0.pem"),
			},
		}
	}

	return []map[string]interface{}{result}
}

func flattenProofOfPossession(proofOfPossession *management.ResourceServerProofOfPossession) []map[string]interface{} {
	if proofOfPossession == nil {
		return []map[string]interface{}{
			{
				"disable": true,
			},
		}
	}
	result := map[string]interface{}{
		"mechanism": proofOfPossession.GetMechanism(),
		"required":  proofOfPossession.GetRequired(),
	}

	return []map[string]interface{}{result}
}
