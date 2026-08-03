package client

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTokenVaultResourceData builds a *schema.ResourceData for a brand-new
// resource (IsNewResource() == true) whose GetRawConfig() reflects the given
// token_vault_privileged_access value, mirroring how Terraform populates
// RawConfig on a real plan/apply (TestResourceDataRaw alone leaves it null).
func newTokenVaultResourceData(t *testing.T, tokenVaultPrivilegedAccess cty.Value) *schema.ResourceData {
	t.Helper()

	sm := schema.InternalMap(NewResource().Schema)
	impliedType := sm.CoreConfigSchema().ImpliedType()

	attrTypes := impliedType.AttributeTypes()
	vals := make(map[string]cty.Value, len(attrTypes))
	for name, ty := range attrTypes {
		vals[name] = cty.NullVal(ty)
	}
	vals["token_vault_privileged_access"] = tokenVaultPrivilegedAccess

	data, err := sm.Data(nil, &terraform.InstanceDiff{RawConfig: cty.ObjectVal(vals)})
	require.NoError(t, err)

	data.MarkNewResource()

	return data
}

// updateTokenVaultResourceData builds a *schema.ResourceData for an update to
// an existing resource, computing a real diff between a prior state holding
// oldTokenVaultPrivilegedAccess and a new config holding
// newTokenVaultPrivilegedAccess. This is what makes HasChange behave as it
// would during a real plan/apply, unlike hand-assembling an InstanceDiff.
func updateTokenVaultResourceData(t *testing.T, oldTokenVaultPrivilegedAccess, newTokenVaultPrivilegedAccess cty.Value) *schema.ResourceData {
	t.Helper()

	res := NewResource()
	sm := schema.InternalMap(res.Schema)
	impliedType := sm.CoreConfigSchema().ImpliedType()
	attrTypes := impliedType.AttributeTypes()

	oldVals := make(map[string]cty.Value, len(attrTypes))
	newVals := make(map[string]cty.Value, len(attrTypes))
	for name, ty := range attrTypes {
		oldVals[name] = cty.NullVal(ty)
		newVals[name] = cty.NullVal(ty)
	}
	oldVals["token_vault_privileged_access"] = oldTokenVaultPrivilegedAccess
	newVals["token_vault_privileged_access"] = newTokenVaultPrivilegedAccess

	state := terraform.NewInstanceStateShimmedFromValue(cty.ObjectVal(oldVals), 0)
	state.RawConfig = cty.ObjectVal(oldVals)
	resourceConfig := terraform.NewResourceConfigShimmed(cty.ObjectVal(newVals), sm.CoreConfigSchema())

	instanceDiff, err := res.Diff(context.Background(), state, resourceConfig, nil)
	require.NoError(t, err)
	instanceDiff.RawConfig = cty.ObjectVal(newVals)

	data, err := sm.Data(state, instanceDiff)
	require.NoError(t, err)

	return data
}

func tokenVaultCredentialsType() cty.Type {
	return cty.List(cty.Object(map[string]cty.Type{
		"credential_type": cty.String,
		"pem":             cty.String,
	}))
}

func tokenVaultGrantsType() cty.Type {
	return cty.List(cty.Object(map[string]cty.Type{
		"connection": cty.String,
		"scopes":     cty.Set(cty.String),
	}))
}

func tokenVaultBlockType() cty.Type {
	return cty.List(cty.Object(map[string]cty.Type{
		"credentials":  tokenVaultCredentialsType(),
		"ip_allowlist": cty.Set(cty.String),
		"grants":       tokenVaultGrantsType(),
	}))
}

func emptyTokenVaultBlock() cty.Value {
	return cty.ListVal([]cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"credentials":  cty.ListValEmpty(tokenVaultCredentialsType().ElementType()),
			"ip_allowlist": cty.SetValEmpty(cty.String),
			"grants":       cty.ListValEmpty(tokenVaultGrantsType().ElementType()),
		}),
	})
}

func marshalTokenVaultPrivilegedAccess(access *management.ClientTokenVaultPrivilegedAccess) string {
	payload, err := json.Marshal(struct {
		TokenVaultPrivilegedAccess *management.ClientTokenVaultPrivilegedAccess `json:"token_vault_privileged_access,omitempty"`
	}{access})
	if err != nil {
		panic(err)
	}
	return string(payload)
}

func TestExpandTokenVaultPrivilegedAccess(t *testing.T) {
	blockType := tokenVaultBlockType()

	t.Run("absent block emits nil, so token_vault_privileged_access is omitted entirely", func(t *testing.T) {
		data := newTokenVaultResourceData(t, cty.NullVal(blockType))

		access := expandTokenVaultPrivilegedAccess(data)
		require.Nil(t, access)
		assert.JSONEq(t, `{}`, marshalTokenVaultPrivilegedAccess(access))
	})

	t.Run("empty sub-fields emit empty arrays, not omitted keys", func(t *testing.T) {
		data := newTokenVaultResourceData(t, emptyTokenVaultBlock())

		access := expandTokenVaultPrivilegedAccess(data)
		require.NotNil(t, access)
		assert.JSONEq(
			t,
			`{"token_vault_privileged_access":{"credentials":[],"ip_allowlist":[],"grants":[]}}`,
			marshalTokenVaultPrivilegedAccess(access),
		)
	})

	t.Run("populated block emits all three sub-fields with values", func(t *testing.T) {
		data := newTokenVaultResourceData(t, cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"credentials": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"credential_type": cty.StringVal("public_key"),
						"pem":             cty.StringVal("-----BEGIN PUBLIC KEY-----"),
					}),
				}),
				"ip_allowlist": cty.SetVal([]cty.Value{cty.StringVal("10.0.0.1"), cty.StringVal("192.168.1.0/24")}),
				"grants": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"connection": cty.StringVal("google-oauth2"),
						"scopes":     cty.SetVal([]cty.Value{cty.StringVal("openid")}),
					}),
				}),
			}),
		}))

		access := expandTokenVaultPrivilegedAccess(data)
		require.NotNil(t, access)
		assert.JSONEq(t, `{"token_vault_privileged_access":{
			"credentials":[{"credential_type":"public_key","pem":"-----BEGIN PUBLIC KEY-----"}],
			"ip_allowlist":["10.0.0.1","192.168.1.0/24"],
			"grants":[{"connection":"google-oauth2","scopes":["openid"]}]
		}}`, marshalTokenVaultPrivilegedAccess(access))
	})

	t.Run("removal on update is signalled via isTokenVaultPrivilegedAccessNull, not the expand result", func(t *testing.T) {
		blockType := tokenVaultBlockType()
		data := updateTokenVaultResourceData(t, emptyTokenVaultBlock(), cty.NullVal(blockType))

		access := expandTokenVaultPrivilegedAccess(data)
		assert.Nil(t, access)
		assert.True(t, isTokenVaultPrivilegedAccessNull(data))
	})
}

func TestIsTokenVaultPrivilegedAccessNull(t *testing.T) {
	blockType := tokenVaultBlockType()

	t.Run("returns false on update when the block is unchanged and absent", func(t *testing.T) {
		data := updateTokenVaultResourceData(t, cty.NullVal(blockType), cty.NullVal(blockType))

		assert.False(t, isTokenVaultPrivilegedAccessNull(data))
	})

	t.Run("returns true when the block was removed on update", func(t *testing.T) {
		data := updateTokenVaultResourceData(t, emptyTokenVaultBlock(), cty.NullVal(blockType))

		assert.True(t, isTokenVaultPrivilegedAccessNull(data))
	})

	t.Run("returns false when the block is populated on update", func(t *testing.T) {
		data := updateTokenVaultResourceData(t, cty.NullVal(blockType), emptyTokenVaultBlock())

		assert.False(t, isTokenVaultPrivilegedAccessNull(data))
	})

	t.Run("returns true for a brand-new resource with the block absent", func(t *testing.T) {
		// fetchNullableFields (the only caller) never runs during create, so this
		// branch is unreachable in practice, but it matches every sibling
		// isXNull helper (isFedCMLoginNull, isTokenExchangeNull, ...).
		data := newTokenVaultResourceData(t, cty.NullVal(blockType))

		assert.True(t, isTokenVaultPrivilegedAccessNull(data))
	})
}
