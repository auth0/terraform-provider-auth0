package experimentfeatureflag

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestStringToTyped(t *testing.T) {
	testCases := []struct {
		name      string
		paramType string
		value     string
		wantErr   string
	}{
		{name: "boolean", paramType: "boolean", value: "true"},
		{name: "invalid boolean", paramType: "boolean", value: "yes", wantErr: `value "yes" is not a valid boolean`},
		{name: "number", paramType: "number", value: "12.5"},
		{name: "invalid number", paramType: "number", value: "abc", wantErr: `value "abc" is not a valid finite number`},
		{name: "NaN number", paramType: "number", value: "NaN", wantErr: `value "NaN" is not a valid finite number`},
		{name: "infinite number", paramType: "number", value: "Inf", wantErr: `value "Inf" is not a valid finite number`},
		{name: "object", paramType: "object", value: `{"a":1}`},
		{name: "malformed object", paramType: "object", value: `{bad}`, wantErr: `value "{bad}" is not a JSON object`},
		{name: "string JSON for object", paramType: "object", value: `"abc"`, wantErr: `value "\"abc\"" is not a JSON object`},
		{name: "array JSON for object", paramType: "object", value: `[1]`, wantErr: `value "[1]" is not a JSON object`},
		{name: "null for object", paramType: "object", value: `null`, wantErr: `value "null" is not a JSON object`},
		{name: "array", paramType: "array", value: `["a","b"]`},
		{name: "unquoted array items", paramType: "array", value: `[a,b]`, wantErr: `value "[a,b]" is not a JSON array`},
		{name: "object JSON for array", paramType: "array", value: `{"a":1}`, wantErr: `value "{\"a\":1}" is not a JSON array`},
		{name: "null for array", paramType: "array", value: `null`, wantErr: `value "null" is not a JSON array`},
		{name: "string", paramType: "string", value: "anything"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := stringToTyped(testCase.paramType, testCase.value)
			if testCase.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.EqualError(t, err, testCase.wantErr)
		})
	}
}

func TestTypedToStringNumbers(t *testing.T) {
	testCases := []struct {
		value float64
		want  string
	}{
		{value: 5, want: "5"},
		{value: 12.34, want: "12.34"},
		{value: 1000000, want: "1000000"},
		{value: 0.000001, want: "0.000001"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.want, func(t *testing.T) {
			got, err := typedToString(testCase.value)
			assert.NoError(t, err)
			assert.Equal(t, testCase.want, got)
		})
	}
}

func TestValidateParameters(t *testing.T) {
	parameterType := cty.Object(map[string]cty.Type{
		"name":        cty.String,
		"type":        cty.String,
		"value":       cty.String,
		"description": cty.String,
	})
	parameter := func(name, paramType, value cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"name":        name,
			"type":        paramType,
			"value":       value,
			"description": cty.NullVal(cty.String),
		})
	}

	t.Run("valid values pass", func(t *testing.T) {
		parameters := cty.SetVal([]cty.Value{
			parameter(cty.StringVal("flag"), cty.StringVal("boolean"), cty.StringVal("true")),
			parameter(cty.StringVal("obj"), cty.StringVal("object"), cty.StringVal(`{"a":1}`)),
		})
		assert.NoError(t, validateParameters(parameters))
	})

	t.Run("invalid values are all reported with their names", func(t *testing.T) {
		parameters := cty.SetVal([]cty.Value{
			parameter(cty.StringVal("flag"), cty.StringVal("boolean"), cty.StringVal("yes")),
			parameter(cty.StringVal("obj"), cty.StringVal("object"), cty.StringVal(`[1]`)),
		})
		err := validateParameters(parameters)
		assert.ErrorContains(t, err, `parameter "flag": value "yes" is not a valid boolean`)
		assert.ErrorContains(t, err, `parameter "obj": value "[1]" is not a JSON object`)
	})

	t.Run("unknown values are skipped", func(t *testing.T) {
		parameters := cty.SetVal([]cty.Value{
			parameter(cty.StringVal("flag"), cty.StringVal("boolean"), cty.UnknownVal(cty.String)),
			parameter(cty.StringVal("obj"), cty.UnknownVal(cty.String), cty.StringVal("yes")),
		})
		assert.NoError(t, validateParameters(parameters))
	})

	t.Run("unknown or null set is skipped", func(t *testing.T) {
		assert.NoError(t, validateParameters(cty.UnknownVal(cty.Set(parameterType))))
		assert.NoError(t, validateParameters(cty.NullVal(cty.Set(parameterType))))
	})

	t.Run("duplicate name is reported", func(t *testing.T) {
		parameters := cty.SetVal([]cty.Value{
			parameter(cty.StringVal("show_new_login"), cty.StringVal("boolean"), cty.StringVal("false")),
			parameter(cty.StringVal("show_new_login"), cty.StringVal("boolean"), cty.StringVal("true")),
			parameter(cty.StringVal("label"), cty.StringVal("string"), cty.StringVal("a")),
		})
		assert.EqualError(t, validateParameters(parameters),
			`parameter "show_new_login" is defined more than once; each parameter name must be unique`)
	})

	t.Run("unknown names are not treated as duplicates", func(t *testing.T) {
		parameters := cty.SetVal([]cty.Value{
			parameter(cty.UnknownVal(cty.String), cty.StringVal("boolean"), cty.StringVal("false")),
			parameter(cty.UnknownVal(cty.String), cty.StringVal("boolean"), cty.StringVal("true")),
		})
		assert.NoError(t, validateParameters(parameters))
	})
}

func TestExpandParametersRejectsDuplicateNames(t *testing.T) {
	parameters := schema.NewSet(hashFeatureFlagParameter, []interface{}{
		map[string]interface{}{"name": "show_new_login", "type": "boolean", "value": "false", "description": ""},
		map[string]interface{}{"name": "show_new_login", "type": "boolean", "value": "true", "description": ""},
	})

	_, err := expandParameters(parameters)
	assert.EqualError(t, err,
		`parameter "show_new_login" is defined more than once; each parameter name must be unique`)
}

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
