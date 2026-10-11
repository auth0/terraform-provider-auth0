package experimentfeatureflag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashFeatureFlagParameter(t *testing.T) {
	parameter := func(name, paramType, value, description string) map[string]interface{} {
		return map[string]interface{}{
			"name":        name,
			"type":        paramType,
			"value":       value,
			"description": description,
		}
	}

	testCases := []struct {
		name  string
		a, b  map[string]interface{}
		equal bool
	}{
		{
			name:  "number text forms of the same value",
			a:     parameter("num", "number", "5", ""),
			b:     parameter("num", "number", "5.0", ""),
			equal: true,
		},
		{
			name:  "JSON with different key order and spacing",
			a:     parameter("obj", "object", `{"label":"B","enabled":true}`, ""),
			b:     parameter("obj", "object", `{ "enabled": true, "label": "B" }`, ""),
			equal: true,
		},
		{
			name:  "array with different spacing",
			a:     parameter("arr", "array", `[1,2,3]`, ""),
			b:     parameter("arr", "array", `[1, 2, 3]`, ""),
			equal: true,
		},
		{
			name:  "boolean text forms of the same value",
			a:     parameter("bool", "boolean", "false", ""),
			b:     parameter("bool", "boolean", "False", ""),
			equal: true,
		},
		{
			name: "different values",
			a:    parameter("num", "number", "5", ""),
			b:    parameter("num", "number", "6", ""),
		},
		{
			name: "different names",
			a:    parameter("a", "string", "x", ""),
			b:    parameter("b", "string", "x", ""),
		},
		{
			name: "same text with a different type",
			a:    parameter("v", "string", "true", ""),
			b:    parameter("v", "boolean", "true", ""),
		},
		{
			name: "different descriptions",
			a:    parameter("v", "string", "x", "before"),
			b:    parameter("v", "string", "x", "after"),
		},
		{
			name: "invalid JSON uses its original text",
			a:    parameter("obj", "object", `{bad}`, ""),
			b:    parameter("obj", "object", `{ bad }`, ""),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.equal,
				hashFeatureFlagParameter(testCase.a) == hashFeatureFlagParameter(testCase.b))
		})
	}
}
