package client

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExpandMyOrganizationConfigurationThirdPartyClientAccess_AlwaysSendsDefaultValue pins the wire
// format confirmed live: the API rejects the object with "Missing required property: default_value"
// if it's absent, even though default_value is never user-settable (Constraint 4 — the API
// unconditionally rejects any value other than "block"). Expand must always inject the literal
// "block" value whenever allowed_values is non-empty.
func TestExpandMyOrganizationConfigurationThirdPartyClientAccess_AlwaysSendsDefaultValue(t *testing.T) {
	tpca := expandMyOrganizationConfigurationThirdPartyClientAccess(cty.ListVal([]cty.Value{
		cty.ObjectVal(map[string]cty.Value{
			"default_value":  cty.StringVal(""),
			"allowed_values": cty.ListVal([]cty.Value{cty.StringVal("block")}),
		}),
	}))

	require.NotNil(t, tpca)
	require.NotNil(t, tpca.DefaultValue, "a nil default_value would omit the key and 400")
	assert.Equal(t, "block", *tpca.DefaultValue)
	require.NotNil(t, tpca.AllowedValues)
	assert.Equal(t, []string{"block"}, *tpca.AllowedValues)
}

func TestExpandMyOrganizationConfigurationThirdPartyClientAccess_NullConfigIsNil(t *testing.T) {
	tpca := expandMyOrganizationConfigurationThirdPartyClientAccess(cty.NullVal(cty.List(cty.Object(map[string]cty.Type{
		"default_value":  cty.String,
		"allowed_values": cty.List(cty.String),
	}))))

	assert.Nil(t, tpca)
}

func oidcSupportBlock(isAllowed cty.Value, allowAll cty.Value, scopes cty.Value) cty.Value {
	return cty.ObjectVal(map[string]cty.Value{
		"is_allowed":       isAllowed,
		"allow_all_scopes": allowAll,
		"allowed_scopes":   scopes,
	})
}

func TestExpandOIDCSupportBlock(t *testing.T) {
	nullBool := cty.NullVal(cty.Bool)
	nullScopes := cty.NullVal(cty.Set(cty.String))

	t.Run("absent block is omitted", func(t *testing.T) {
		assert.Nil(t, expandOIDCSupportBlock(cty.NullVal(cty.DynamicPseudoType)))
	})

	t.Run("disabled sends is_allowed only", func(t *testing.T) {
		result := expandOIDCSupportBlock(oidcSupportBlock(cty.False, nullBool, nullScopes))
		require.NotNil(t, result)
		assert.JSONEq(t, `{"is_allowed":false}`, result.String())
	})

	t.Run("scopes omit allow_all_scopes", func(t *testing.T) {
		scopes := cty.SetVal([]cty.Value{cty.StringVal("profile"), cty.StringVal("email")})
		result := expandOIDCSupportBlock(oidcSupportBlock(cty.True, cty.False, scopes))
		require.NotNil(t, result)
		assert.JSONEq(t, `{"is_allowed":true,"allowed_scopes":["email","profile"]}`, result.String())
	})

	t.Run("all scopes", func(t *testing.T) {
		result := expandOIDCSupportBlock(oidcSupportBlock(cty.True, cty.True, nullScopes))
		require.NotNil(t, result)
		assert.JSONEq(t, `{"is_allowed":true,"allow_all_scopes":true}`, result.String())
	})

	t.Run("empty scopes set is omitted", func(t *testing.T) {
		result := expandOIDCSupportBlock(oidcSupportBlock(cty.True, nullBool, cty.SetValEmpty(cty.String)))
		require.NotNil(t, result)
		assert.JSONEq(t, `{"is_allowed":true}`, result.String())
	})
}
