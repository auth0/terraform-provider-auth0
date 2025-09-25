package userattributeprofile

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

func TestNullDetection(t *testing.T) {
	// Test null value.
	nullValue := cty.NullVal(cty.String)
	if !nullValue.IsNull() {
		t.Errorf("Expected null value to be null")
	}

	// Test unknown value.
	unknownValue := cty.UnknownVal(cty.String)
	if unknownValue.IsNull() {
		t.Errorf("Expected unknown value to not be null")
	}

	// Test empty string.
	emptyValue := cty.StringVal("")
	if emptyValue.IsNull() {
		t.Errorf("Expected empty string to not be null")
	}

	// Test empty list.
	emptyList := cty.ListValEmpty(cty.String)
	if emptyList.IsNull() {
		t.Errorf("Expected empty list to not be null")
	}

	t.Logf("Null value: IsNull=%t", nullValue.IsNull())
	t.Logf("Unknown value: IsNull=%t", unknownValue.IsNull())
	t.Logf("Empty string: IsNull=%t", emptyValue.IsNull())
	t.Logf("Empty list: IsNull=%t", emptyList.IsNull())
}
