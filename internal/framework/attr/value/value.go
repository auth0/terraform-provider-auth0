package value

import (
	"encoding/json"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// HasChange returns true if the attribute is modified from the previous value.
func HasChange(before, after attr.Value) bool {
	if HasValue(before) {
		return !HasValue(after) || !before.Equal(after)
	}
	return HasValue(after)
}

// HasValue returns returns true if an attribute has a known value.
func HasValue(val attr.Value) bool {
	return !val.IsNull() && !val.IsUnknown()
}

// Bool evaluates the typed value of the value
// and converts to a pointer of a boolean.
func Bool(attrValue attr.Value) *bool {
	if attrValue.IsUnknown() {
		return nil
	}

	return attrValue.(types.Bool).ValueBoolPointer()
}

// String evaluates the typed value of the value
// and converts to a pointer of a string.
func String(attrValue attr.Value) *string {
	if attrValue.IsUnknown() || attrValue.IsNull() {
		return nil
	}

	var rval string
	switch t := attrValue.(type) {
	case types.String:
		rval = t.ValueString()
	default:
		rval = attrValue.String()
	}
	return &rval
}

// Int evaluates the typed value of the value
// and converts to a pointer of an int.
func Int(attrValue attr.Value) *int {
	if attrValue.IsUnknown() || attrValue.IsNull() {
		return nil
	}

	var rval int
	switch t := attrValue.(type) {
	case types.Int32:
		rval = int(t.ValueInt32())
	case types.Int64:
		rval = int(t.ValueInt64())
	case types.Float32:
		rval = int(t.ValueFloat32())
	case types.Float64:
		rval = int(t.ValueFloat64())
	default:
		bigInt, _ := attrValue.(types.Number).ValueBigFloat().Int(nil)
		rval = int(bigInt.Int64())
	}
	return &rval
}

// Float64 evaluates the typed value of the value
// and converts to a pointer of a float64.
func Float64(attrValue attr.Value) *float64 {
	if attrValue.IsUnknown() || attrValue.IsNull() {
		return nil
	}

	var rval float64
	switch t := attrValue.(type) {
	case types.Int32:
		rval = float64(t.ValueInt32())
	case types.Int64:
		rval = float64(t.ValueInt64())
	case types.Float32:
		rval = float64(t.ValueFloat32())
	case types.Float64:
		rval = t.ValueFloat64()
	default:
		rval, _ = attrValue.(types.Number).ValueBigFloat().Float64()
	}
	return &rval
}

// Time evaluates the typed value of the value
// and converts to a pointer of a string, which
// is then converted to a `time.Time` according
// to ISO 3339 (ISO 8601 is largely the same in
// common use cases, see https://ijmacd.github.io/rfc3339-iso8601/
// for differences).
func Time(attrValue attr.Value) *time.Time {
	if attrValue.IsUnknown() || attrValue.IsNull() {
		return nil
	}

	var rval time.Time
	switch t := attrValue.(type) {
	case timetypes.RFC3339:
		rval, _ = t.ValueRFC3339Time()
	default:
		rval, _ = time.Parse(time.RFC3339, attrValue.String())
	}
	return &rval
}

// Strings evaluates the typed value of the value
// and converts to a pointer of a slice of strings.
func Strings(attrValue attr.Value) *[]string {
	if attrValue.IsNull() || attrValue.IsUnknown() {
		return nil
	}

	var elements []attr.Value
	switch t := attrValue.(type) {
	case types.Set:
		elements = t.Elements()
	default:
		elements = attrValue.(types.List).Elements()
	}

	rval := make([]string, 0, len(elements))
	for _, element := range elements {
		rval = append(rval, element.String())
	}
	return &rval
}

// MapOfStrings evaluates the typed value of the value
// and converts to a pointer of a map of strings.
func MapOfStrings(attrValue attr.Value) *map[string]string {
	if attrValue.IsNull() || attrValue.IsUnknown() {
		return nil
	}

	var elements map[string]attr.Value
	switch t := attrValue.(type) {
	case types.Object:
		elements = t.Attributes()
	default:
		elements = attrValue.(types.Map).Elements()
	}

	rval := make(map[string]string)
	for key, element := range elements {
		rval[key] = element.String()
	}
	return &rval
}

// MapFromJSON evaluates the typed value of the value
// and converts to a map[string]interface{}.
func MapFromJSON(attrValue attr.Value) (map[string]interface{}, error) {
	if attrValue.IsUnknown() || attrValue.IsNull() {
		return nil, nil
	}

	var rval map[string]interface{}
	if err := json.Unmarshal([]byte(attrValue.String()), &rval); err != nil {
		return nil, err
	}

	return rval, nil
}

// Difference compares two sets for changes, if any and returns what needs to be added
// and what needs to be removed.
func Difference(before, after attr.Value) ([]attr.Value, []attr.Value, error) {
	// Zero the add and rm sets. These may be modified if the diff observed any changes.
	toAdd := make([]attr.Value, 0)
	toRemove := make([]attr.Value, 0)

	var beforeSet, afterSet []attr.Value
	if !before.IsNull() && !before.IsUnknown() {
		switch t := before.(type) {
		case types.Set:
			beforeSet = t.Elements()
		default:
			beforeSet = before.(types.List).Elements()
		}
	}
	if !after.IsNull() && !after.IsUnknown() {
		switch t := after.(type) {
		case types.Set:
			afterSet = t.Elements()
		default:
			afterSet = after.(types.List).Elements()
		}
	}

	for _, item := range beforeSet {
		if !contains(afterSet, item) && !contains(toRemove, item) {
			toRemove = append(toRemove, item)
		}
	}
	for _, item := range afterSet {
		if !contains(beforeSet, item) && !contains(toAdd, item) {
			toAdd = append(toAdd, item)
		}
	}

	return toAdd, toRemove, nil
}

func contains(set []attr.Value, value attr.Value) bool {
	for _, item := range set {
		if item.Equal(value) {
			return true
		}
	}

	return false
}

// AttrBool converts a *bool to a types.Bool value.
func AttrBool(value *bool) types.Bool {
	if value == nil {
		return basetypes.NewBoolNull()
	}

	return basetypes.NewBoolValue(*value)
}

// AttrString converts a *string to a types.String value.
func AttrString(value *string) types.String {
	if value == nil {
		return basetypes.NewStringNull()
	}

	return basetypes.NewStringValue(*value)
}

// AttrInt64 converts an *int64, *int32, or *int to a types.Int64 value.
func AttrInt64[T int | int64 | int32](value *T) types.Int64 {
	if value == nil {
		return basetypes.NewInt64Null()
	}

	return basetypes.NewInt64Value(int64(*value))
}

// AttrInt32 converts an *int32 to a types.Int64 value.
func AttrInt32(value *int32) types.Int32 {
	if value == nil {
		return basetypes.NewInt32Null()
	}

	return basetypes.NewInt32Value(*value)
}

// AttrFloat64 converts a *float64 to a types.Float64 value.
func AttrFloat64(value *float64) types.Float64 {
	if value == nil {
		return basetypes.NewFloat64Null()
	}

	return basetypes.NewFloat64Value(*value)
}

// AttrFloat32 converts a *float32 to a types.Float32 value.
func AttrFloat32(value *float32) types.Float32 {
	if value == nil {
		return basetypes.NewFloat32Null()
	}

	return basetypes.NewFloat32Value(*value)
}

// AttrTime converts a *time.Time to a timetypes.RFC3339 value.
func AttrTime(value *time.Time) timetypes.RFC3339 {
	return timetypes.NewRFC3339TimePointerValue(value)
}
