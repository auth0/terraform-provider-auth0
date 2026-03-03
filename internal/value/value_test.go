package value

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBool(t *testing.T) {
	t.Run("it returns nil when given a null value", func(t *testing.T) {
		actual := Bool(cty.NullVal(cty.Bool))
		assert.Nil(t, actual)
	})

	t.Run("it returns true when given a true bool value", func(t *testing.T) {
		actual := Bool(cty.BoolVal(true))
		require.NotNil(t, actual)
		assert.True(t, *actual)
	})

	t.Run("it returns false when given a false bool value", func(t *testing.T) {
		actual := Bool(cty.BoolVal(false))
		require.NotNil(t, actual)
		assert.False(t, *actual)
	})
}

func TestBoolPtrToString(t *testing.T) {
	t.Run("it returns 'null' for nil pointer", func(t *testing.T) {
		var ptr *bool
		assert.Equal(t, "null", BoolPtrToString(ptr))
	})

	t.Run("it returns 'true' for true pointer", func(t *testing.T) {
		val := true
		assert.Equal(t, "true", BoolPtrToString(&val))
	})

	t.Run("it returns 'false' for false pointer", func(t *testing.T) {
		val := false
		assert.Equal(t, "false", BoolPtrToString(&val))
	})
}

func TestBoolPtr(t *testing.T) {
	t.Run("it returns true pointer for 'true' string", func(t *testing.T) {
		result := BoolPtr("true")
		require.NotNil(t, result)
		assert.True(t, *result)
	})

	t.Run("it returns false pointer for 'false' string", func(t *testing.T) {
		result := BoolPtr("false")
		require.NotNil(t, result)
		assert.False(t, *result)
	})

	t.Run("it returns nil for any other string", func(t *testing.T) {
		result := BoolPtr("foo")
		assert.Nil(t, result)
	})

	t.Run("it returns nil for non-string input", func(t *testing.T) {
		result := BoolPtr(123)
		assert.Nil(t, result)
	})

	t.Run("it returns nil for nil input", func(t *testing.T) {
		result := BoolPtr(nil)
		assert.Nil(t, result)
	})
}

func TestString(t *testing.T) {
	t.Run("it returns nil when given a null value", func(t *testing.T) {
		actual := String(cty.NullVal(cty.String))
		assert.Nil(t, actual)
	})

	t.Run("it returns an empty string when given an empty string value", func(t *testing.T) {
		actual := String(cty.StringVal(""))
		require.NotNil(t, actual)
		assert.Equal(t, "", *actual)
	})

	t.Run("it returns a string when given a string value", func(t *testing.T) {
		expected := "foo bar"
		actual := String(cty.StringVal(expected))
		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})
}

func TestInt(t *testing.T) {
	t.Run("it returns nil when given a null value", func(t *testing.T) {
		actual := Int(cty.NullVal(cty.Number))
		assert.Nil(t, actual)
	})

	t.Run("it returns 0 when given a 0 value", func(t *testing.T) {
		var expected int64
		actual := Int(cty.NumberIntVal(expected))
		require.NotNil(t, actual)
		assert.Equal(t, int(expected), *actual)
	})

	t.Run("it returns a negative integer when given a negative integer value", func(t *testing.T) {
		var expected int64 = -math.MaxInt64
		actual := Int(cty.NumberIntVal(expected))
		require.NotNil(t, actual)
		assert.Equal(t, int(expected), *actual)
	})

	t.Run("it returns a positive integer when given a positive integer value", func(t *testing.T) {
		var expected int64 = math.MaxInt64
		actual := Int(cty.NumberIntVal(expected))
		require.NotNil(t, actual)
		assert.Equal(t, int(expected), *actual)
	})
}

func TestFloat64(t *testing.T) {
	t.Run("it returns nil when given a null value", func(t *testing.T) {
		actual := Float64(cty.NullVal(cty.Number))
		assert.Nil(t, actual)
	})

	t.Run("it returns 0 when given a 0 value", func(t *testing.T) {
		var expected float64
		actual := Float64(cty.NumberFloatVal(expected))
		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})

	t.Run("it returns a negative float when given a negative float value", func(t *testing.T) {
		var expected = -math.MaxFloat64
		actual := Float64(cty.NumberFloatVal(expected))
		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})

	t.Run("it returns a positive float when given a positive float value", func(t *testing.T) {
		expected := math.MaxFloat64
		actual := Float64(cty.NumberFloatVal(expected))
		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})
}

// Time evaluates the typed value of the value
// and coerces to a pointer of a string, which
// is then converted to a `time.Time` according
// to ISO 3339 (ISO 8601 is largely the same in
// common use cases, see https://ijmacd.github.io/rfc3339-iso8601/
// for differences).
func TestTime(t *testing.T) {
	t.Run("it returns nil when given a null value", func(t *testing.T) {
		actual := Time(cty.NullVal(cty.String))
		assert.Nil(t, actual)
	})

	t.Run("it returns the correct Time when given a correctly formatted string", func(t *testing.T) {
		checkTime, _ := time.Parse(time.RFC3339, "2024-09-06T20:00:00Z")
		actual := Time(cty.StringVal("2024-09-06T20:00:00Z"))
		assert.Equal(t, checkTime, *actual)
	})
}

func TestStrings(t *testing.T) {
	t.Run("it returns nil when given a null value", func(t *testing.T) {
		actual := Strings(cty.NilVal)
		assert.Nil(t, actual)
	})

	t.Run("it returns an empty slice when given an empty list value", func(t *testing.T) {
		actual := Strings(cty.ListValEmpty(cty.String))
		require.NotNil(t, actual)
		assert.Equal(t, []string{}, *actual)
	})

	t.Run("it returns a slice of strings when given a string list value", func(t *testing.T) {
		expected := []string{
			"localhost/logout",
			"https://app.domain.com/logout",
		}

		var testInput []cty.Value
		for _, value := range expected {
			testInput = append(testInput, cty.StringVal(value))
		}

		actual := Strings(cty.ListVal(testInput))
		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})
}

func TestMap(t *testing.T) {
	t.Run("it returns nil when given a null value", func(t *testing.T) {
		actual := Map(cty.NilVal)
		assert.Nil(t, actual)
	})

	t.Run("it returns an empty map when given an empty map value", func(t *testing.T) {
		actual := Map(cty.MapValEmpty(cty.String))
		require.NotNil(t, actual)
		assert.Equal(t, map[string]interface{}{}, actual)
	})

	t.Run("it returns a map when given a map value", func(t *testing.T) {
		expected := map[string]interface{}{
			"logout": "https://app.domain.com/logout",
			"login":  "https://app.domain.com/login",
		}

		testInput := make(map[string]cty.Value)
		for key, value := range expected {
			testInput[key] = cty.StringVal(value.(string))
		}

		actual := Map(cty.MapVal(testInput))
		require.NotNil(t, actual)
		assert.Equal(t, expected, actual)
	})

	t.Run("it ignores null values", func(t *testing.T) {
		expected := map[string]interface{}{
			"logout": "https://app.domain.com/logout",
		}

		testInput := map[string]cty.Value{
			"logout": cty.StringVal("https://app.domain.com/logout"),
			"login":  cty.NullVal(cty.String),
		}

		actual := Map(cty.MapVal(testInput))
		require.NotNil(t, actual)
		assert.Equal(t, expected, actual)
	})
}

func TestMapOfStrings(t *testing.T) {
	t.Run("it returns nil when given a null value", func(t *testing.T) {
		actual := MapOfStrings(cty.NilVal)
		assert.Nil(t, actual)
	})

	t.Run("it returns an empty map when given an empty map value", func(t *testing.T) {
		actual := MapOfStrings(cty.MapValEmpty(cty.String))
		require.NotNil(t, actual)
		assert.Equal(t, map[string]string{}, *actual)
	})

	t.Run("it returns a map when given a map value", func(t *testing.T) {
		expected := map[string]string{
			"logout": "https://app.domain.com/logout",
			"login":  "https://app.domain.com/login",
		}
		testInput := make(map[string]cty.Value)
		for key, value := range expected {
			testInput[key] = cty.StringVal(value)
		}

		actual := MapOfStrings(cty.MapVal(testInput))
		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})

	t.Run("it ignores null values", func(t *testing.T) {
		expected := map[string]string{
			"logout": "https://app.domain.com/logout",
		}

		testInput := map[string]cty.Value{
			"logout": cty.StringVal("https://app.domain.com/logout"),
			"login":  cty.NullVal(cty.String),
		}

		actual := MapOfStrings(cty.MapVal(testInput))
		require.NotNil(t, actual)
		assert.Equal(t, expected, *actual)
	})
}

func TestMapFromJSON(t *testing.T) {
	t.Run("it returns nil when given a null value", func(t *testing.T) {
		actual, err := MapFromJSON(cty.NilVal)
		assert.NoError(t, err)
		assert.Nil(t, actual)
	})

	t.Run("it returns an empty map when given an empty string value", func(t *testing.T) {
		actual, err := MapFromJSON(cty.NullVal(cty.String))
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}(nil), actual)
	})

	t.Run("it returns an error when given an invalid json value", func(t *testing.T) {
		invalidJSON := "[not valid json"
		actual, err := MapFromJSON(cty.StringVal(invalidJSON))
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	t.Run("it returns a map when given a valid json value", func(t *testing.T) {
		payload := map[string]interface{}{
			"bool": true,
			"int":  5,
			"map": map[string]interface{}{
				"nested": true,
				"slice":  []interface{}{1, 2, 3},
				"string": "foo",
			},
		}
		expected, err := json.Marshal(&payload)
		require.NoError(t, err)

		actual, err := MapFromJSON(cty.StringVal(string(expected)))
		assert.NoError(t, err)
		assert.NotEmpty(t, actual)

		actualString, err := json.Marshal(&actual)
		require.NoError(t, err)

		assert.JSONEq(t, string(expected), string(actualString))
	})
}

func TestDifference(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"key": {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}, map[string]interface{}{
		"key": []interface{}{"a", "b"},
	})

	resourceData.SetId("1")

	oldValue, newValue := Difference(resourceData, "key")

	assert.Equal(t, []interface{}{"a", "b"}, oldValue)
	assert.Equal(t, []interface{}{}, newValue)
}
