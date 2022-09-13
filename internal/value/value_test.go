package value

import (
	"math"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBool(t *testing.T) {
	actual := Bool(cty.NullVal(cty.Bool))
	assert.Nil(t, actual)

	actual = Bool(cty.BoolVal(true))
	require.NotNil(t, actual)
	assert.True(t, *actual)

	actual = Bool(cty.BoolVal(false))
	require.NotNil(t, actual)
	assert.False(t, *actual)
}

func TestString(t *testing.T) {
	actual := String(cty.NullVal(cty.String))
	assert.Nil(t, actual)

	actual = String(cty.StringVal(""))
	require.NotNil(t, actual)
	assert.Empty(t, *actual)

	expected := "foo bar"
	actual = String(cty.StringVal(expected))
	require.NotNil(t, actual)
	assert.Equal(t, expected, *actual)
}

func TestInt(t *testing.T) {
	actual := Int(cty.NullVal(cty.Number))
	assert.Nil(t, actual)

	var expected int64
	actual = Int(cty.NumberIntVal(expected))
	require.NotNil(t, actual)
	assert.Equal(t, int(expected), *actual)

	expected = -math.MaxInt64
	actual = Int(cty.NumberIntVal(expected))
	require.NotNil(t, actual)
	assert.Equal(t, int(expected), *actual)

	expected = math.MaxInt64
	actual = Int(cty.NumberIntVal(expected))
	require.NotNil(t, actual)
	assert.Equal(t, int(expected), *actual)
}

func TestFloat64(t *testing.T) {
	actual := Float64(cty.NullVal(cty.Number))
	assert.Nil(t, actual)

	var expected float64
	actual = Float64(cty.NumberFloatVal(expected))
	require.NotNil(t, actual)
	assert.Equal(t, expected, *actual)

	expected = -math.MaxFloat64
	actual = Float64(cty.NumberFloatVal(expected))
	require.NotNil(t, actual)
	assert.Equal(t, expected, *actual)

	expected = math.MaxFloat64
	actual = Float64(cty.NumberFloatVal(expected))
	require.NotNil(t, actual)
	assert.Equal(t, expected, *actual)
}

func TestStrings(t *testing.T) {
	actual := Strings(cty.NilVal)
	assert.Nil(t, actual)

	actual = Strings(cty.ListValEmpty(cty.String))
	require.NotNil(t, actual)
	assert.Empty(t, *actual)

	expected := []string{"localhost/logout", "https://app.domain.com/logout"}
	var testInput []cty.Value
	for _, value := range expected {
		testInput = append(testInput, cty.StringVal(value))
	}

	actual = Strings(cty.ListVal(testInput))
	require.NotNil(t, actual)
	assert.Equal(t, expected, *actual)
}

func TestMapOfStrings(t *testing.T) {
	actual := MapOfStrings(cty.NilVal)
	assert.Nil(t, actual)

	actual = MapOfStrings(cty.MapValEmpty(cty.String))
	require.NotNil(t, actual)
	assert.Empty(t, actual)

	expected := map[string]string{
		"logout": "http://app.domain.com/logout",
		"login":  "http://app.domain.com/login",
	}
	testInput := make(map[string]cty.Value)
	for key, value := range expected {
		testInput[key] = cty.StringVal(value)
	}

	actual = MapOfStrings(cty.MapVal(testInput))
	require.NotNil(t, actual)
	assert.Equal(t, expected, *actual)
}
