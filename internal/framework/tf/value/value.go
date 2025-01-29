package value

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// HasChange returns true if the attribute is modified in the plan from the value in the state.
// The stateValue and planValue arguments should be the Raw values from the state and the plan.
func HasChange(stateValue, planValue tftypes.Value, attribute string) bool {
	path := tftypes.NewAttributePath().WithAttributeName(attribute)
	rawPlanAttribute, _, err := tftypes.WalkAttributePath(planValue, path)
	if err != nil {
		return false
	}
	planAttribute := rawPlanAttribute.(tftypes.Value)

	rawStateAttribute, _, err := tftypes.WalkAttributePath(stateValue, path)
	if err != nil {
		return false
	}
	stateAttribute := rawStateAttribute.(tftypes.Value)

	return !stateAttribute.Equal(planAttribute)
}

// GetAttribute returns an attribute from a Raw state, plan, or config.
func GetAttribute(state tftypes.Value, attribute string) (tftypes.Value, error) {
	path := tftypes.NewAttributePath().WithAttributeName(attribute)
	rawValue, _, err := tftypes.WalkAttributePath(state, path)
	if err != nil {
		return tftypes.Value{}, err
	}
	return rawValue.(tftypes.Value), nil
}

// HasValue returns returns true if an attribute from a Raw state, plan, or config
// has a known value.
func HasValue(state tftypes.Value, attribute string) bool {
	val, err := GetAttribute(state, attribute)
	if err != nil {
		return false
	}

	return !val.IsNull() && val.IsKnown()
}

// Bool evaluates the typed value of the value
// and coerces to a pointer of a boolean.
func Bool(rawValue tftypes.Value) *bool {
	var ptr *bool
	if err := rawValue.As(&ptr); err != nil {
		return nil
	}
	return ptr
}

// String evaluates the typed value of the value
// and coerces to a pointer of a string.
func String(rawValue tftypes.Value) *string {
	var ptr *string
	if err := rawValue.As(&ptr); err != nil {
		return nil
	}
	return ptr
}

// Int evaluates the typed value of the value
// and coerces to a pointer of an int.
func Int(rawValue tftypes.Value) *int {
	var floatPtr *big.Float
	if err := rawValue.As(&floatPtr); err != nil || floatPtr == nil {
		return nil
	}
	i64, _ := floatPtr.Int64()
	value := int(i64)
	return &value
}

// Float64 evaluates the typed value of the value
// and coerces to a pointer of a float64.
func Float64(rawValue tftypes.Value) *float64 {
	var floatPtr *big.Float
	if err := rawValue.As(&floatPtr); err != nil || floatPtr == nil {
		return nil
	}
	value, _ := floatPtr.Float64()
	return &value
}

// Time evaluates the typed value of the value
// and coerces to a pointer of a string, which
// is then converted to a `time.Time` according
// to ISO 3339 (ISO 8601 is largely the same in
// common use cases, see https://ijmacd.github.io/rfc3339-iso8601/
// for differences).
func Time(rawValue tftypes.Value) *time.Time {
	var stringPtr *string
	if err := rawValue.As(&stringPtr); err != nil || stringPtr == nil {
		return nil
	}

	value, _ := time.Parse(time.RFC3339, *stringPtr)
	return &value
}

// Strings evaluates the typed value of the value
// and coerces to a pointer of a slice of strings.
func Strings(rawValue tftypes.Value) *[]string {
	var slicePtr *[]tftypes.Value
	if err := rawValue.As(&slicePtr); err != nil || slicePtr == nil {
		return nil
	}

	values := make([]string, 0, len(*slicePtr))
	for _, value := range *slicePtr {
		var stringPtr *string
		if err := value.As(&stringPtr); err != nil || stringPtr == nil {
			// This should never happen.
			continue
		}
		values = append(values, *stringPtr)
	}

	return &values
}

// MapOfStrings evaluates the typed value of the value
// and coerces to a pointer of a map of strings.
func MapOfStrings(rawValue tftypes.Value) *map[string]string {
	var mapPtr *map[string]tftypes.Value
	if err := rawValue.As(&mapPtr); err != nil || mapPtr == nil {
		return nil
	}

	m := make(map[string]string)
	for key, value := range *mapPtr {
		var stringPtr *string
		if err := value.As(&stringPtr); err != nil || stringPtr == nil {
			// This should never happen.
			continue
		}

		m[key] = *stringPtr
	}

	return &m
}

// MapFromJSON evaluates the typed value of the value
// and coerces to a map[string]interface{}.
func MapFromJSON(rawValue tftypes.Value) (map[string]interface{}, error) {
	var stringPtr *string
	if err := rawValue.As(&stringPtr); err != nil || stringPtr == nil {
		return nil, err
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(*stringPtr), &resultMap); err != nil {
		return nil, err
	}

	return resultMap, nil
}

// Difference accesses the value held by key and type asserts it to a set. It then
// compares its changes if any and returns what needs to be added and what
// needs to be removed. The stateValue and planValue arguments should be
// the Raw values from the state and the plan.
func Difference(stateValue, planValue tftypes.Value, attribute string) ([]tftypes.Value, []tftypes.Value, error) {
	// Zero the add and rm sets. These may be modified if the diff observed any changes.
	toAdd := make([]tftypes.Value, 0)
	toRemove := make([]tftypes.Value, 0)
	afterAttribute, err := GetAttribute(planValue, attribute)
	if err != nil {
		return nil, nil, err
	}
	beforeAttribute, err := GetAttribute(stateValue, attribute)
	if err != nil {
		return nil, nil, err
	}
	var beforeSet, afterSet *[]tftypes.Value
	if err := beforeAttribute.As(&beforeSet); err != nil {
		return nil, nil, err
	}
	if err := afterAttribute.As(&afterSet); err != nil {
		return nil, nil, err
	}
	if beforeSet != nil {
		for _, item := range *beforeSet {
			if !contains(afterSet, item) && !contains(&toRemove, item) {
				toRemove = append(toRemove, item)
			}
		}
	}
	if afterSet != nil {
		for _, item := range *afterSet {
			if !contains(beforeSet, item) && !contains(&toAdd, item) {
				toAdd = append(toAdd, item)
			}
		}
	}

	return toAdd, toRemove, nil
}

func contains(set *[]tftypes.Value, value tftypes.Value) bool {
	if set != nil {
		for _, item := range *set {
			if item.Equal(value) {
				return true
			}
		}
	}

	return false
}
