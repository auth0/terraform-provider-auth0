package value

import (
	"github.com/hashicorp/go-cty/cty"
)

// Bool evaluates the typed value of the value and coerces to a pointer of a boolean.
func Bool(rawValue cty.Value) *bool {
	if rawValue.IsNull() {
		return nil
	}

	r := rawValue.True()
	return &r
}

// String evaluates the typed value of the value and coerces to a pointer of a string.
func String(rawValue cty.Value) *string {
	if rawValue.IsNull() {
		return nil
	}

	r := rawValue.AsString()
	return &r
}

// Int evaluates the typed value of the value and coerces to a pointer of an int.
func Int(rawValue cty.Value) *int {
	if rawValue.IsNull() {
		return nil
	}

	intValue, _ := rawValue.AsBigFloat().Int64()

	i := int(intValue)

	return &i
}

// Strings evaluates the typed value of the value and coerces to a pointer of a slice of strings.
func Strings(rawValue cty.Value) *[]string {
	if rawValue.IsNull() {
		return nil
	}

	var s []string
	for _, val := range rawValue.AsValueSlice() {
		s = append(s, val.AsString())
	}

	return &s
}

// MapOfStrings evaluates the typed value of the value and coerces to a pointer of a map of strings.
func MapOfStrings(rawValue cty.Value) *map[string]string {
	if rawValue.IsNull() {
		return nil
	}

	m := make(map[string]string)
	for key, val := range rawValue.AsValueMap() {
		m[key] = val.AsString()
	}

	return &m
}
