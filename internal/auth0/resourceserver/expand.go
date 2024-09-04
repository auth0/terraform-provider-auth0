package resourceserver

import (
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

	if !resourceServerIsAuth0ManagementAPI(data.GetRawState()) {
		resourceServer.Name = value.String(cfg.GetAttr("name"))
		resourceServer.SigningAlgorithm = value.String(cfg.GetAttr("signing_alg"))
		resourceServer.SigningSecret = value.String(cfg.GetAttr("signing_secret"))
		resourceServer.AllowOfflineAccess = value.Bool(cfg.GetAttr("allow_offline_access"))
		resourceServer.TokenLifetimeForWeb = value.Int(cfg.GetAttr("token_lifetime_for_web"))
		resourceServer.EnforcePolicies = value.Bool(cfg.GetAttr("enforce_policies"))
		resourceServer.TokenDialect = value.String(cfg.GetAttr("token_dialect"))
		resourceServer.VerificationLocation = value.String(cfg.GetAttr("verification_location"))
		resourceServer.AuthorizationDetails = expandAuthorizationDetails(cfg.GetAttr("authorization_details"))
		resourceServer.ConsentPolicy = expandConsentPolicy(cfg.GetAttr("consent_policy"))
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

func isConsentPolicyNull(config cty.Value) bool {
	consentPolicy := value.String(config.GetAttr("consent_policy"))
	return consentPolicy != nil && *consentPolicy == "null"
}

func expandConsentPolicy(config cty.Value) *string {
	consentPolicy := value.String(config)
	if consentPolicy == nil || *consentPolicy == "null" {
		return nil
	}

	return nil
}

func isAuthorizationDetailsNull(config cty.Value) bool {
	empty := true

	detailsConfig := config.GetAttr("authorization_details")
	if detailsConfig.IsNull() || detailsConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
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

func expandAuthorizationDetails(config cty.Value) *[]management.ResourceServerAuthorizationDetails {
	if config.IsNull() {
		return nil
	}

	authorizationDetails := make([]management.ResourceServerAuthorizationDetails, 0, config.LengthInt())
	if config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		disable := cfg.GetAttr("disable")
		if !disable.IsNull() && disable.True() {
			// Force it to exit the ForEachElement and return nil.
			stop = true
		} else {
			authorizationDetails = append(authorizationDetails, management.ResourceServerAuthorizationDetails{
				Type: value.String(cfg.GetAttr("type")),
			})
		}

		return stop
	}) {
		// We forced an early return because it was disabled.
		return nil
	}

	if len(authorizationDetails) == 0 {
		return nil
	}

	return &authorizationDetails
}

func resourceServerIsAuth0ManagementAPI(state cty.Value) bool {
	if state.IsNull() {
		return false
	}

	return state.GetAttr("name").AsString() == auth0ManagementAPIName
}
