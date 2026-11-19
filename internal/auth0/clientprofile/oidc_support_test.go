package clientprofile

import (
	"encoding/json"
	"testing"

	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// oidcSupportConfig builds the oidc_support block with its single default_value item.
func oidcSupportConfig(isAllowed, allowAll, scopes cty.Value) cty.Value {
	return wrap(cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
		"is_allowed":       isAllowed,
		"allow_all_scopes": allowAll,
		"allowed_scopes":   scopes,
	})}))
}

func TestExpandOidcSupport(t *testing.T) {
	nullBool := cty.NullVal(cty.Bool)
	nullScopes := cty.NullVal(cty.Set(cty.String))

	expand := func(block cty.Value) string {
		configuration := expandClientConfiguration(rawConfig(map[string]cty.Value{"oidc_support": block}), true)
		require.NotNil(t, configuration.OidcSupport)

		return marshal(t, configuration.OidcSupport)
	}

	t.Run("absent block is omitted", func(t *testing.T) {
		configuration := expandClientConfiguration(rawConfig(nil), false)
		assert.Nil(t, configuration.OidcSupport)
	})

	t.Run("disabled", func(t *testing.T) {
		assert.JSONEq(t, `{"default_value":{"is_allowed":false}}`, expand(oidcSupportConfig(cty.False, nullBool, nullScopes)))
	})

	t.Run("scopes", func(t *testing.T) {
		scopes := cty.SetVal([]cty.Value{cty.StringVal("profile"), cty.StringVal("email")})
		assert.JSONEq(t,
			`{"default_value":{"is_allowed":true,"allowed_scopes":["email","profile"]}}`,
			expand(oidcSupportConfig(cty.True, cty.False, scopes)),
		)
	})

	t.Run("all scopes", func(t *testing.T) {
		assert.JSONEq(t,
			`{"default_value":{"is_allowed":true,"allow_all_scopes":true}}`,
			expand(oidcSupportConfig(cty.True, cty.True, nullScopes)),
		)
	})
}

func TestFlattenOidcSupport(t *testing.T) {
	flatten := func(body string) []interface{} {
		var oidcSupport management.ClientProfileClientConfigurationOidcSupport
		require.NoError(t, json.Unmarshal([]byte(body), &oidcSupport))

		return flattenOidcSupport(&oidcSupport)
	}

	assert.Nil(t, flattenOidcSupport(nil))
	assert.Nil(t, flatten(`{}`))

	assert.Equal(t,
		flattenEnvelope([]interface{}{map[string]interface{}{"is_allowed": false}}),
		flatten(`{"default_value":{"is_allowed":false}}`),
	)
	assert.Equal(t,
		flattenEnvelope([]interface{}{map[string]interface{}{
			"is_allowed":     true,
			"allowed_scopes": []interface{}{"profile", "phone"},
		}}),
		flatten(`{"default_value":{"is_allowed":true,"allowed_scopes":["profile","phone"],"extra":1}}`),
	)
	assert.Equal(t,
		flattenEnvelope([]interface{}{map[string]interface{}{
			"is_allowed":       true,
			"allow_all_scopes": true,
		}}),
		flatten(`{"default_value":{"is_allowed":true,"allow_all_scopes":true}}`),
	)
}

func TestCheckClientProfileConfigOidcSupport(t *testing.T) {
	nullBool := cty.NullVal(cty.Bool)
	nullScopes := cty.NullVal(cty.Set(cty.String))
	profile := cty.SetVal([]cty.Value{cty.StringVal("profile")})

	check := func(block cty.Value) error {
		return checkClientProfileConfig(rawConfig(map[string]cty.Value{"oidc_support": block}))
	}

	assert.NoError(t, check(oidcSupportConfig(cty.False, nullBool, nullScopes)))
	assert.NoError(t, check(oidcSupportConfig(cty.True, nullBool, profile)))
	assert.NoError(t, check(oidcSupportConfig(cty.True, cty.True, nullScopes)))
	assert.NoError(t, check(oidcSupportConfig(cty.UnknownVal(cty.Bool), nullBool, nullScopes)))
	assert.Error(t, check(oidcSupportConfig(cty.False, nullBool, profile)))
	assert.Error(t, check(oidcSupportConfig(cty.True, cty.True, profile)))
	assert.Error(t, check(oidcSupportConfig(cty.True, nullBool, nullScopes)))
}
