package value

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
)

func TestBool(t *testing.T) {
	v := Bool(cty.NullVal(cty.Bool))
	if v != nil {
		t.Errorf("Expected to be null, got %t", *v)
	}

	v = Bool(cty.BoolVal(true))
	if *v != true {
		t.Errorf("expected to be true, got %t", *v)
	}

	v = Bool(cty.BoolVal(false))
	if *v != false {
		t.Errorf("expected to be false, got %t", *v)
	}
}

func TestString(t *testing.T) {
	v := String(cty.NullVal(cty.String))
	if v != nil {
		t.Errorf("Expected to be null, got %s", *v)
	}

	v = String(cty.StringVal(""))
	if *v != "" {
		t.Errorf("expected to be empty string, got %s", *v)
	}

	v = String(cty.StringVal("foo bar"))
	if *v != "foo bar" {
		t.Errorf("expected to be \"foo bar\", got %s", *v)
	}
}

func TestInt(t *testing.T) {
	v := Int(cty.NullVal(cty.Number))
	if v != nil {
		t.Errorf("Expected to be null, got %d", *v)
	}

	v = Int(cty.NumberIntVal(0))
	if *v != 0 {
		t.Errorf("expected to be 0, got %d", *v)
	}

	v = Int(cty.NumberIntVal(-math.MaxInt64))
	if *v != -math.MaxInt64 {
		t.Errorf("Expected to be %d, got %d", -math.MaxInt64, *v)
	}

	v = Int(cty.NumberIntVal(math.MaxInt64))
	if *v != math.MaxInt64 {
		t.Errorf("Expected to be %d, got %d", math.MaxInt64, *v)
	}
}

func TestFloat64(t *testing.T) {
	v := Float64(cty.NullVal(cty.Number))
	assert.Nil(t, v)

	v = Float64(cty.NumberFloatVal(0))
	assert.Equal(t, float64(0), *v)

	v = Float64(cty.NumberFloatVal(-math.MaxFloat64))
	if *v != -math.MaxFloat64 {
		t.Errorf("Expected to be %v, got %v", -math.MaxFloat64, *v)
	}

	v = Float64(cty.NumberFloatVal(math.MaxFloat64))
	if *v != math.MaxFloat64 {
		t.Errorf("Expected to be %v, got %v", math.MaxFloat64, *v)
	}
}

func TestStrings(t *testing.T) {
	mockSliceVals := []string{"localhost/logout", "https://app.domain.com/logout"}
	var mockSlice []cty.Value
	for _, v := range mockSliceVals {
		mockSlice = append(mockSlice, cty.StringVal(v))
	}
	r := Strings(cty.ListVal(mockSlice))
	if !reflect.DeepEqual(mockSliceVals, *r) {
		t.Errorf("expected to be %s, got %s", strings.Join(mockSliceVals, ", "), *r)
	}

	r = Strings(cty.ListValEmpty(cty.String))
	if len(*r) != 0 {
		t.Errorf("expected to be empty slice, got %s", *r)
	}

	r = Strings(cty.NilVal)
	if r != nil {
		t.Errorf("expected to be nil, got %s", *r)
	}
}

func TestMap(t *testing.T) {
	v := Map(cty.MapValEmpty(cty.String))
	if len(v) != 0 {
		t.Errorf("expected to be empty map, got %v", v)
	}

	v = Map(cty.NilVal)
	if v != nil {
		t.Errorf("expected to be nil, got %v", v)
	}

	mockMap := cty.ObjectVal(map[string]cty.Value{
		"logout": cty.StringVal("http://app.domain.com/logout"),
		"login":  cty.StringVal("http://app.domain.com/login"),
		"metadata": cty.ListVal([]cty.Value{
			cty.StringVal("one"),
			cty.StringVal("two"),
		}),
	})

	v = Map(mockMap)
	assert.Equal(
		t,
		map[string]interface{}{
			"logout":   "http://app.domain.com/logout",
			"login":    "http://app.domain.com/login",
			"metadata": []string{"one", "two"},
		},
		v,
	)
}

func TestMapOfStrings(t *testing.T) {
	r := MapOfStrings(cty.MapValEmpty(cty.String))
	if len(*r) != 0 {
		t.Errorf("expected to be empty map, got %s", *r)
	}

	r = MapOfStrings(cty.NilVal)
	if r != nil {
		t.Errorf("expected to be nil, got %s", *r)
	}

	mockMapVals := map[string]string{"logout": "http://app.domain.com/logout", "login": "http://app.domain.com/login"}
	mockMap := make(map[string]cty.Value)
	for key, v := range mockMapVals {
		mockMap[key] = cty.StringVal(v)
	}

	r = MapOfStrings(cty.MapVal(mockMap))
	if !reflect.DeepEqual(mockMapVals, *r) {
		t.Errorf("expected to be %s, got %s", mockMapVals, *r)
	}
}
