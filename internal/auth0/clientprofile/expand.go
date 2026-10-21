package clientprofile

import (
	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// The expand functions read the raw configuration instead of the resource data, because
// d.GetOk and d.Get cannot tell an explicit `0` or `false` apart from an unset attribute.

// envelopeValue returns the `default_value` of the single-item block with the given name.
func envelopeValue(object cty.Value, name string) cty.Value {
	return value.Attr(value.FirstBlock(value.Attr(object, name)), "default_value")
}

func expandClientProfileCreate(raw cty.Value) *management.CreateClientProfileRequestContent {
	request := &management.CreateClientProfileRequestContent{
		ClientConfiguration: expandClientConfiguration(raw, false),
		ApplyPolicy:         expandApplyPolicy(raw),
	}

	if name := value.StringIfSet(value.Attr(raw, "name")); name != nil {
		request.Name = *name
	}

	if description := value.StringIfSet(value.Attr(raw, "description")); description != nil {
		request.Description = &description
	}

	return request
}

// expandClientProfileUpdate always sends the complete desired state so that it converges
// regardless of what changed. The top level is merge-patch, while client_configuration is
// replaced as a whole by the API.
func expandClientProfileUpdate(raw cty.Value) *management.UpdateClientProfileRequestContent {
	request := &management.UpdateClientProfileRequestContent{}

	request.SetName(value.StringIfSet(value.Attr(raw, "name")))

	// A description removed from the configuration is sent as an explicit null, which clears it.
	description := value.StringIfSet(value.Attr(raw, "description"))
	request.SetDescription(&description)

	request.SetClientConfiguration(expandClientConfiguration(raw, true))
	request.SetApplyPolicy(expandApplyPolicy(raw))

	return request
}

func expandApplyPolicy(raw cty.Value) *management.ClientProfileApplyPolicy {
	policy := &management.ClientProfileApplyPolicy{
		ExternalMetadataType:      management.ClientProfileApplyPolicyExternalMetadataTypeEnumCimd,
		ExternalMetadataCreatedBy: management.ClientProfileApplyPolicyExternalMetadataCreatedByEnumClient,
	}

	block := value.FirstBlock(value.Attr(raw, "apply_policy"))
	if metadataType := value.StringIfSet(value.Attr(block, "external_metadata_type")); metadataType != nil {
		policy.ExternalMetadataType = management.ClientProfileApplyPolicyExternalMetadataTypeEnum(*metadataType)
	}
	if createdBy := value.StringIfSet(value.Attr(block, "external_metadata_created_by")); createdBy != nil {
		policy.ExternalMetadataCreatedBy = management.ClientProfileApplyPolicyExternalMetadataCreatedByEnum(*createdBy)
	}

	return policy
}

// expandClientConfiguration only builds the envelope of attributes that are configured, so unset
// attributes are omitted rather than sent as `{}`. On update a removed refresh_token block is sent
// as an empty object, which clears it.
func expandClientConfiguration(raw cty.Value, update bool) *management.ClientProfileClientConfiguration {
	configuration := &management.ClientProfileClientConfiguration{}
	raw = value.FirstBlock(value.Attr(raw, "client_configuration"))

	if block := value.FirstBlock(value.Attr(raw, "refresh_token")); value.IsSet(block) {
		refreshToken := &management.ClientProfileClientConfigurationRefreshToken{}
		if v := value.IntIfSet(envelopeValue(block, "token_lifetime")); v != nil {
			refreshToken.TokenLifetime = &management.ClientProfileRefreshTokenTokenLifetime{DefaultValue: v}
		}
		if v := value.IntIfSet(envelopeValue(block, "idle_token_lifetime")); v != nil {
			refreshToken.IdleTokenLifetime = &management.ClientProfileRefreshTokenIdleTokenLifetime{DefaultValue: v}
		}
		if v := value.IntIfSet(envelopeValue(block, "leeway")); v != nil {
			refreshToken.Leeway = &management.ClientProfileRefreshTokenLeeway{DefaultValue: v}
		}
		configuration.RefreshToken = refreshToken
	} else if update {
		configuration.RefreshToken = &management.ClientProfileClientConfigurationRefreshToken{}
	}

	if v := value.StringIfSet(envelopeValue(raw, "organization_usage")); v != nil {
		usage := management.ClientProfileClientConfigurationOrganizationUsageEnum(*v)
		configuration.OrganizationUsage = &management.ClientProfileClientConfigurationOrganizationUsage{DefaultValue: &usage}
	}

	if v := value.StringIfSet(envelopeValue(raw, "organization_require_behavior")); v != nil {
		behavior := management.ClientProfileClientConfigurationOrganizationRequireBehaviorEnum(*v)
		configuration.OrganizationRequireBehavior = &management.ClientProfileClientConfigurationOrganizationRequireBehavior{DefaultValue: &behavior}
	}

	if methods := value.SortedStrings(envelopeValue(raw, "organization_discovery_methods")); len(methods) > 0 {
		discovery := make([]management.ClientProfileClientConfigurationOrganizationDiscoveryMethodEnum, 0, len(methods))
		for _, method := range methods {
			discovery = append(discovery, management.ClientProfileClientConfigurationOrganizationDiscoveryMethodEnum(method))
		}
		configuration.OrganizationDiscoveryMethods = &management.ClientProfileClientConfigurationOrganizationDiscoveryMethods{DefaultValue: discovery}
	}

	if v := value.BoolIfSet(envelopeValue(raw, "require_proof_of_possession")); v != nil {
		configuration.RequireProofOfPossession = &management.ClientProfileClientConfigurationRequireProofOfPossession{DefaultValue: v}
	}

	return configuration
}
