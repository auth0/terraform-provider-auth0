package experiment

import (
	"testing"

	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
)

func allocationConfig(overrides map[string]cty.Value) cty.Value {
	attributes := map[string]cty.Value{
		"variation_id": cty.StringVal("var_control"),
		"is_control":   cty.True,
		"weight":       cty.NullVal(cty.Number),
		"segment_id":   cty.NullVal(cty.String),
		"priority":     cty.NullVal(cty.Number),
		"is_fallback":  cty.NullVal(cty.Bool),
	}
	for key, value := range overrides {
		attributes[key] = value
	}

	return cty.ObjectVal(attributes)
}

func TestExpandExperimentAllocations(t *testing.T) {
	t.Run("it returns nil when allocations are not configured", func(t *testing.T) {
		allocations, err := expandExperimentAllocations(cty.NullVal(cty.List(allocationConfig(nil).Type())))

		assert.NoError(t, err)
		assert.Nil(t, allocations)
	})

	t.Run("it sends only the configured percentage fields", func(t *testing.T) {
		allocations, err := expandExperimentAllocations(cty.ListVal([]cty.Value{
			allocationConfig(map[string]cty.Value{"weight": cty.NumberIntVal(70)}),
			allocationConfig(map[string]cty.Value{
				"variation_id": cty.StringVal("var_treatment"),
				"is_control":   cty.False,
				"weight":       cty.NumberIntVal(30),
			}),
		}))

		assert.NoError(t, err)
		assert.Equal(t, []*management.AllocationRequestItem{
			{VariationID: "var_control", IsControl: true, Weight: intPointer(70)},
			{VariationID: "var_treatment", IsControl: false, Weight: intPointer(30)},
		}, allocations)
	})

	t.Run("it keeps the order of segment and fallback allocations", func(t *testing.T) {
		allocations, err := expandExperimentAllocations(cty.ListVal([]cty.Value{
			allocationConfig(map[string]cty.Value{
				"segment_id": cty.StringVal("seg_us"),
				"priority":   cty.NumberIntVal(1),
			}),
			allocationConfig(map[string]cty.Value{
				"variation_id": cty.StringVal("var_treatment"),
				"is_control":   cty.False,
				"is_fallback":  cty.True,
			}),
		}))

		assert.NoError(t, err)
		assert.Equal(t, []*management.AllocationRequestItem{
			{VariationID: "var_control", IsControl: true, SegmentID: stringPointer("seg_us"), Priority: intPointer(1)},
			{VariationID: "var_treatment", IsControl: false, IsFallback: boolPointer(true)},
		}, allocations)
	})
}

func TestExpandExperimentLevels(t *testing.T) {
	t.Run("it returns nil when levels are not configured", func(t *testing.T) {
		assert.Nil(t, expandExperimentLevels(nil))
		assert.Nil(t, expandExperimentLevels([]interface{}{}))
	})

	t.Run("it keeps the configured order", func(t *testing.T) {
		assert.Equal(t, []int{10, 50, 100}, expandExperimentLevels([]interface{}{10, 50, 100}))
	})
}

func intPointer(value int) *int {
	return &value
}

func stringPointer(value string) *string {
	return &value
}

func boolPointer(value bool) *bool {
	return &value
}
