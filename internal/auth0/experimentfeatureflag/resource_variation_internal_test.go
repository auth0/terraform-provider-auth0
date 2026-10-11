package experimentfeatureflag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
