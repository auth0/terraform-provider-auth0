package value

import (
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/go-cty/cty"
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
