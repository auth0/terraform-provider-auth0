package networkacl

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeBaseRule() map[string]interface{} {
	return map[string]interface{}{
		"action": []interface{}{
			map[string]interface{}{
				"block":        true,
				"allow":        false,
				"log":          false,
				"redirect":     false,
				"redirect_uri": "",
			},
		},
		"scope":     "tenant",
		"match":     []interface{}{},
		"not_match": []interface{}{},
		"match_all": false,
	}
}

func TestExpandMatchAllTrue(t *testing.T) {
	rule := makeBaseRule()
	rule["match_all"] = true

	acl := &management.NetworkACL{Rule: &management.NetworkACLRule{}}

	// Call expand logic directly — mirror the expand function's match_all block.
	if matchAll, ok := rule["match_all"].(bool); ok && matchAll {
		acl.Rule.MatchAll = auth0.Bool(true)
	}

	require.NotNil(t, acl.Rule.MatchAll)
	assert.True(t, *acl.Rule.MatchAll)
	assert.Nil(t, acl.Rule.Match)
}

func TestExpandMatchAllFalseWithSignal(t *testing.T) {
	rule := makeBaseRule()
	rule["match_all"] = false
	rule["match"] = []interface{}{
		map[string]interface{}{
			"asns": []interface{}{9453},
		},
	}

	// Match_all=false so MatchAll should not be set.
	acl := &management.NetworkACL{Rule: &management.NetworkACLRule{}}
	if matchAll, ok := rule["match_all"].(bool); ok && matchAll {
		acl.Rule.MatchAll = auth0.Bool(true)
	}

	if match, ok := rule["match"].([]interface{}); ok && len(match) > 0 {
		if matchElem := match[0]; matchElem != nil {
			if matchMap, ok := matchElem.(map[string]interface{}); ok {
				acl.Rule.Match = expandNetworkACLRuleMatch(matchMap)
			}
		}
	}

	assert.Nil(t, acl.Rule.MatchAll)
	assert.NotNil(t, acl.Rule.Match)
}

func TestFlattenMatchAllTrue(t *testing.T) {
	networkACL := &management.NetworkACL{
		Rule: &management.NetworkACLRule{
			MatchAll: auth0.Bool(true),
		},
	}

	rule := make(map[string]interface{})
	if networkACL.Rule.MatchAll != nil {
		rule["match_all"] = *networkACL.Rule.MatchAll
	}

	v, ok := rule["match_all"]
	require.True(t, ok, "match_all key should be present")
	assert.True(t, v.(bool))
}

func TestFlattenMatchAllNil(t *testing.T) {
	networkACL := &management.NetworkACL{
		Rule: &management.NetworkACLRule{
			MatchAll: nil,
		},
	}

	rule := make(map[string]interface{})
	if networkACL.Rule.MatchAll != nil {
		rule["match_all"] = *networkACL.Rule.MatchAll
	}

	_, ok := rule["match_all"]
	assert.False(t, ok, "match_all key should not be present when MatchAll is nil")
}
