package networkacl

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// flattenNetworkACL flattens a NetworkACL object into the Terraform schema.
// It handles nil pointers and collects errors using multierror.
func flattenNetworkACL(data *schema.ResourceData, networkACL *management.NetworkACL) error {
	if networkACL == nil {
		return nil
	}

	result := multierror.Append(
		data.Set("description", networkACL.Description),
		data.Set("active", networkACL.Active),
		data.Set("priority", networkACL.Priority),
	)

	if networkACL.Rule != nil {
		rule := make(map[string]interface{})

		if networkACL.Rule.Action != nil {
			actionResult := flattenNetworkACLRuleAction(networkACL.Rule.Action)
			if actionResult != nil {
				rule["action"] = actionResult
			}
		}

		if networkACL.Rule.Match != nil {
			matchResult := flattenNetworkACLRule(networkACL.Rule.Match)
			if matchResult != nil {
				rule["match"] = matchResult
			}
		}

		if networkACL.Rule.NotMatch != nil {
			notMatchResult := flattenNetworkACLRule(networkACL.Rule.NotMatch)
			if notMatchResult != nil {
				rule["not_match"] = notMatchResult
			}
		}

		if networkACL.Rule.Scope != nil {
			rule["scope"] = networkACL.Rule.Scope
		}

		// Only set the rule if we have at least one field.
		if len(rule) > 0 {
			if err := data.Set("rule", []interface{}{rule}); err != nil {
				result = multierror.Append(result, err)
			}
		}
	}

	return result.ErrorOrNil()
}

// flattenNetworkACLRuleAction flattens a NetworkACLRuleAction into a slice of interfaces.
// It handles nil pointers and returns nil if the action is nil.
func flattenNetworkACLRuleAction(action *management.NetworkACLRuleAction) []interface{} {
	if action == nil {
		return nil
	}

	actionMap := map[string]interface{}{
		"block":    action.Block,
		"allow":    action.Allow,
		"log":      action.Log,
		"redirect": action.Redirect,
	}

	// Only include redirect_uri if it's not nil.
	if action.RedirectURI != nil {
		actionMap["redirect_uri"] = action.RedirectURI
	}

	return []interface{}{actionMap}
}

// flattenNetworkACLRule flattens a NetworkACLRuleMatch into a slice of interfaces.
// It handles nil pointers and returns nil if the match is nil.
func flattenNetworkACLRule(match *management.NetworkACLRuleMatch) []interface{} {
	if match == nil {
		return nil
	}

	m := map[string]interface{}{}

	// Handle slice of integers.
	if len(match.Asns) > 0 {
		m["asns"] = match.Asns
	}

	// Handle string slices - only set if not nil and not empty.
	if match.GeoCountryCodes != nil && len(*match.GeoCountryCodes) > 0 {
		m["geo_country_codes"] = *match.GeoCountryCodes
	}

	if match.GeoSubdivisionCodes != nil && len(*match.GeoSubdivisionCodes) > 0 {
		m["geo_subdivision_codes"] = *match.GeoSubdivisionCodes
	}

	if match.IPv4Cidrs != nil && len(*match.IPv4Cidrs) > 0 {
		m["ipv4_cidrs"] = *match.IPv4Cidrs
	}

	if match.IPv6Cidrs != nil && len(*match.IPv6Cidrs) > 0 {
		m["ipv6_cidrs"] = *match.IPv6Cidrs
	}

	if match.Ja3Fingerprints != nil && len(*match.Ja3Fingerprints) > 0 {
		m["ja3_fingerprints"] = *match.Ja3Fingerprints
	}

	if match.Ja4Fingerprints != nil && len(*match.Ja4Fingerprints) > 0 {
		m["ja4_fingerprints"] = *match.Ja4Fingerprints
	}

	if match.UserAgents != nil && len(*match.UserAgents) > 0 {
		m["user_agents"] = *match.UserAgents
	}

	// Only return a non-empty map.
	if len(m) > 0 {
		return []interface{}{m}
	}

	return []interface{}{}
}
