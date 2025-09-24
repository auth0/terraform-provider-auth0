package user_attribute_profile

import (
	"sort"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenUserAttributeProfile(data *schema.ResourceData, userAttributeProfile *management.UserAttributeProfile) error {
	result := multierror.Append(
		data.Set("name", userAttributeProfile.GetName()),
		data.Set("user_id", flattenUserAttributeProfileUserID(userAttributeProfile.UserID)),
		data.Set("user_attributes", flattenUserAttributeProfileUserAttributes(userAttributeProfile.UserAttributes)),
	)

	return result.ErrorOrNil()
}

func flattenUserAttributeProfileUserID(userID *management.UserAttributeProfileUserID) []interface{} {
	m := map[string]interface{}{}

	// Set oidc_mapping only if it has a value
	if oidcMapping := userID.GetOIDCMapping(); oidcMapping != "" {
		m["oidc_mapping"] = oidcMapping
	}

	// Set scim_mapping only if it has a value
	if scimMapping := userID.GetSCIMMapping(); scimMapping != "" {
		m["scim_mapping"] = scimMapping
	}

	// Set saml_mapping only if it has values
	if userID.SAMLMapping != nil && len(*userID.SAMLMapping) > 0 {
		m["saml_mapping"] = *userID.SAMLMapping
	}

	// Set strategy_overrides only if they exist
	if len(userID.StrategyOverrides) > 0 {
		m["strategy_overrides"] = flattenUserIDStrategyOverrides(userID.StrategyOverrides)
	}

	// Only return the user_id block if it has any configured fields
	if len(m) > 0 {
		return []interface{}{m}
	}

	return nil
}

func flattenUserIDStrategyOverrides(overrides map[string]*management.UserAttributeProfileStrategyOverrides) []interface{} {
	if len(overrides) == 0 {
		return nil
	}

	// Create a sorted slice of strategy names for consistent ordering
	var strategyNames []string
	for strategy := range overrides {
		strategyNames = append(strategyNames, strategy)
	}
	sort.Strings(strategyNames)

	result := make([]interface{}, 0, len(overrides))

	for _, strategy := range strategyNames {
		override := overrides[strategy]
		overrideMap := map[string]interface{}{
			"strategy": strategy,
		}

		// Set oidc_mapping only if it has a value
		if oidcMapping := override.GetOIDCMapping(); oidcMapping != "" {
			overrideMap["oidc_mapping"] = oidcMapping
		}

		// Set scim_mapping only if it has a value
		if scimMapping := override.GetSCIMMapping(); scimMapping != "" {
			overrideMap["scim_mapping"] = scimMapping
		}

		// Set saml_mapping only if it has values
		if override.SAMLMapping != nil && len(*override.SAMLMapping) > 0 {
			overrideMap["saml_mapping"] = *override.SAMLMapping
		}

		result = append(result, overrideMap)
	}

	return result
}

func flattenUserAttributeProfileUserAttributes(userAttributes map[string]*management.UserAttributeProfileUserAttributes) []interface{} {
	if len(userAttributes) == 0 {
		return nil
	}

	// Create a sorted slice of attribute names for consistent ordering
	var attrNames []string
	for attrName := range userAttributes {
		attrNames = append(attrNames, attrName)
	}
	sort.Strings(attrNames)

	result := make([]interface{}, 0, len(userAttributes))

	for _, attrName := range attrNames {
		userAttr := userAttributes[attrName]
		attrMap := map[string]interface{}{
			"name":             attrName,
			"description":      userAttr.GetDescription(),
			"label":            userAttr.GetLabel(),
			"profile_required": userAttr.GetProfileRequired(),
			"auth0_mapping":    userAttr.GetAuth0Mapping(),
		}

		// Set saml_mapping only if it has values (allows proper field removal detection)
		if userAttr.SAMLMapping != nil && len(*userAttr.SAMLMapping) > 0 {
			attrMap["saml_mapping"] = *userAttr.SAMLMapping
		}

		// Set scim_mapping only if it has a value (allows proper field removal detection)
		if scimMapping := userAttr.GetSCIMMapping(); scimMapping != "" {
			attrMap["scim_mapping"] = scimMapping
		}

		// Set oidc_mapping only if it has values (allows proper field removal detection)
		if userAttr.OIDCMapping != nil {
			attrMap["oidc_mapping"] = []interface{}{
				map[string]interface{}{
					"mapping":      userAttr.OIDCMapping.GetMapping(),
					"display_name": userAttr.OIDCMapping.GetDisplayName(),
				},
			}
		}

		// Set strategy_overrides only if they exist (allows proper field removal detection)
		if len(userAttr.StrategyOverrides) > 0 {
			attrMap["strategy_overrides"] = flattenUserAttributeStrategyOverrides(userAttr.StrategyOverrides)
		}

		result = append(result, attrMap)
	}

	return result
}

func flattenUserAttributeStrategyOverrides(overrides map[string]*management.UserAttributesStrategyOverride) []interface{} {
	if len(overrides) == 0 {
		return nil
	}

	// Create a sorted slice of strategy names for consistent ordering
	var strategyNames []string
	for strategy := range overrides {
		strategyNames = append(strategyNames, strategy)
	}
	sort.Strings(strategyNames)

	result := make([]interface{}, 0, len(overrides))

	for _, strategy := range strategyNames {
		override := overrides[strategy]
		overrideMap := map[string]interface{}{
			"strategy":     strategy,
			"scim_mapping": override.GetSCIMMapping(),
		}

		// Always set saml_mapping, even if empty (since it's Optional + Computed)
		if override.SAMLMapping != nil {
			overrideMap["saml_mapping"] = *override.SAMLMapping
		} else {
			overrideMap["saml_mapping"] = []string{}
		}

		// Always set oidc_mapping, even if nil (since it's Optional + Computed)
		if override.OIDCMapping != nil {
			overrideMap["oidc_mapping"] = []interface{}{
				map[string]interface{}{
					"mapping":      override.OIDCMapping.GetMapping(),
					"display_name": override.OIDCMapping.GetDisplayName(),
				},
			}
		} else {
			overrideMap["oidc_mapping"] = []interface{}{}
		}

		result = append(result, overrideMap)
	}

	return result
}
