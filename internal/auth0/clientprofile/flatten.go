package clientprofile

import (
	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func flattenClientProfile(data *schema.ResourceData, profile *management.ClientProfile) diag.Diagnostics {
	description := ""
	if d := profile.GetDescription(); d != nil {
		description = *d
	}

	result := multierror.Append(
		data.Set("name", profile.GetName()),
		data.Set("description", description),
		data.Set("apply_policy", flattenApplyPolicy(profile.GetApplyPolicy())),
		data.Set("created_at", value.FormatTime(profile.GetCreatedAt())),
		data.Set("updated_at", value.FormatTime(profile.GetUpdatedAt())),
	)

	// Every envelope is optional in the API response, so each one is read defensively.
	configuration := profile.GetClientConfiguration()
	if configuration == nil {
		configuration = &management.ClientProfileClientConfiguration{}
	}

	var usage, behavior, methods, proofOfPossession []interface{}
	if v := configuration.OrganizationUsage; v != nil && v.DefaultValue != nil {
		usage = flattenEnvelope(string(*v.DefaultValue))
	}
	if v := configuration.OrganizationRequireBehavior; v != nil && v.DefaultValue != nil {
		behavior = flattenEnvelope(string(*v.DefaultValue))
	}
	if v := configuration.OrganizationDiscoveryMethods; v != nil && v.DefaultValue != nil {
		discovery := make([]interface{}, 0, len(v.DefaultValue))
		for _, method := range v.DefaultValue {
			discovery = append(discovery, string(method))
		}
		methods = flattenEnvelope(discovery)
	}
	if v := configuration.RequireProofOfPossession; v != nil && v.DefaultValue != nil {
		proofOfPossession = flattenEnvelope(*v.DefaultValue)
	}

	flattened := map[string]interface{}{}
	for key, v := range map[string][]interface{}{
		"refresh_token":                  flattenRefreshToken(configuration.RefreshToken),
		"organization_usage":             usage,
		"organization_require_behavior":  behavior,
		"organization_discovery_methods": methods,
		"require_proof_of_possession":    proofOfPossession,
	} {
		if len(v) > 0 {
			flattened[key] = v
		}
	}

	// An empty client configuration is treated like an absent block.
	var configurationBlock []interface{}
	if len(flattened) > 0 {
		configurationBlock = []interface{}{flattened}
	}

	result = multierror.Append(result, data.Set("client_configuration", configurationBlock))

	return diag.FromErr(result.ErrorOrNil())
}

// flattenEnvelope wraps a value in the single-item block that mirrors the API's `default_value` wrapper.
func flattenEnvelope(v interface{}) []interface{} {
	return []interface{}{map[string]interface{}{"default_value": v}}
}

func flattenApplyPolicy(policy *management.ClientProfileApplyPolicy) []interface{} {
	if policy == nil {
		return nil
	}

	return []interface{}{map[string]interface{}{
		"external_metadata_type":       string(policy.GetExternalMetadataType()),
		"external_metadata_created_by": string(policy.GetExternalMetadataCreatedBy()),
	}}
}

// flattenRefreshToken writes the server values back, including the lifetime Auth0 derives when
// only one of the two is configured. An empty object is treated like an absent block.
func flattenRefreshToken(refreshToken *management.ClientProfileClientConfigurationRefreshToken) []interface{} {
	if refreshToken == nil {
		return nil
	}

	flattened := map[string]interface{}{}
	if v := refreshToken.TokenLifetime; v != nil && v.DefaultValue != nil {
		flattened["token_lifetime"] = flattenEnvelope(*v.DefaultValue)
	}
	if v := refreshToken.IdleTokenLifetime; v != nil && v.DefaultValue != nil {
		flattened["idle_token_lifetime"] = flattenEnvelope(*v.DefaultValue)
	}
	if v := refreshToken.Leeway; v != nil && v.DefaultValue != nil {
		flattened["leeway"] = flattenEnvelope(*v.DefaultValue)
	}

	if len(flattened) == 0 {
		return nil
	}

	return []interface{}{flattened}
}
