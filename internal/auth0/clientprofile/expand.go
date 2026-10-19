package clientprofile

import (
	"sort"

	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// The expand functions read the raw configuration instead of the resource data, because
// d.GetOk and d.Get cannot tell an explicit `0` or `false` apart from an unset attribute.

// attr returns the named attribute of an object value, a null value when it is absent, or an
// unknown value when the object itself is not yet known.
func attr(object cty.Value, name string) cty.Value {
	if !object.IsKnown() {
		return cty.UnknownVal(cty.DynamicPseudoType)
	}
	if object.IsNull() || !object.Type().IsObjectType() || !object.Type().HasAttribute(name) {
		return cty.NullVal(cty.DynamicPseudoType)
	}

	return object.GetAttr(name)
}

// isSet reports whether a configuration value is both known and non-null.
func isSet(v cty.Value) bool {
	return v.IsKnown() && !v.IsNull()
}

// firstBlock returns the first element of a nested block list, a null value when the block is absent,
// or an unknown value when the list is not yet known.
func firstBlock(list cty.Value) cty.Value {
	if !list.IsKnown() {
		return cty.UnknownVal(cty.DynamicPseudoType)
	}
	if list.IsNull() || list.LengthInt() == 0 {
		return cty.NullVal(cty.DynamicPseudoType)
	}

	return list.AsValueSlice()[0]
}

// envelopeValue returns the `default_value` of the single-item block with the given name.
func envelopeValue(object cty.Value, name string) cty.Value {
	return attr(firstBlock(attr(object, name)), "default_value")
}

func stringVal(v cty.Value) *string {
	if isSet(v) {
		return value.String(v)
	}

	return nil
}

func intVal(v cty.Value) *int {
	if isSet(v) {
		return value.Int(v)
	}

	return nil
}

func boolVal(v cty.Value) *bool {
	if isSet(v) {
		return value.Bool(v)
	}

	return nil
}

func stringsVal(v cty.Value) []string {
	if !isSet(v) {
		return nil
	}

	var result []string
	for _, item := range v.AsValueSlice() {
		if isSet(item) {
			result = append(result, item.AsString())
		}
	}
	sort.Strings(result)

	return result
}

func expandClientProfileCreate(raw cty.Value) *management.CreateClientProfileRequestContent {
	request := &management.CreateClientProfileRequestContent{
		ClientConfiguration: expandClientConfiguration(raw, false),
		ApplyPolicy:         expandApplyPolicy(raw),
	}

	if name := stringVal(attr(raw, "name")); name != nil {
		request.Name = *name
	}

	if description := stringVal(attr(raw, "description")); description != nil {
		request.Description = &description
	}

	return request
}

// expandClientProfileUpdate always sends the complete desired state so that it converges
// regardless of what changed. The top level is merge-patch, while client_configuration is
// replaced as a whole by the API.
func expandClientProfileUpdate(raw cty.Value) *management.UpdateClientProfileRequestContent {
	request := &management.UpdateClientProfileRequestContent{}

	request.SetName(stringVal(attr(raw, "name")))

	// A description removed from the configuration is sent as an explicit null, which clears it.
	description := stringVal(attr(raw, "description"))
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

	block := firstBlock(attr(raw, "apply_policy"))
	if metadataType := stringVal(attr(block, "external_metadata_type")); metadataType != nil {
		policy.ExternalMetadataType = management.ClientProfileApplyPolicyExternalMetadataTypeEnum(*metadataType)
	}
	if createdBy := stringVal(attr(block, "external_metadata_created_by")); createdBy != nil {
		policy.ExternalMetadataCreatedBy = management.ClientProfileApplyPolicyExternalMetadataCreatedByEnum(*createdBy)
	}

	return policy
}

// expandClientConfiguration only builds the envelope of attributes that are configured, so unset
// attributes are omitted rather than sent as `{}`. On update a removed refresh_token block is sent
// as an empty object, which clears it.
func expandClientConfiguration(raw cty.Value, update bool) *management.ClientProfileClientConfiguration {
	configuration := &management.ClientProfileClientConfiguration{}
	raw = firstBlock(attr(raw, "client_configuration"))

	if block := firstBlock(attr(raw, "refresh_token")); isSet(block) {
		refreshToken := &management.ClientProfileClientConfigurationRefreshToken{}
		if v := intVal(envelopeValue(block, "token_lifetime")); v != nil {
			refreshToken.TokenLifetime = &management.ClientProfileRefreshTokenTokenLifetime{DefaultValue: v}
		}
		if v := intVal(envelopeValue(block, "idle_token_lifetime")); v != nil {
			refreshToken.IdleTokenLifetime = &management.ClientProfileRefreshTokenIdleTokenLifetime{DefaultValue: v}
		}
		if v := intVal(envelopeValue(block, "leeway")); v != nil {
			refreshToken.Leeway = &management.ClientProfileRefreshTokenLeeway{DefaultValue: v}
		}
		configuration.RefreshToken = refreshToken
	} else if update {
		configuration.RefreshToken = &management.ClientProfileClientConfigurationRefreshToken{}
	}

	if v := stringVal(envelopeValue(raw, "organization_usage")); v != nil {
		usage := management.ClientProfileClientConfigurationOrganizationUsageEnum(*v)
		configuration.OrganizationUsage = &management.ClientProfileClientConfigurationOrganizationUsage{DefaultValue: &usage}
	}

	if v := stringVal(envelopeValue(raw, "organization_require_behavior")); v != nil {
		behavior := management.ClientProfileClientConfigurationOrganizationRequireBehaviorEnum(*v)
		configuration.OrganizationRequireBehavior = &management.ClientProfileClientConfigurationOrganizationRequireBehavior{DefaultValue: &behavior}
	}

	if methods := stringsVal(envelopeValue(raw, "organization_discovery_methods")); len(methods) > 0 {
		discovery := make([]management.ClientProfileClientConfigurationOrganizationDiscoveryMethodEnum, 0, len(methods))
		for _, method := range methods {
			discovery = append(discovery, management.ClientProfileClientConfigurationOrganizationDiscoveryMethodEnum(method))
		}
		configuration.OrganizationDiscoveryMethods = &management.ClientProfileClientConfigurationOrganizationDiscoveryMethods{DefaultValue: discovery}
	}

	if v := boolVal(envelopeValue(raw, "require_proof_of_possession")); v != nil {
		configuration.RequireProofOfPossession = &management.ClientProfileClientConfigurationRequireProofOfPossession{DefaultValue: v}
	}

	return configuration
}
