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
