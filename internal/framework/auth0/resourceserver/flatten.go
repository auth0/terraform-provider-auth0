package resourceserver

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/auth0/terraform-provider-auth0/internal/framework/attr/value"
)

func flattenResourceServerBase(ctx context.Context,
	state *tfsdk.State,
	resourceServer *management.ResourceServer,
	model *resourceModel,
) diag.Diagnostics {
	// We never get the PEM back from the server, so we need to find it
	// in the state.
	var pem attr.Value = value.AttrString(nil)
	if value.HasValue(model.TokenEncryption) {
		stateEncryptionKey := model.TokenEncryption.Attributes()["encryption_key"].(types.Object)
		if value.HasValue(stateEncryptionKey) {
			pem = stateEncryptionKey.Attributes()["pem"]
		}
	}
	tokenEncryption, diagnostics := flattenTokenEncryption(pem, resourceServer.GetTokenEncryption(), value.HasValue(model.TokenEncryption))
	diagnostics.Append(state.SetAttribute(ctx, path.Root("token_encryption"), tokenEncryption)...)

	authorizationDetails, d := flattenAuthorizationDetails(resourceServer.GetAuthorizationDetails(), value.HasValue(model.AuthorizationDetails))
	diagnostics.Append(d...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("authorization_details"), authorizationDetails)...)

	proofOfPossession, d := flattenProofOfPossession(resourceServer.GetProofOfPossession(), value.HasValue(model.ProofOfPossession))
	diagnostics.Append(d...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("proof_of_possession"), proofOfPossession)...)

	diagnostics.Append(state.SetAttribute(ctx, path.Root("name"), value.AttrString(resourceServer.Name))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("identifier"), value.AttrString(resourceServer.Identifier))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("token_lifetime"), value.AttrInt64(resourceServer.TokenLifetime))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("allow_offline_access"), value.AttrBool(resourceServer.AllowOfflineAccess))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("token_lifetime_for_web"), value.AttrInt64(resourceServer.TokenLifetimeForWeb))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("signing_alg"), value.AttrString(resourceServer.SigningAlgorithm))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("signing_secret"), value.AttrString(resourceServer.SigningSecret))...)
	diagnostics.Append(
		state.SetAttribute(ctx, path.Root("skip_consent_for_verifiable_first_party_clients"),
			value.AttrBool(resourceServer.SkipConsentForVerifiableFirstPartyClients))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("consent_policy"), flattenConsentPolicy(resourceServer.ConsentPolicy, value.HasValue(model.ProofOfPossession)))...)

	if resourceServer.GetName() != auth0ManagementAPIName {
		diagnostics.Append(state.SetAttribute(ctx, path.Root("verification_location"), value.AttrString(resourceServer.VerificationLocation))...)
		diagnostics.Append(state.SetAttribute(ctx, path.Root("enforce_policies"), value.AttrBool(resourceServer.EnforcePolicies))...)
		diagnostics.Append(state.SetAttribute(ctx, path.Root("token_dialect"), value.AttrString(resourceServer.TokenDialect))...)
	}

	return diagnostics
}

func flattenResourceServer(ctx context.Context,
	state *tfsdk.State,
	resourceServer *management.ResourceServer,
) diag.Diagnostics {
	var model resourceModel
	diagnostics := state.Get(ctx, &model)
	if diagnostics.HasError() {
		return diagnostics
	}
	diagnostics.Append(flattenResourceServerBase(ctx, state, resourceServer, &model)...)

	return diagnostics
}

func flattenResourceServerForDataSource(
	ctx context.Context,
	state *tfsdk.State,
	resourceServer *management.ResourceServer,
) diag.Diagnostics {
	var model dataSourceModel
	diagnostics := state.Get(ctx, &model)
	if diagnostics.HasError() {
		return diagnostics
	}
	diagnostics.Append(flattenResourceServerBase(ctx, state, resourceServer, &(model.resourceModel))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("verification_location"), value.AttrString(resourceServer.VerificationLocation))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("enforce_policies"), value.AttrBool(resourceServer.EnforcePolicies))...)
	diagnostics.Append(state.SetAttribute(ctx, path.Root("token_dialect"), value.AttrString(resourceServer.TokenDialect))...)

	scopes, diagnostics := flattenResourceServerScopesSet(resourceServer.Scopes)
	if !diagnostics.HasError() {
		diagnostics.Append(state.SetAttribute(ctx, path.Root("scopes"), scopes)...)
	}

	return diagnostics
}

func flattenResourceServerScopesSet(scopes *[]management.ResourceServerScope) (types.Set, diag.Diagnostics) {
	elements, diagnostics := flattenResourceServerScopes(scopes)

	if elements == nil {
		return basetypes.NewSetNull(scopesElementType), diagnostics
	}

	rval, d := basetypes.NewSetValue(scopesElementType, elements)
	diagnostics.Append(d...)

	return rval, diagnostics
}

func flattenResourceServerScopes(scopes *[]management.ResourceServerScope) ([]attr.Value, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if scopes == nil {
		return nil, nil
	}

	elements := make([]attr.Value, 0, len(*scopes))
	for _, scope := range *scopes {
		element, d := basetypes.NewObjectValue(
			scopesElementTypeMap,
			map[string]attr.Value{
				"name":        value.AttrString(scope.Value),
				"description": value.AttrString(scope.Description),
			},
		)
		diagnostics.Append(d...)
		elements = append(elements, element)
	}

	return elements, diagnostics
}

func flattenConsentPolicy(consentPolicy *string, hasStateValue bool) types.String {
	if consentPolicy == nil {
		if !hasStateValue {
			return basetypes.NewStringNull()
		}
		return basetypes.NewStringValue("null")
	}
	return value.AttrString(consentPolicy)
}

func flattenAuthorizationDetails(authorizationDetails []management.ResourceServerAuthorizationDetails, hasStateValue bool) (types.List, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	elements := make([]attr.Value, 0)
	if authorizationDetails == nil {
		if !hasStateValue {
			return basetypes.NewListNull(authorizationDetailsElementType), diag.Diagnostics{}
		}
	} else {
		for _, item := range authorizationDetails {
			element, d := basetypes.NewObjectValue(
				authorizationDetailsElementTypeMap,
				map[string]attr.Value{
					"type": value.AttrString(item.Type),
				},
			)
			diagnostics.Append(d...)
			elements = append(elements, element)
		}
	}
	rval, d := basetypes.NewListValue(authorizationDetailsElementType, elements)
	diagnostics.Append(d...)

	return rval, diagnostics
}

func flattenTokenEncryption(pem attr.Value, tokenEncryption *management.ResourceServerTokenEncryption, hasStateValue bool) (types.Object, diag.Diagnostics) {
	if tokenEncryption == nil {
		if !hasStateValue {
			return basetypes.NewObjectNull(tokenEncryptionTypeMap), diag.Diagnostics{}
		}
		return basetypes.NewObjectValue(
			tokenEncryptionTypeMap,
			map[string]attr.Value{
				"format":         basetypes.NewStringNull(),
				"encryption_key": basetypes.NewObjectNull(encryptionKeyTypeMap),
			},
		)
	}

	var diagnostics diag.Diagnostics
	var encryptionKey types.Object

	if tokenEncryption.EncryptionKey == nil {
		encryptionKey = basetypes.NewObjectNull(
			encryptionKeyTypeMap,
		)
	} else {
		if len(tokenEncryption.EncryptionKey.GetPem()) > 0 {
			pem = value.AttrString(tokenEncryption.EncryptionKey.Pem)
		} else if pem == nil {
			pem = basetypes.NewStringNull()
		}
		encryptionKey, diagnostics = basetypes.NewObjectValue(
			encryptionKeyTypeMap,
			map[string]attr.Value{
				"name":      value.AttrString(tokenEncryption.EncryptionKey.Name),
				"algorithm": value.AttrString(tokenEncryption.EncryptionKey.Alg),
				"kid":       value.AttrString(tokenEncryption.EncryptionKey.Kid),
				"pem":       pem,
			},
		)
	}
	rval, d := basetypes.NewObjectValue(
		tokenEncryptionTypeMap,
		map[string]attr.Value{
			"format":         value.AttrString(tokenEncryption.Format),
			"encryption_key": encryptionKey,
		},
	)
	diagnostics.Append(d...)

	return rval, diagnostics
}

func flattenProofOfPossession(proofOfPossession *management.ResourceServerProofOfPossession, hasStateValue bool) (types.Object, diag.Diagnostics) {
	if proofOfPossession == nil {
		if !hasStateValue {
			return basetypes.NewObjectNull(proofOfPossessionTypeMap), diag.Diagnostics{}
		}
		return basetypes.NewObjectValue(
			proofOfPossessionTypeMap,
			map[string]attr.Value{
				"mechanism": basetypes.NewStringNull(),
				"required":  basetypes.NewBoolNull(),
			},
		)
	}

	return basetypes.NewObjectValue(
		proofOfPossessionTypeMap,
		map[string]attr.Value{
			"mechanism": value.AttrString(proofOfPossession.Mechanism),
			"required":  value.AttrBool(proofOfPossession.Required),
		},
	)
}
