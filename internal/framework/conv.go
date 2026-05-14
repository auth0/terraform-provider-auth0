// Package framework — typed helpers for converting between go-auth0 SDK
// values and terraform-plugin-framework `types.*` values.
//
// The SDK overwhelmingly uses `*string`, `*bool`, `*int`; we also see typed
// enum aliases (`type FooEnum string`) and `[]string`. The helpers below
// remove the boilerplate of nil-checks at every call site so resource files
// can stay short.
package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// -- pointer scalars -------------------------------------------------------

// EnumPtrToString converts any `type FooEnum string` pointer to a TF string
// value (null when the pointer is nil).
func EnumPtrToString[E ~string](p *E) types.String {
	if p == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*p))
}

// IntPtrToInt64 converts a *int to a TF int64 value.
func IntPtrToInt64(p *int) types.Int64 {
	if p == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*p))
}

// Int64ToIntPtr is the Terraform → Go counterpart of IntPtrToInt64.
func Int64ToIntPtr(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}

// -- string slices / enum slices ------------------------------------------

// StringSliceToList converts []string into a TF list (null for nil input).
func StringSliceToList(in []string) types.List {
	if in == nil {
		return types.ListNull(types.StringType)
	}
	vals := make([]attr.Value, len(in))
	for i, s := range in {
		vals[i] = types.StringValue(s)
	}
	lv, _ := types.ListValue(types.StringType, vals)
	return lv
}

// StringListToGo is the inverse of StringSliceToList. Returns nil for
// null/unknown so optional fields are omitted from API requests.
func StringListToGo(ctx context.Context, l types.List, diags *diag.Diagnostics) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]string, 0, len(l.Elements()))
	diags.Append(l.ElementsAs(ctx, &out, false)...)
	return out
}

// EnumSliceToList converts a slice of `~string` enums into a TF list.
func EnumSliceToList[E ~string](in []E) types.List {
	if in == nil {
		return types.ListNull(types.StringType)
	}
	vals := make([]attr.Value, len(in))
	for i, e := range in {
		vals[i] = types.StringValue(string(e))
	}
	lv, _ := types.ListValue(types.StringType, vals)
	return lv
}
