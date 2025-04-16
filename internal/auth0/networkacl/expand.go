package networkacl

import (
	"errors"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandNetworkACL(data *schema.ResourceData) (*management.NetworkACL, error) {
	cfg := data.GetRawConfig()
	networkACL := &management.NetworkACL{}

	// Required fields should always be included.
	networkACL.Description = value.String(cfg.GetAttr("description"))
	networkACL.Active = value.Bool(cfg.GetAttr("active"))
	networkACL.Priority = value.Int(cfg.GetAttr("priority"))

	// Rule is required, so we can directly access it.
	rule := data.Get("rule").([]interface{})[0].(map[string]interface{})
	networkACL.Rule = &management.NetworkACLRule{}

	if action, ok := rule["action"].([]interface{}); ok && len(action) > 0 {
		actionMap := action[0].(map[string]interface{})
		networkACL.Rule.Action = &management.NetworkACLRuleAction{}

		// Only set properties that are explicitly true.
		if block, ok := actionMap["block"].(bool); ok && block {
			networkACL.Rule.Action.Block = auth0.Bool(true)
		}

		if allow, ok := actionMap["allow"].(bool); ok && allow {
			networkACL.Rule.Action.Allow = auth0.Bool(true)
		}

		if log, ok := actionMap["log"].(bool); ok && log {
			networkACL.Rule.Action.Log = auth0.Bool(true)
		}

		// Validate the redirect action.
		if err := validateActionRedirect(actionMap); err != nil {
			return nil, err
		}

		if redirect, ok := actionMap["redirect"].(bool); ok && redirect {
			networkACL.Rule.Action.Redirect = auth0.Bool(true)
			// Only set RedirectURI if redirect is true.
			if redirectURI, ok := actionMap["redirect_uri"].(string); ok && redirectURI != "" {
				networkACL.Rule.Action.RedirectURI = auth0.String(redirectURI)
			}
		}
	}

	// Validate match and not_match.
	if err := validateMatchAndNotMatch(rule); err != nil {
		return nil, err
	}

	if match, ok := rule["match"].([]interface{}); ok && len(match) > 0 {
		if matchElem := match[0]; matchElem != nil {
			if matchMap, ok := matchElem.(map[string]interface{}); ok {
				networkACL.Rule.Match = expandNetworkACLRuleMatch(matchMap)
			}
		}
	}

	if notMatch, ok := rule["not_match"].([]interface{}); ok && len(notMatch) > 0 {
		if notMatchElem := notMatch[0]; notMatchElem != nil {
			if notMatchMap, ok := notMatchElem.(map[string]interface{}); ok {
				networkACL.Rule.NotMatch = expandNetworkACLRuleMatch(notMatchMap)
			}
		}
	}

	if scope, ok := rule["scope"].(string); ok {
		networkACL.Rule.Scope = auth0.String(scope)
	}

	return networkACL, nil
}

func expandNetworkACLRuleMatch(m map[string]interface{}) *management.NetworkACLRuleMatch {
	if m == nil {
		return nil
	}

	match := &management.NetworkACLRuleMatch{}

	if asns, ok := m["asns"].([]interface{}); ok {
		if len(asns) == 0 {
			match.Asns = nil
		} else {
			asnsList := make([]int, len(asns))
			for i, v := range asns {
				asnsList[i] = v.(int)
			}
			match.Asns = asnsList
		}
	}

	if v, ok := m["geo_country_codes"].([]interface{}); ok {
		match.GeoCountryCodes = expandStringList(v)
	}

	if v, ok := m["geo_subdivision_codes"].([]interface{}); ok {
		match.GeoSubdivisionCodes = expandStringList(v)
	}

	if v, ok := m["ipv4_cidrs"].([]interface{}); ok {
		match.IPv4Cidrs = expandStringList(v)
	}

	if v, ok := m["ipv6_cidrs"].([]interface{}); ok {
		match.IPv6Cidrs = expandStringList(v)
	}

	if v, ok := m["ja3_fingerprints"].([]interface{}); ok {
		match.Ja3Fingerprints = expandStringList(v)
	}

	if v, ok := m["ja4_fingerprints"].([]interface{}); ok {
		match.Ja4Fingerprints = expandStringList(v)
	}

	if v, ok := m["user_agents"].([]interface{}); ok {
		match.UserAgents = expandStringList(v)
	}

	return match
}

func expandStringList(list []interface{}) *[]string {
	if len(list) == 0 {
		return nil
	}

	result := make([]string, len(list))
	for i, v := range list {
		if str, ok := v.(string); ok {
			result[i] = str
		}
	}
	return &result
}

func validateMatchAndNotMatch(rule map[string]interface{}) error {
	matchExists := false
	notMatchExists := false

	if match, ok := rule["match"].([]interface{}); ok && len(match) > 0 {
		matchExists = true
		// Check if match has any criteria.
		if matchMap, ok := match[0].(map[string]interface{}); ok {
			matchObj := expandNetworkACLRuleMatch(matchMap)
			if isEmptyNetworkACLRuleMatch(matchObj) {
				return errors.New("when 'match' is specified, it must contain at least one match criteria")
			}
		}
	}

	if notMatch, ok := rule["not_match"].([]interface{}); ok && len(notMatch) > 0 {
		notMatchExists = true
		// Check if not_match has any criteria.
		if notMatchMap, ok := notMatch[0].(map[string]interface{}); ok {
			notMatchObj := expandNetworkACLRuleMatch(notMatchMap)
			if isEmptyNetworkACLRuleMatch(notMatchObj) {
				return errors.New("when 'not_match' is specified, it must contain at least one match criteria")
			}
		}
	}

	if matchExists && notMatchExists {
		return errors.New("a network ACL rule cannot specify both 'match' and 'not_match' simultaneously")
	}

	return nil
}

func isEmptyNetworkACLRuleMatch(match *management.NetworkACLRuleMatch) bool {
	// Check if all fields are nil or empty.
	return len(match.Asns) == 0 && match.GetGeoCountryCodes() == nil && match.GetGeoSubdivisionCodes() == nil &&
		match.GetIPv4Cidrs() == nil && match.GetIPv6Cidrs() == nil && match.GetJa3Fingerprints() == nil && match.GetJa4Fingerprints() == nil && match.GetUserAgents() == nil
}

func validateActionRedirect(actionMap map[string]interface{}) error {
	// Add check when 'redirect_uri' action is specified, 'redirect' must also be true.
	if redirectURI, ok := actionMap["redirect_uri"].(string); ok && redirectURI != "" {
		if redirect, ok := actionMap["redirect"].(bool); ok && !redirect {
			return errors.New("the 'redirect' must be set to true when a 'redirect_uri' is specified")
		}
	}
	return nil
}
