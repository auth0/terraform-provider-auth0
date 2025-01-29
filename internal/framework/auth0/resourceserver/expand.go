package resourceserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/auth0/go-auth0/management"

	"github.com/auth0/terraform-provider-auth0/internal/framework/attr/value"
)

func expandResourceServer(ctx context.Context, _, planData, stateData resourceModel) *management.ResourceServer {
	// The config ang the plan should only differ in cases where
	// attributes have default values. These will be nil in the
	// config but will be set in the plan. The other difference
	// between the two is that optional computed attributes will
	// be nil in the config but unknown in the plan.
	resourceServer := &management.ResourceServer{
		TokenLifetime: value.Int(planData.TokenLifetime),
		SkipConsentForVerifiableFirstPartyClients: value.Bool(planData.SkipConsent),
	}

	// Only add the Identifier field if it is new.
	if stateData.Identifier.IsNull() {
		resourceServer.Identifier = value.String(planData.Identifier)
	}

	if planData.Name.ValueString() != auth0ManagementAPIName {
		if value.HasValue(planData.Name) {
			resourceServer.Name = value.String(planData.Name)
		}
		if value.HasValue(planData.SigningAlgorithm) {
			resourceServer.SigningAlgorithm = value.String(planData.SigningAlgorithm)
		}
		if value.HasValue(planData.SigningSecret) {
			resourceServer.SigningSecret = value.String(planData.SigningSecret)
		}
		if value.HasValue(planData.AllowOfflineAccess) {
			resourceServer.AllowOfflineAccess = value.Bool(planData.AllowOfflineAccess)
		}
		if value.HasValue(planData.TokenLifetimeForWeb) {
			resourceServer.TokenLifetimeForWeb = value.Int(planData.TokenLifetimeForWeb)
		}
		if value.HasValue(planData.EnforcePolicies) {
			resourceServer.EnforcePolicies = value.Bool(planData.EnforcePolicies)
		}
		if value.HasValue(planData.TokenDialect) {
			resourceServer.TokenDialect = value.String(planData.TokenDialect)
		}
		if value.HasValue(planData.VerificationLocation) {
			resourceServer.VerificationLocation = value.String(planData.VerificationLocation)
		}
		if value.HasValue(planData.AuthorizationDetails) {
			resourceServer.AuthorizationDetails = expandAuthorizationDetails(ctx, stateData.AuthorizationDetails, planData.AuthorizationDetails)
		}
		if value.HasValue(planData.TokenEncryption) {
			resourceServer.TokenEncryption = expandTokenEncryption(ctx, stateData.TokenEncryption, planData.TokenEncryption)
		}
		if value.HasValue(planData.ConsentPolicy) {
			resourceServer.ConsentPolicy = expandConsentPolicy(stateData.ConsentPolicy, planData.ConsentPolicy)
		}
		if value.HasValue(planData.ProofOfPossession) {
			resourceServer.ProofOfPossession = expandProofOfPossession(ctx, stateData.ProofOfPossession, planData.ProofOfPossession)
		}
	}
	return resourceServer
}

func expandResourceServerScopes(ctx context.Context, scopesSet types.Set) *[]management.ResourceServerScope {
	if !value.HasValue(scopesSet) {
		return nil
	}

	var scopes []scopesElementModel
	diagnostics := scopesSet.ElementsAs(ctx, &scopes, false)
	if diagnostics.HasError() {
		// This should never happen.
		return nil
	}
	resourceServerScopes := make([]management.ResourceServerScope, 0, len(scopes))

	for _, scope := range scopes {
		resourceServerScopes = append(resourceServerScopes, management.ResourceServerScope{
			Value:       value.String(scope.Name),
			Description: value.String(scope.Description),
		})
	}

	return &resourceServerScopes
}

func isConsentPolicyNull(before, after types.String) bool {
	if !value.HasChange(before, after) {
		return false
	}
	consentPolicy := value.String(after)
	// If it existed before, but doesn't now, remove it.
	return consentPolicy == nil || *consentPolicy == "null"
}

func expandConsentPolicy(before, after types.String) *string {
	if !value.HasChange(before, after) || isConsentPolicyNull(before, after) {
		return nil
	}

	return value.String(after)
}

func isAuthorizationDetailsNull(ctx context.Context, before, after types.List) bool {
	if !value.HasChange(before, after) {
		return false
	}

	if !value.HasValue(after) {
		// If it existed before, but doesn't now, remove it.
		return true
	}
	empty := true

	var details []authorizationDetailsModel
	diagnostics := after.ElementsAs(ctx, &details, false)
	if diagnostics.HasError() {
		// This should never happen.
		return false
	}

	for _, detail := range details {
		if value.HasValue(detail.Type) {
			empty = false
		}
	}

	return empty
}

func expandAuthorizationDetails(ctx context.Context, before, after types.List) *[]management.ResourceServerAuthorizationDetails {
	if !value.HasChange(before, after) || isAuthorizationDetailsNull(ctx, before, after) {
		return nil
	}

	authorizationDetails := make([]management.ResourceServerAuthorizationDetails, 0)
	if value.HasValue(after) {
		var details []authorizationDetailsModel
		diagnostics := after.ElementsAs(ctx, &details, false)
		if diagnostics.HasError() {
			// This should never happen.
			return nil
		}

		for _, detail := range details {
			authorizationDetails = append(authorizationDetails, management.ResourceServerAuthorizationDetails{
				Type: value.String(detail.Type),
			})
		}
	}

	if len(authorizationDetails) == 0 {
		return nil
	}

	return &authorizationDetails
}

func isTokenEncryptionNull(ctx context.Context, before, after types.Object) bool {
	if !value.HasChange(before, after) {
		return false
	}
	// If it existed before, but doesn't now, remove it.
	if !value.HasValue(after) {
		return true
	}

	var model tokenEncryptionModel
	diagnostics := after.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diagnostics.HasError() {
		return false
	}

	return !value.HasValue(model.Format)
}

func expandTokenEncryption(ctx context.Context, before, after types.Object) *management.ResourceServerTokenEncryption {
	if !value.HasChange(before, after) || isTokenEncryptionNull(ctx, before, after) {
		return nil
	}

	var model tokenEncryptionModel
	diagnostics := after.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diagnostics.HasError() {
		return nil
	}

	return &management.ResourceServerTokenEncryption{
		Format:        value.String(model.Format),
		EncryptionKey: expandTokenEncryptionKey(ctx, model.EncryptionKey),
	}
}

func expandTokenEncryptionKey(ctx context.Context, encryptionKeyObject types.Object) *management.ResourceServerTokenEncryptionKey {
	if !value.HasValue(encryptionKeyObject) {
		return nil
	}

	var model encryptionKeyModel
	diagnostics := encryptionKeyObject.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diagnostics.HasError() {
		return nil
	}

	return &management.ResourceServerTokenEncryptionKey{
		Name: value.String(model.Name),
		Alg:  value.String(model.Algorithm),
		Kid:  value.String(model.KID),
		Pem:  value.String(model.PEM),
	}
}

func isProofOfPossessionNull(ctx context.Context, before, after types.Object) bool {
	if !value.HasChange(before, after) {
		return false
	}
	if !value.HasValue(after) {
		return true
	}

	var model proofOfPossessionModel
	diagnostics := after.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diagnostics.HasError() {
		return false
	}

	return !value.HasValue(model.Mechanism)
}

func expandProofOfPossession(ctx context.Context, before, after types.Object) *management.ResourceServerProofOfPossession {
	if !value.HasChange(before, after) || isProofOfPossessionNull(ctx, before, after) {
		return nil
	}

	var model proofOfPossessionModel
	diagnostics := after.As(ctx, &model, basetypes.ObjectAsOptions{})
	if diagnostics.HasError() {
		return nil
	}

	return &management.ResourceServerProofOfPossession{
		Mechanism: value.String(model.Mechanism),
		Required:  value.Bool(model.Required),
	}
}
