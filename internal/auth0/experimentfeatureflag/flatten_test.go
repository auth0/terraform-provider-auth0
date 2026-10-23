package experimentfeatureflag

import (
	"testing"

	"github.com/auth0/go-auth0"
	management "github.com/auth0/go-auth0/v3/management"
	"github.com/stretchr/testify/assert"
)

func TestFlattenVariations(t *testing.T) {
	t.Run("it returns empty values when the feature flag has no variations", func(t *testing.T) {
		variations, ids, err := flattenVariations(nil)

		assert.NoError(t, err)
		assert.Empty(t, variations)
		assert.NotNil(t, ids)
		assert.Empty(t, ids)
	})

	t.Run("it returns the variations and a map of names to IDs", func(t *testing.T) {
		variations, ids, err := flattenVariations([]*management.Variation{
			{ID: "var_1", Name: "control", Overrides: map[string]interface{}{}},
			{ID: "var_2", Name: "treatment", Description: auth0.String("Redesigned"), Overrides: map[string]interface{}{"num": float64(5)}},
		})

		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"control": "var_1", "treatment": "var_2"}, ids)
		assert.Equal(t, []interface{}{
			map[string]interface{}{"name": "control", "description": "", "overrides": []interface{}{}},
			map[string]interface{}{
				"name":        "treatment",
				"description": "Redesigned",
				"overrides": []interface{}{
					map[string]interface{}{"name": "num", "type": "number", "value": "5"},
				},
			},
		}, variations)
	})

	t.Run("it names the variation in an override error", func(t *testing.T) {
		_, _, err := flattenVariations([]*management.Variation{
			{ID: "var_1", Name: "treatment", Overrides: map[string]interface{}{"empty": nil}},
		})

		assert.EqualError(t, err, `variation "treatment": override "empty" has an unsupported value type <nil>`)
	})
}

func TestFlattenOverrides(t *testing.T) {
	t.Run("it returns an empty list for a control variation", func(t *testing.T) {
		overrides, err := flattenOverrides(nil)

		assert.NoError(t, err)
		assert.Empty(t, overrides)
	})

	t.Run("it derives each type from the JSON value", func(t *testing.T) {
		overrides, err := flattenOverrides(map[string]interface{}{
			"str":  "Continue",
			"bool": false,
			"num":  float64(5),
			"obj":  map[string]interface{}{"label": "B"},
			"arr":  []interface{}{float64(1), float64(2)},
		})

		assert.NoError(t, err)
		assert.ElementsMatch(t, []interface{}{
			map[string]interface{}{"name": "str", "type": "string", "value": "Continue"},
			map[string]interface{}{"name": "bool", "type": "boolean", "value": "false"},
			map[string]interface{}{"name": "num", "type": "number", "value": "5"},
			map[string]interface{}{"name": "obj", "type": "object", "value": `{"label":"B"}`},
			map[string]interface{}{"name": "arr", "type": "array", "value": `[1,2]`},
		}, overrides)
	})

	t.Run("it rejects a value with no parameter type", func(t *testing.T) {
		_, err := flattenOverrides(map[string]interface{}{"empty": nil})

		assert.EqualError(t, err, `override "empty" has an unsupported value type <nil>`)
	})
}
