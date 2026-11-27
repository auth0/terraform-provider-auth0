package client

import (
	"maps"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cimdData builds a *schema.ResourceData whose GetRawConfig() holds the given attribute values and
// null for every other attribute, mirroring how Terraform populates RawConfig on a real plan/apply.
// A nil state marks a new resource.
func cimdData(
	t *testing.T,
	config map[string]cty.Value,
	state *terraform.InstanceState,
	attrDiffs map[string]*terraform.ResourceAttrDiff,
) *schema.ResourceData {
	t.Helper()

	sm := schema.InternalMap(cimdClientSchema())

	attrTypes := sm.CoreConfigSchema().ImpliedType().AttributeTypes()
	vals := make(map[string]cty.Value, len(attrTypes))
	for name, ty := range attrTypes {
		vals[name] = cty.NullVal(ty)
	}
	maps.Copy(vals, config)

	diff := &terraform.InstanceDiff{RawConfig: cty.ObjectVal(vals), Attributes: attrDiffs}

	data, err := sm.Data(state, diff)
	require.NoError(t, err)

	if state == nil {
		data.MarkNewResource()
	}

	return data
}

func cimdDataWithAgentID(
	t *testing.T,
	configAgentID cty.Value,
	state *terraform.InstanceState,
	agentIDDiff *terraform.ResourceAttrDiff,
) *schema.ResourceData {
	t.Helper()

	var attrDiffs map[string]*terraform.ResourceAttrDiff
	if agentIDDiff != nil {
		attrDiffs = map[string]*terraform.ResourceAttrDiff{"agent_id": agentIDDiff}
	}

	return cimdData(t, map[string]cty.Value{"agent_id": configAgentID}, state, attrDiffs)
}

// expandedCIMDRequest returns the JSON body that createCIMDClient and updateCIMDClient would send.
func expandedCIMDRequest(t *testing.T, data *schema.ResourceData) string {
	t.Helper()

	req := expandCIMDClient(data)
	if !data.IsNewResource() {
		applyCIMDNullFields(data, req)
	}

	body, err := req.MarshalJSON()
	require.NoError(t, err)

	return string(body)
}

func TestExpandCIMDClientAgentID(t *testing.T) {
	const (
		agentA = "agt_aaaaaaaaaaaaaaaaaaaaaa"
		agentB = "agt_bbbbbbbbbbbbbbbbbbbbbb"
	)

	linkedState := func() *terraform.InstanceState {
		return &terraform.InstanceState{
			ID:         "tpc_test",
			Attributes: map[string]string{"agent_id": agentA},
		}
	}

	expand := expandedCIMDRequest

	t.Run("it sends agent_id on create", func(t *testing.T) {
		data := cimdDataWithAgentID(t, cty.StringVal(agentA), nil, nil)

		assert.JSONEq(t, `{"agent_id":"`+agentA+`"}`, expand(t, data))
	})

	t.Run("it omits agent_id on create when not configured", func(t *testing.T) {
		data := cimdDataWithAgentID(t, cty.NullVal(cty.String), nil, nil)

		assert.JSONEq(t, `{}`, expand(t, data))
	})

	t.Run("it sends agent_id on update when it did not change", func(t *testing.T) {
		data := cimdDataWithAgentID(t, cty.StringVal(agentA), linkedState(), nil)

		assert.JSONEq(t, `{"agent_id":"`+agentA+`"}`, expand(t, data))
	})

	t.Run("it omits agent_id on update when it is not configured and did not change", func(t *testing.T) {
		data := cimdDataWithAgentID(t, cty.NullVal(cty.String), &terraform.InstanceState{ID: "tpc_test"}, nil)

		assert.JSONEq(t, `{}`, expand(t, data))
	})

	t.Run("it sends the new agent_id on update when it changes", func(t *testing.T) {
		data := cimdDataWithAgentID(t, cty.StringVal(agentB), linkedState(), &terraform.ResourceAttrDiff{
			Old: agentA,
			New: agentB,
		})

		assert.JSONEq(t, `{"agent_id":"`+agentB+`"}`, expand(t, data))
	})

	t.Run("it sends a null agent_id on update when it is removed", func(t *testing.T) {
		data := cimdDataWithAgentID(t, cty.NullVal(cty.String), linkedState(), &terraform.ResourceAttrDiff{
			Old:        agentA,
			New:        "",
			NewRemoved: true,
		})

		assert.JSONEq(t, `{"agent_id":null}`, expand(t, data))
	})
}

func cimdFedCMLoginConfig(isEnabled bool) cty.Value {
	return cty.ListVal([]cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"google": cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{"is_enabled": cty.BoolVal(isEnabled)}),
			}),
		}),
	})
}

func TestExpandCIMDClientFedCMLogin(t *testing.T) {
	enabledState := func() *terraform.InstanceState {
		return &terraform.InstanceState{
			ID: "tpc_test",
			Attributes: map[string]string{
				"fedcm_login.#":                     "1",
				"fedcm_login.0.google.#":            "1",
				"fedcm_login.0.google.0.is_enabled": "true",
			},
		}
	}

	t.Run("it sends fedcm_login on create", func(t *testing.T) {
		data := cimdData(t, map[string]cty.Value{"fedcm_login": cimdFedCMLoginConfig(true)}, nil, nil)

		assert.JSONEq(t, `{"fedcm_login":{"google":{"is_enabled":true}}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends is_enabled false on create", func(t *testing.T) {
		data := cimdData(t, map[string]cty.Value{"fedcm_login": cimdFedCMLoginConfig(false)}, nil, nil)

		assert.JSONEq(t, `{"fedcm_login":{"google":{"is_enabled":false}}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it omits fedcm_login on create when not configured", func(t *testing.T) {
		data := cimdData(t, nil, nil, nil)

		assert.JSONEq(t, `{}`, expandedCIMDRequest(t, data))
	})

	t.Run("it omits fedcm_login on update when it did not change", func(t *testing.T) {
		data := cimdData(t, map[string]cty.Value{"fedcm_login": cimdFedCMLoginConfig(true)}, enabledState(), nil)

		assert.JSONEq(t, `{}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends the new fedcm_login on update when it changes", func(t *testing.T) {
		data := cimdData(t, map[string]cty.Value{"fedcm_login": cimdFedCMLoginConfig(false)}, enabledState(), map[string]*terraform.ResourceAttrDiff{
			"fedcm_login.0.google.0.is_enabled": {Old: "true", New: "false"},
		})

		assert.JSONEq(t, `{"fedcm_login":{"google":{"is_enabled":false}}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it omits fedcm_login on update when it is not configured and did not change", func(t *testing.T) {
		data := cimdData(t, nil, &terraform.InstanceState{ID: "tpc_test"}, nil)

		assert.JSONEq(t, `{}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends a null fedcm_login on update when the block is removed", func(t *testing.T) {
		data := cimdData(t, nil, enabledState(), map[string]*terraform.ResourceAttrDiff{
			"fedcm_login.#":                     {Old: "1", New: "0"},
			"fedcm_login.0.google.#":            {Old: "1", New: "0"},
			"fedcm_login.0.google.0.is_enabled": {Old: "true", New: "", NewRemoved: true},
		})

		assert.JSONEq(t, `{"fedcm_login":null}`, expandedCIMDRequest(t, data))
	})
}

func cimdIdentityAssertionAuthorizationGrantConfig(active bool) cty.Value {
	return cty.ListVal([]cty.Value{
		cty.ObjectVal(map[string]cty.Value{"active": cty.BoolVal(active)}),
	})
}

func TestExpandCIMDClientIdentityAssertionAuthorizationGrant(t *testing.T) {
	activeState := func() *terraform.InstanceState {
		return &terraform.InstanceState{
			ID: "tpc_test",
			Attributes: map[string]string{
				"identity_assertion_authorization_grant.#":        "1",
				"identity_assertion_authorization_grant.0.active": "true",
			},
		}
	}

	t.Run("it sends identity_assertion_authorization_grant on create", func(t *testing.T) {
		config := map[string]cty.Value{
			"identity_assertion_authorization_grant": cimdIdentityAssertionAuthorizationGrantConfig(true),
		}
		data := cimdData(t, config, nil, nil)

		assert.JSONEq(t, `{"identity_assertion_authorization_grant":{"active":true}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends active false on create", func(t *testing.T) {
		config := map[string]cty.Value{
			"identity_assertion_authorization_grant": cimdIdentityAssertionAuthorizationGrantConfig(false),
		}
		data := cimdData(t, config, nil, nil)

		assert.JSONEq(t, `{"identity_assertion_authorization_grant":{"active":false}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it omits identity_assertion_authorization_grant on create when not configured", func(t *testing.T) {
		data := cimdData(t, nil, nil, nil)

		assert.JSONEq(t, `{}`, expandedCIMDRequest(t, data))
	})

	t.Run("it omits identity_assertion_authorization_grant on update when it did not change", func(t *testing.T) {
		config := map[string]cty.Value{
			"identity_assertion_authorization_grant": cimdIdentityAssertionAuthorizationGrantConfig(true),
		}
		data := cimdData(t, config, activeState(), nil)

		assert.JSONEq(t, `{}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends the new value on update when it changes", func(t *testing.T) {
		config := map[string]cty.Value{
			"identity_assertion_authorization_grant": cimdIdentityAssertionAuthorizationGrantConfig(false),
		}
		data := cimdData(t, config, activeState(), map[string]*terraform.ResourceAttrDiff{
			"identity_assertion_authorization_grant.0.active": {Old: "true", New: "false"},
		})

		assert.JSONEq(t, `{"identity_assertion_authorization_grant":{"active":false}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends a null identity_assertion_authorization_grant on update when the block is removed", func(t *testing.T) {
		data := cimdData(t, nil, activeState(), map[string]*terraform.ResourceAttrDiff{
			"identity_assertion_authorization_grant.#":        {Old: "1", New: "0"},
			"identity_assertion_authorization_grant.0.active": {Old: "true", New: "", NewRemoved: true},
		})

		assert.JSONEq(t, `{"identity_assertion_authorization_grant":null}`, expandedCIMDRequest(t, data))
	})
}

func cimdOIDCSupportConfig(isAllowed bool, allowAllScopes cty.Value, allowedScopes cty.Value) cty.Value {
	return cty.ListVal([]cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"is_allowed":       cty.BoolVal(isAllowed),
			"allow_all_scopes": allowAllScopes,
			"allowed_scopes":   allowedScopes,
		}),
	})
}

func TestExpandCIMDClientOIDCSupport(t *testing.T) {
	var (
		noAllowAll = cty.NullVal(cty.Bool)
		noScopes   = cty.NullVal(cty.Set(cty.String))
		someScopes = cty.SetVal([]cty.Value{cty.StringVal("profile"), cty.StringVal("email")})
	)

	allowAllState := func() *terraform.InstanceState {
		return &terraform.InstanceState{
			ID: "tpc_test",
			Attributes: map[string]string{
				"oidc_support.#":                  "1",
				"oidc_support.0.is_allowed":       "true",
				"oidc_support.0.allow_all_scopes": "true",
			},
		}
	}

	t.Run("it sends allow_all_scopes on create", func(t *testing.T) {
		config := map[string]cty.Value{"oidc_support": cimdOIDCSupportConfig(true, cty.True, noScopes)}
		data := cimdData(t, config, nil, nil)

		assert.JSONEq(t, `{"oidc_support":{"is_allowed":true,"allow_all_scopes":true}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends sorted allowed_scopes on create", func(t *testing.T) {
		config := map[string]cty.Value{"oidc_support": cimdOIDCSupportConfig(true, noAllowAll, someScopes)}
		data := cimdData(t, config, nil, nil)

		assert.JSONEq(t, `{"oidc_support":{"is_allowed":true,"allowed_scopes":["email","profile"]}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends only is_allowed when it is true and no scopes are set", func(t *testing.T) {
		config := map[string]cty.Value{"oidc_support": cimdOIDCSupportConfig(true, cty.False, noScopes)}
		data := cimdData(t, config, nil, nil)

		assert.JSONEq(t, `{"oidc_support":{"is_allowed":true}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends only is_allowed when it is false", func(t *testing.T) {
		config := map[string]cty.Value{"oidc_support": cimdOIDCSupportConfig(false, cty.False, noScopes)}
		data := cimdData(t, config, nil, nil)

		assert.JSONEq(t, `{"oidc_support":{"is_allowed":false}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it omits oidc_support on create when not configured", func(t *testing.T) {
		data := cimdData(t, nil, nil, nil)

		assert.JSONEq(t, `{}`, expandedCIMDRequest(t, data))
	})

	t.Run("it omits oidc_support on update when it did not change", func(t *testing.T) {
		config := map[string]cty.Value{"oidc_support": cimdOIDCSupportConfig(true, cty.True, noScopes)}
		data := cimdData(t, config, allowAllState(), nil)

		assert.JSONEq(t, `{}`, expandedCIMDRequest(t, data))
	})

	t.Run("it sends the whole object on update when it changes", func(t *testing.T) {
		config := map[string]cty.Value{"oidc_support": cimdOIDCSupportConfig(true, noAllowAll, someScopes)}
		data := cimdData(t, config, allowAllState(), map[string]*terraform.ResourceAttrDiff{
			"oidc_support.0.allow_all_scopes": {Old: "true", New: "false"},
		})

		assert.JSONEq(t, `{"oidc_support":{"is_allowed":true,"allowed_scopes":["email","profile"]}}`, expandedCIMDRequest(t, data))
	})

	t.Run("it never sends null when the block is removed", func(t *testing.T) {
		data := cimdData(t, nil, allowAllState(), map[string]*terraform.ResourceAttrDiff{
			"oidc_support.#": {Old: "1", New: "0"},
		})

		assert.JSONEq(t, `{}`, expandedCIMDRequest(t, data))
	})
}
