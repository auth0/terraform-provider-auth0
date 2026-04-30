package resourceserver

import (
	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandResourceServer(data *schema.ResourceData) *management.ResourceServer {
	cfg := data.GetRawConfig()

	resourceServer := &management.ResourceServer{
		TokenLifetime: value.Int(cfg.GetAttr("token_lifetime")),
		SkipConsentForVerifiableFirstPartyClients: value.Bool(
			cfg.GetAttr("skip_consent_for_verifiable_first_party_clients"),
		),
	}

	if data.IsNewResource() {
		resourceServer.Identifier = value.String(cfg.GetAttr("identifier"))
	}

	// Allow updating SubjectTypeAuthorization for Auth0 Management API as well as non-management API.
	resourceServer.SubjectTypeAuthorization = expandSubjectTypeAuthorization(data)
	resourceServer.AuthorizationPolicy = expandAuthorizationPolicy(data)

	if !resourceServerIsAuth0ManagementAPI(data.GetRawState()) {
		if resourceServerIsAuth0MyAccountAPI(data.GetRawState()) {
			resourceServer.Name = auth0.String(auth0MyAccountAPIName)
		}
		resourceServer.SigningAlgorithm = value.String(cfg.GetAttr("signing_alg"))
		resourceServer.SigningSecret = value.String(cfg.GetAttr("signing_secret"))
		resourceServer.AllowOfflineAccess = value.Bool(cfg.GetAttr("allow_offline_access"))
		resourceServer.TokenLifetimeForWeb = value.Int(cfg.GetAttr("token_lifetime_for_web"))
		resourceServer.EnforcePolicies = value.Bool(cfg.GetAttr("enforce_policies"))
		resourceServer.TokenDialect = value.String(cfg.GetAttr("token_dialect"))
		resourceServer.VerificationLocation = value.String(cfg.GetAttr("verification_location"))
		resourceServer.AuthorizationDetails = expandAuthorizationDetails(data)
		resourceServer.TokenEncryption = expandTokenEncryption(data)
		resourceServer.ConsentPolicy = expandConsentPolicy(data)
		resourceServer.ProofOfPossession = expandProofOfPossession(data)
	}
	return resourceServer
}

func expandAuthorizationPolicy(data *schema.ResourceData) *management.ResourceServerAuthorizationPolicy {
	if !data.IsNewResource() && !data.HasChange("authorization_policy") {
		return nil
	}

	config := data.GetRawConfig().GetAttr("authorization_policy")
	if config.IsNull() || config.LengthInt() == 0 {
		return nil
	}

	var policy management.ResourceServerAuthorizationPolicy

	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		policy.PolicyID = value.String(cfg.GetAttr("policy_id"))
		return stop
	})

	if policy == (management.ResourceServerAuthorizationPolicy{}) {
		return nil
	}

	return &policy
}

func isAuthorizationPolicyNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("authorization_policy") {
		return false
	}
	return data.GetRawConfig().IsNull() ||
		data.GetRawConfig().GetAttr("authorization_policy").IsNull() ||
		data.GetRawConfig().GetAttr("authorization_policy").LengthInt() == 0
}

// fetchNullableFields returns a map of fields that need to be explicitly set
// to null on the resource server via a follow-up PATCH request, since the
// regular Update call uses `omitempty` and cannot transmit nil values.
func fetchNullableFields(data *schema.ResourceData) map[string]interface{} {
	type nullCheckFunc func(*schema.ResourceData) bool

	checks := map[string]nullCheckFunc{
		"consent_policy":        isConsentPolicyNull,
		"authorization_details": isAuthorizationDetailsNull,
		"token_encryption":      isTokenEncryptionNull,
		"proof_of_possession":   isProofOfPossessionNull,
		"authorization_policy":  isAuthorizationPolicyNull,
	}

	nullableMap := make(map[string]interface{})

	for field, checkFunc := range checks {
		if checkFunc(data) {
			nullableMap[field] = nil
		}
	}

	return nullableMap
}

func expandSubjectTypeAuthorization(data *schema.ResourceData) *management.ResourceServerSubjectTypeAuthorization {
	config := data.GetRawConfig().GetAttr("subject_type_authorization")
	if config.IsNull() {
		return nil
	}

	var sta management.ResourceServerSubjectTypeAuthorization

	isManagementAPI := resourceServerIsAuth0ManagementAPI(data.GetRawState())
	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		sta.User = expandSubjectTypeAuthorizationUser(cfg.GetAttr("user"))
		if !isManagementAPI {
			sta.Client = expandSubjectTypeAuthorizationClient(cfg.GetAttr("client"))
		} else if data.HasChange("subject_type_authorization.0.client") {
			// Changes to the client block in subject_type_authorization are not allowed for the management API.
			// This check prevents silently ignoring such errors.
			sta.Client = expandSubjectTypeAuthorizationClient(cfg.GetAttr("client"))
		}

		return stop
	})

	if sta == (management.ResourceServerSubjectTypeAuthorization{}) {
		return nil
	}

	return &sta
}

func expandSubjectTypeAuthorizationUser(userConfig cty.Value) *management.ResourceServerSubjectTypeAuthorizationUser {
	if userConfig.IsNull() {
		return nil
	}

	var user management.ResourceServerSubjectTypeAuthorizationUser

	userConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		user.Policy = value.String(cfg.GetAttr("policy"))
		return stop
	})

	if user == (management.ResourceServerSubjectTypeAuthorizationUser{}) {
		return nil
	}

	return &user
}

func expandSubjectTypeAuthorizationClient(clientConfig cty.Value) *management.ResourceServerSubjectTypeAuthorizationClient {
	if clientConfig.IsNull() {
		return nil
	}

	var client management.ResourceServerSubjectTypeAuthorizationClient

	clientConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		client.Policy = value.String(cfg.GetAttr("policy"))
		return stop
	})

	if client == (management.ResourceServerSubjectTypeAuthorizationClient{}) {
		return nil
	}

	return &client
}

func expandResourceServerScopes(scopes cty.Value) *[]management.ResourceServerScope {
	resourceServerScopes := make([]management.ResourceServerScope, 0)

	scopes.ForEachElement(func(_ cty.Value, scope cty.Value) (stop bool) {
		resourceServerScopes = append(resourceServerScopes, management.ResourceServerScope{
			Value:       value.String(scope.GetAttr("name")),
			Description: value.String(scope.GetAttr("description")),
		})

		return stop
	})

	return &resourceServerScopes
}

func isConsentPolicyNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("consent_policy") {
		return false
	}
	consentPolicy := value.String(data.GetRawConfig().GetAttr("consent_policy"))
	return consentPolicy != nil && *consentPolicy == "null"
}

func expandConsentPolicy(data *schema.ResourceData) *string {
	if !data.IsNewResource() && !data.HasChange("consent_policy") {
		return nil
	} else if isConsentPolicyNull(data) {
		return nil
	}

	return value.String(data.GetRawConfig().GetAttr("consent_policy"))
}

func isAuthorizationDetailsNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("authorization_details") {
		return false
	}
	empty := true

	config := data.GetRawConfig().GetAttr("authorization_details")
	if config.IsNull() || config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		disable := cfg.GetAttr("disable")
		if !disable.IsNull() && disable.True() {
			stop = true
		} else {
			empty = false
		}
		return stop
	}) {
		// We forced an early return because it was disabled.
		return true
	}

	return empty
}

func expandAuthorizationDetails(data *schema.ResourceData) *[]management.ResourceServerAuthorizationDetails {
	if !data.IsNewResource() && !data.HasChange("authorization_details") {
		return nil
	} else if isAuthorizationDetailsNull(data) {
		return nil
	}

	config := data.GetRawConfig().GetAttr("authorization_details")
	authorizationDetails := make([]management.ResourceServerAuthorizationDetails, 0, config.LengthInt())

	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		authorizationDetails = append(authorizationDetails, management.ResourceServerAuthorizationDetails{
			Type: value.String(cfg.GetAttr("type")),
		})

		return stop
	})

	if len(authorizationDetails) == 0 {
		return nil
	}

	return &authorizationDetails
}

func isTokenEncryptionNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("token_encryption") {
		return false
	}
	empty := true

	config := data.GetRawConfig().GetAttr("token_encryption")
	if config.IsNull() || config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		disable := cfg.GetAttr("disable")
		if !disable.IsNull() && disable.True() {
			stop = true
		} else {
			empty = false
		}
		return stop
	}) {
		// We forced an early return because it was disabled.
		return true
	}

	return empty
}

func expandTokenEncryption(data *schema.ResourceData) *management.ResourceServerTokenEncryption {
	if !data.IsNewResource() && !data.HasChange("token_encryption") {
		return nil
	} else if isTokenEncryptionNull(data) {
		return nil
	}

	var tokenEncryption management.ResourceServerTokenEncryption

	config := data.GetRawConfig().GetAttr("token_encryption")
	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		tokenEncryption.Format = value.String(cfg.GetAttr("format"))
		tokenEncryption.EncryptionKey = expandTokenEncryptionKey(cfg.GetAttr("encryption_key"))
		return stop
	})

	if tokenEncryption == (management.ResourceServerTokenEncryption{}) {
		return nil
	}

	return &tokenEncryption
}

func expandTokenEncryptionKey(config cty.Value) *management.ResourceServerTokenEncryptionKey {
	if config.IsNull() {
		return nil
	}

	var tokenEncryptionKey management.ResourceServerTokenEncryptionKey

	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		tokenEncryptionKey.Name = value.String(cfg.GetAttr("name"))
		tokenEncryptionKey.Alg = value.String(cfg.GetAttr("algorithm"))
		tokenEncryptionKey.Kid = value.String(cfg.GetAttr("kid"))
		tokenEncryptionKey.Pem = value.String(cfg.GetAttr("pem"))
		return stop
	})

	if tokenEncryptionKey == (management.ResourceServerTokenEncryptionKey{}) {
		return nil
	}

	return &tokenEncryptionKey
}

func isProofOfPossessionNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("proof_of_possession") {
		return false
	}
	empty := true

	config := data.GetRawConfig().GetAttr("proof_of_possession")
	if config.IsNull() || config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		disable := cfg.GetAttr("disable")
		if !disable.IsNull() && disable.True() {
			stop = true
		} else {
			empty = false
		}
		return stop
	}) {
		// We forced an early return because it was disabled.
		return true
	}

	return empty
}

func expandProofOfPossession(data *schema.ResourceData) *management.ResourceServerProofOfPossession {
	if !data.IsNewResource() && !data.HasChange("proof_of_possession") {
		return nil
	} else if isProofOfPossessionNull(data) {
		return nil
	}

	var proofOfPossession management.ResourceServerProofOfPossession

	config := data.GetRawConfig().GetAttr("proof_of_possession")
	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		proofOfPossession.Mechanism = value.String(cfg.GetAttr("mechanism"))
		proofOfPossession.Required = value.Bool(cfg.GetAttr("required"))
		proofOfPossession.RequiredFor = value.String(cfg.GetAttr("required_for"))
		return stop
	})

	if proofOfPossession == (management.ResourceServerProofOfPossession{}) {
		return nil
	}

	return &proofOfPossession
}

func resourceServerIsAuth0ManagementAPI(state cty.Value) bool {
	if state.IsNull() {
		return false
	}

	return state.GetAttr("name").AsString() == auth0ManagementAPIName
}

func resourceServerIsAuth0MyAccountAPI(state cty.Value) bool {
	if state.IsNull() {
		return false
	}

	return state.GetAttr("name").AsString() == auth0MyAccountAPIName
}
