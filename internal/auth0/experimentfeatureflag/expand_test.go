package experimentfeatureflag

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newOverrideSet(overrides ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, 0, len(overrides))
	for _, override := range overrides {
		items = append(items, override)
	}
	return schema.NewSet(hashOverride, items)
}

func newVariation(name, description string, overrides ...map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":        name,
		"description": description,
		"overrides":   newOverrideSet(overrides...),
	}
}

func TestExpandVariation(t *testing.T) {
	t.Run("it sends an empty overrides map and no description for a control variation", func(t *testing.T) {
		request, err := expandVariation(newVariation("control", ""))

		require.NoError(t, err)
		assert.Equal(t, "control", request.Name)
		assert.Nil(t, request.Description)
		assert.NotNil(t, request.Overrides)
		assert.Empty(t, request.Overrides)
	})

	t.Run("it sends the description and the typed overrides", func(t *testing.T) {
		request, err := expandVariation(newVariation("treatment", "Redesigned checkout",
			map[string]interface{}{"name": "num", "type": "number", "value": "5"},
		))

		require.NoError(t, err)
		assert.Equal(t, "Redesigned checkout", *request.Description)
		assert.Equal(t, map[string]interface{}{"num": float64(5)}, request.Overrides)
	})

	t.Run("it returns a conversion error", func(t *testing.T) {
		_, err := expandVariation(newVariation("treatment", "",
			map[string]interface{}{"name": "num", "type": "number", "value": "many"},
		))

		assert.EqualError(t, err, `override "num": value "many" is not a valid number`)
	})
}

func TestExpandVariationUpdate(t *testing.T) {
	str := map[string]interface{}{"name": "str", "type": "string", "value": "Continue"}
	num := map[string]interface{}{"name": "num", "type": "number", "value": "5"}

	var testCases = []struct {
		name         string
		oldVariation map[string]interface{}
		newVariation map[string]interface{}
		expectedBody string
	}{
		{
			name:         "it sends a changed description",
			oldVariation: newVariation("treatment", "before", str),
			newVariation: newVariation("treatment", "after", str),
			expectedBody: `{"description":"after"}`,
		},
		{
			name:         "it sends a removed description as null",
			oldVariation: newVariation("treatment", "before", str),
			newVariation: newVariation("treatment", "", str),
			expectedBody: `{"description":null}`,
		},
		{
			name:         "it sends the new overrides and a removed override as null",
			oldVariation: newVariation("treatment", "", str, num),
			newVariation: newVariation("treatment", "", str),
			expectedBody: `{"overrides":{"str":"Continue","num":null}}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			request, changed, err := expandVariationUpdate(testCase.oldVariation, testCase.newVariation)
			require.NoError(t, err)
			assert.True(t, changed)

			body, err := json.Marshal(request)
			require.NoError(t, err)
			assert.JSONEq(t, testCase.expectedBody, string(body))
		})
	}

	t.Run("it reports no change for the same value in other text", func(t *testing.T) {
		_, changed, err := expandVariationUpdate(
			newVariation("treatment", "", num),
			newVariation("treatment", "", map[string]interface{}{"name": "num", "type": "number", "value": "5.0"}),
		)

		require.NoError(t, err)
		assert.False(t, changed)
	})

	t.Run("it returns a conversion error", func(t *testing.T) {
		_, _, err := expandVariationUpdate(
			newVariation("treatment", "", num),
			newVariation("treatment", "", map[string]interface{}{"name": "num", "type": "number", "value": "many"}),
		)

		assert.EqualError(t, err, `override "num": value "many" is not a valid number`)
	})
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
