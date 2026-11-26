package client

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cimdDataWithAgentID builds a *schema.ResourceData whose GetRawConfig() holds the given agent_id,
// mirroring how Terraform populates RawConfig on a real plan/apply. A nil state marks a new resource.
func cimdDataWithAgentID(
	t *testing.T,
	configAgentID cty.Value,
	state *terraform.InstanceState,
	agentIDDiff *terraform.ResourceAttrDiff,
) *schema.ResourceData {
	t.Helper()

	sm := schema.InternalMap(cimdClientSchema())

	attrTypes := sm.CoreConfigSchema().ImpliedType().AttributeTypes()
	vals := make(map[string]cty.Value, len(attrTypes))
	for name, ty := range attrTypes {
		vals[name] = cty.NullVal(ty)
	}
	vals["agent_id"] = configAgentID

	diff := &terraform.InstanceDiff{RawConfig: cty.ObjectVal(vals)}
	if agentIDDiff != nil {
		diff.Attributes = map[string]*terraform.ResourceAttrDiff{"agent_id": agentIDDiff}
	}

	data, err := sm.Data(state, diff)
	require.NoError(t, err)

	if state == nil {
		data.MarkNewResource()
	}

	return data
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

	expand := func(t *testing.T, data *schema.ResourceData) string {
		t.Helper()

		req := expandCIMDClient(data)
		if !data.IsNewResource() {
			applyCIMDNullFields(data, req)
		}

		body, err := req.MarshalJSON()
		require.NoError(t, err)

		return string(body)
	}

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
