package experimentfeatureflag

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func newOverrideSet(overrides ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, 0, len(overrides))
	for _, override := range overrides {
		items = append(items, override)
	}
	return schema.NewSet(hashOverride, items)
}

func TestExpandOverrides(t *testing.T) {
	t.Run("it returns an empty map for a control variation", func(t *testing.T) {
		overrides, err := expandOverrides(newOverrideSet())

		assert.NoError(t, err)
		assert.NotNil(t, overrides)
		assert.Empty(t, overrides)
	})

	t.Run("it converts each value to its type", func(t *testing.T) {
		overrides, err := expandOverrides(newOverrideSet(
			map[string]interface{}{"name": "str", "type": "string", "value": "Continue"},
			map[string]interface{}{"name": "bool", "type": "boolean", "value": "true"},
			map[string]interface{}{"name": "num", "type": "number", "value": "1.5"},
			map[string]interface{}{"name": "obj", "type": "object", "value": `{"label":"B"}`},
			map[string]interface{}{"name": "arr", "type": "array", "value": `[1,2]`},
		))

		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{
			"str":  "Continue",
			"bool": true,
			"num":  1.5,
			"obj":  map[string]interface{}{"label": "B"},
			"arr":  []interface{}{float64(1), float64(2)},
		}, overrides)
	})

	t.Run("it names the override in a conversion error", func(t *testing.T) {
		_, err := expandOverrides(newOverrideSet(
			map[string]interface{}{"name": "num", "type": "number", "value": "many"},
		))

		assert.EqualError(t, err, `override "num": value "many" is not a valid number`)
	})
}

func TestExpandOverridesUpdate(t *testing.T) {
	str := map[string]interface{}{"name": "str", "type": "string", "value": "Continue"}
	num := map[string]interface{}{"name": "num", "type": "number", "value": "5"}
	boolean := map[string]interface{}{"name": "bool", "type": "boolean", "value": "true"}

	t.Run("it sends the full set when no override is removed", func(t *testing.T) {
		overrides, err := expandOverridesUpdate(
			newOverrideSet(str),
			newOverrideSet(str, num),
		)

		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"str": "Continue", "num": float64(5)}, overrides)
	})

	t.Run("it sends a removed override as null", func(t *testing.T) {
		overrides, err := expandOverridesUpdate(
			newOverrideSet(str, num),
			newOverrideSet(str),
		)

		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"str": "Continue", "num": nil}, overrides)
	})

	t.Run("it sends the old name as null and the new name with its value on a rename", func(t *testing.T) {
		overrides, err := expandOverridesUpdate(
			newOverrideSet(str, num),
			newOverrideSet(str, boolean),
		)

		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"str": "Continue", "num": nil, "bool": true}, overrides)
	})

	t.Run("it sends every old override as null for a control variation", func(t *testing.T) {
		overrides, err := expandOverridesUpdate(
			newOverrideSet(str, num),
			newOverrideSet(),
		)

		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"str": nil, "num": nil}, overrides)
	})

	t.Run("it returns a conversion error", func(t *testing.T) {
		_, err := expandOverridesUpdate(
			newOverrideSet(str),
			newOverrideSet(map[string]interface{}{"name": "num", "type": "number", "value": "many"}),
		)

		assert.EqualError(t, err, `override "num": value "many" is not a valid number`)
	})
}
