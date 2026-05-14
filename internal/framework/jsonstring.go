// Package framework — JSON-string helper for very large/polymorphic API
// shapes (e.g. auth0_client.addons) that are impractical to enumerate as a
// typed schema. Users pass HCL-encoded JSON; we round-trip through `any` so
// the SDK can marshal it back. State stores the canonical (sorted-key) JSON
// so plan diffs are stable.
package framework

import (
	"bytes"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CanonicalJSON returns a stable string representation of v: an empty string
// when v is nil/empty, otherwise compact JSON with sorted map keys.
//
// The function deliberately round-trips through `any` (which decodes JSON
// objects as `map[string]any`) before re-encoding. Go's encoding/json sorts
// map keys but emits struct fields in declaration order, so a direct
// json.Marshal of the SDK's typed structs would produce a different byte
// sequence than the user's `jsonencode({...})` (which sorts keys). Without
// this round-trip the provider would surface "Provider produced inconsistent
// result after apply" diagnostics for every JSON-string passthrough field.
func CanonicalJSON(v any) (string, error) {
	if v == nil {
		return "", nil
	}
	first, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if len(first) == 0 || string(first) == "null" {
		return "", nil
	}
	// Decode into a generic value, then re-encode — this normalises any
	// SDK struct into a sorted-key JSON object.
	var generic any
	if err := json.Unmarshal(first, &generic); err != nil {
		return "", err
	}
	second, err := json.Marshal(generic)
	if err != nil {
		return "", err
	}
	var compacted bytes.Buffer
	if err := json.Compact(&compacted, second); err != nil {
		return "", err
	}
	return compacted.String(), nil
}

// ParseJSONString parses a TF JSON string into a generic any value. Returns
// (nil, ok=false) when the value is null/unknown/empty so callers know to
// skip the field. Adds an error diagnostic when the JSON is malformed.
func ParseJSONString(v types.String, attrName string, diags *diag.Diagnostics) (any, bool) {
	if v.IsNull() || v.IsUnknown() {
		return nil, false
	}
	s := v.ValueString()
	if s == "" {
		return nil, false
	}
	var out any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		diags.AddError(
			"Invalid JSON in attribute "+attrName,
			err.Error(),
		)
		return nil, false
	}
	return out, true
}

// FlattenJSONToString marshals v back into a TF string. nil → null.
func FlattenJSONToString(v any, diags *diag.Diagnostics) types.String {
	if v == nil {
		return types.StringNull()
	}
	s, err := CanonicalJSON(v)
	if err != nil {
		diags.AddError("Failed to encode JSON attribute", err.Error())
		return types.StringNull()
	}
	if s == "" {
		return types.StringNull()
	}
	return types.StringValue(s)
}
