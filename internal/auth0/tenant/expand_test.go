package tenant

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// resourceDataWithRawConfig builds a *schema.ResourceData whose GetRawConfig()
// reflects the given country_codes value, mirroring how Terraform populates
// RawConfig on a real plan/apply (TestResourceDataRaw alone leaves it null).
func resourceDataWithRawConfig(t *testing.T, countryCodes cty.Value) *schema.ResourceData {
	t.Helper()

	sm := schema.InternalMap(NewResource().Schema)
	impliedType := sm.CoreConfigSchema().ImpliedType()

	attrTypes := impliedType.AttributeTypes()
	vals := make(map[string]cty.Value, len(attrTypes))
	for name, ty := range attrTypes {
		vals[name] = cty.NullVal(ty)
	}
	vals["country_codes"] = countryCodes

	data, err := sm.Data(nil, &terraform.InstanceDiff{RawConfig: cty.ObjectVal(vals)})
	assert.NoError(t, err)

	data.MarkNewResource()

	return data
}

func TestExpandTenantCountryCodes(t *testing.T) {
	countryCodesType := cty.List(cty.Object(map[string]cty.Type{
		"list": cty.Set(cty.String),
		"mode": cty.String,
	}))

	t.Run("it returns nil when country_codes block is absent", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, cty.NullVal(countryCodesType))

		countryCodes := expandTenantCountryCodes(data)

		assert.Nil(t, countryCodes)
	})

	t.Run("it returns the populated struct when country_codes block is configured", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"list": cty.SetVal([]cty.Value{cty.StringVal("US"), cty.StringVal("CA")}),
				"mode": cty.StringVal("allow"),
			}),
		}))

		countryCodes := expandTenantCountryCodes(data)

		assert.NotNil(t, countryCodes)
		assert.Equal(t, "allow", countryCodes.Mode)
		assert.ElementsMatch(t, []string{"US", "CA"}, countryCodes.List)
	})
}

func TestIsCountryCodesNull(t *testing.T) {
	countryCodesType := cty.List(cty.Object(map[string]cty.Type{
		"list": cty.Set(cty.String),
		"mode": cty.String,
	}))

	t.Run("it returns true when country_codes block is absent", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, cty.NullVal(countryCodesType))

		assert.True(t, isCountryCodesNull(data))
	})

	t.Run("it returns false when country_codes block is configured", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, cty.ListVal([]cty.Value{
			cty.ObjectVal(map[string]cty.Value{
				"list": cty.SetVal([]cty.Value{cty.StringVal("US")}),
				"mode": cty.StringVal("deny"),
			}),
		}))

		assert.False(t, isCountryCodesNull(data))
	})
}
