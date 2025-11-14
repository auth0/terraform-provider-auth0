package connection

import (
	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenConnectionProfile(data *schema.ResourceData, profile *management.GetConnectionProfileResponseContent) error {
	result := multierror.Append(
		data.Set("name", profile.GetName()),
	)

	org := profile.GetOrganization()
	orgMap := map[string]interface{}{
		"show_as_button":             org.GetShowAsButton(),
		"assign_membership_on_login": org.GetAssignMembershipOnLogin(),
	}
	result = multierror.Append(result, data.Set("organization", []interface{}{orgMap}))

	result = multierror.Append(result, data.Set("connection_name_prefix_template", profile.GetConnectionNamePrefixTemplate()))

	if features := profile.GetEnabledFeatures(); len(features) > 0 {
		featureList := make([]string, len(features))
		for i, f := range features {
			featureList[i] = string(f)
		}
		result = multierror.Append(result, data.Set("enabled_features", featureList))
	} else {
		result = multierror.Append(result, data.Set("enabled_features", []string{}))
	}

	// ConnectionConfig is empty in the SDK, so preserve from config.
	if connConfig, ok := data.Get("connection_config").([]interface{}); ok && len(connConfig) > 0 {
		result = multierror.Append(result, data.Set("connection_config", connConfig))
	}

	overrides := profile.GetStrategyOverrides()
	overridesMap := map[string]interface{}{}

	if pingfed := overrides.GetPingfederate(); !isEmptyStrategyOverride(pingfed) {
		overridesMap["pingfederate"] = []interface{}{flattenStrategyOverride(pingfed)}
	}
	if ad := overrides.GetAd(); !isEmptyStrategyOverride(ad) {
		overridesMap["ad"] = []interface{}{flattenStrategyOverride(ad)}
	}
	if adfs := overrides.GetAdfs(); !isEmptyStrategyOverride(adfs) {
		overridesMap["adfs"] = []interface{}{flattenStrategyOverride(adfs)}
	}
	if waad := overrides.GetWaad(); !isEmptyStrategyOverride(waad) {
		overridesMap["waad"] = []interface{}{flattenStrategyOverride(waad)}
	}
	if googleApps := overrides.GetGoogleApps(); !isEmptyStrategyOverride(googleApps) {
		overridesMap["google_apps"] = []interface{}{flattenStrategyOverride(googleApps)}
	}
	if okta := overrides.GetOkta(); !isEmptyStrategyOverride(okta) {
		overridesMap["okta"] = []interface{}{flattenStrategyOverride(okta)}
	}
	if oidc := overrides.GetOidc(); !isEmptyStrategyOverride(oidc) {
		overridesMap["oidc"] = []interface{}{flattenStrategyOverride(oidc)}
	}
	if samlp := overrides.GetSamlp(); !isEmptyStrategyOverride(samlp) {
		overridesMap["samlp"] = []interface{}{flattenStrategyOverride(samlp)}
	}

	if len(overridesMap) > 0 {
		result = multierror.Append(result, data.Set("strategy_overrides", []interface{}{overridesMap}))
	}

	return result.ErrorOrNil()
}

func flattenStrategyOverride(override management.ConnectionProfileStrategyOverride) map[string]interface{} {
	result := map[string]interface{}{}

	if features := override.GetEnabledFeatures(); len(features) > 0 {
		featureList := make([]string, 0)
		for _, f := range features {
			featureList = append(featureList, string(f))
		}
		if len(featureList) > 0 {
			result["enabled_features"] = featureList
		}
	}

	config := override.GetConnectionConfig()
	if len(config.GetExtraProperties()) > 0 {
		result["connection_config"] = []interface{}{map[string]interface{}{}}
	}

	return result
}

func isEmptyStrategyOverride(override management.ConnectionProfileStrategyOverride) bool {
	features := override.GetEnabledFeatures()
	config := override.GetConnectionConfig()
	return len(features) == 0 && len(config.GetExtraProperties()) == 0
}
