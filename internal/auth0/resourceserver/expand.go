package resourceserver

import (
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
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

	if !resourceServerIsAuth0ManagementAPI(data.GetRawState()) {
		resourceServer.Name = value.String(cfg.GetAttr("name"))
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

func validateResourceServer(data *schema.ResourceData) error {
	var result *multierror.Error

	authorizationDetailsConfig := data.GetRawConfig().GetAttr("authorization_details")
	if !authorizationDetailsConfig.IsNull() {
		disable := false
		found := false

		authorizationDetailsConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
			if !cfg.GetAttr("disable").IsNull() && cfg.GetAttr("disable").True() {
				disable = true
			}
			if !cfg.GetAttr("type").IsNull() {
				found = true
			}
			return stop
		})
		if disable && found {
			result = multierror.Append(
				result,
				fmt.Errorf("only one of disable and type should be set in the authorization_details block"),
			)
		}
	}

	tokenEncryptionConfig := data.GetRawConfig().GetAttr("token_encryption")
	if !tokenEncryptionConfig.IsNull() {
		disable := false
		found := false

		tokenEncryptionConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
			if !cfg.GetAttr("disable").IsNull() && cfg.GetAttr("disable").True() {
				disable = true
			}
			if !cfg.GetAttr("format").IsNull() {
				found = true
			}
			if !cfg.GetAttr("encryption_key").IsNull() && cfg.GetAttr("encryption_key").LengthInt() > 0 {
				found = true
			}
			return stop
		})
		if disable && found {
			result = multierror.Append(
				result,
				fmt.Errorf("only one of disable and format or encryption_key should be set in the token_encryption blocks"),
			)
		}
	}

	proofOfPossessionConfig := data.GetRawConfig().GetAttr("proof_of_possession")
	if !proofOfPossessionConfig.IsNull() {
		disable := false
		found := false

		proofOfPossessionConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
			if !cfg.GetAttr("disable").IsNull() && cfg.GetAttr("disable").True() {
				disable = true
			}
			if !cfg.GetAttr("mechanism").IsNull() {
				found = true
			}
			if !cfg.GetAttr("required").IsNull() {
				found = true
			}
			return stop
		})
		if disable && found {
			result = multierror.Append(
				result,
				fmt.Errorf("only one of disable and mechanism or required should be set in the proof_of_possession block"),
			)
		}
	}

	return result.ErrorOrNil()
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
