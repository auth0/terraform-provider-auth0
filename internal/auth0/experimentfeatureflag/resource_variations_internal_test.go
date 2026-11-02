package experimentfeatureflag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashVariation(t *testing.T) {
	variation := func(name, description string, overrides ...interface{}) map[string]interface{} {
		return map[string]interface{}{"name": name, "description": description, "overrides": overrides}
	}
	override := func(name, paramType, value string) map[string]interface{} {
		return map[string]interface{}{"name": name, "type": paramType, "value": value}
	}

	var testCases = []struct {
		name  string
		a, b  map[string]interface{}
		equal bool
	}{
		{
			name:  "override text forms of the same value",
			a:     variation("treatment", "", override("num", "number", "5")),
			b:     variation("treatment", "", override("num", "number", "5.0")),
			equal: true,
		},
		{
			name:  "overrides in a different order",
			a:     variation("treatment", "", override("a", "string", "x"), override("b", "string", "y")),
			b:     variation("treatment", "", override("b", "string", "y"), override("a", "string", "x")),
			equal: true,
		},
		{
			name:  "overrides as a set and as a list",
			a:     map[string]interface{}{"name": "treatment", "description": "", "overrides": newOverrideSet(override("num", "number", "5"))},
			b:     variation("treatment", "", override("num", "number", "5")),
			equal: true,
		},
		{
			name: "different names",
			a:    variation("control", ""),
			b:    variation("baseline", ""),
		},
		{
			name: "different descriptions",
			a:    variation("treatment", "before"),
			b:    variation("treatment", "after"),
		},
		{
			name: "different override values",
			a:    variation("treatment", "", override("num", "number", "5")),
			b:    variation("treatment", "", override("num", "number", "6")),
		},
		{
			name: "an override added",
			a:    variation("treatment", ""),
			b:    variation("treatment", "", override("num", "number", "5")),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.equal, hashVariation(testCase.a) == hashVariation(testCase.b))
		})
	}
}

func TestHashOverride(t *testing.T) {
	override := func(name, paramType, value string) map[string]interface{} {
		return map[string]interface{}{"name": name, "type": paramType, "value": value}
	}

	var testCases = []struct {
		name  string
		a, b  map[string]interface{}
		equal bool
	}{
		{
			name:  "number text forms of the same value",
			a:     override("num", "number", "5"),
			b:     override("num", "number", "5.0"),
			equal: true,
		},
		{
			name:  "JSON with different spacing",
			a:     override("obj", "object", `{"label":"B","enabled":true}`),
			b:     override("obj", "object", `{ "enabled": true, "label": "B" }`),
			equal: true,
		},
		{
			name:  "boolean text forms of the same value",
			a:     override("bool", "boolean", "false"),
			b:     override("bool", "boolean", "False"),
			equal: true,
		},
		{
			name:  "different values",
			a:     override("num", "number", "5"),
			b:     override("num", "number", "6"),
			equal: false,
		},
		{
			name:  "different names",
			a:     override("a", "string", "x"),
			b:     override("b", "string", "x"),
			equal: false,
		},
		{
			name:  "same text with a different type",
			a:     override("v", "string", "true"),
			b:     override("v", "boolean", "true"),
			equal: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.equal, hashOverride(testCase.a) == hashOverride(testCase.b))
		})
	}
}
