package value

import (
	"github.com/hashicorp/go-cty/cty"
)

// Bool evaluates the typed value of the value
// and coerces to a pointer of a boolean.
func Bool(rawValue cty.Value) *bool {
	if rawValue.IsNull() {
		return nil
	}

	value := rawValue.True()
	return &value
}

// String evaluates the typed value of the value
// and coerces to a pointer of a string.
func String(rawValue cty.Value) *string {
	if rawValue.IsNull() {
		return nil
	}

	value := rawValue.AsString()
	return &value
}

// Int evaluates the typed value of the value
// and coerces to a pointer of an int.
func Int(rawValue cty.Value) *int {
	if rawValue.IsNull() {
		return nil
	}

	int64Value, _ := rawValue.AsBigFloat().Int64()
	value := int(int64Value)
	return &value
}

// Float64 evaluates the typed value of the value
// and coerces to a pointer of a float64.
func Float64(rawValue cty.Value) *float64 {
	if rawValue.IsNull() {
		return nil
	}

	value, _ := rawValue.AsBigFloat().Float64()
	return &value
}

// Strings evaluates the typed value of the value
// and coerces to a pointer of a slice of strings.
func Strings(rawValues cty.Value) *[]string {
	if rawValues.IsNull() {
		return nil
	}

	var value []string
	for _, rawValue := range rawValues.AsValueSlice() {
		value = append(value, rawValue.AsString())
	}

	return &value
}

// MapOfStrings evaluates the typed value of the value
// and coerces to a pointer of a map of strings.
func MapOfStrings(rawValue cty.Value) *map[string]string {
	if rawValue.IsNull() {
		return nil
	}

	m := make(map[string]string)
	for key, value := range rawValue.AsValueMap() {
		m[key] = value.AsString()
	}

	return &m
}
