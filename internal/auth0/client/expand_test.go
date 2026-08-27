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
