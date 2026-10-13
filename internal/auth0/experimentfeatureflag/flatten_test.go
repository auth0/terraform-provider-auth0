package experimentfeatureflag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
