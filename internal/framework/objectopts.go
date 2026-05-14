package framework

import "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

// ObjectAsOpts returns the standard, lenient ObjectAsOptions used everywhere
// in this provider when decoding nested objects: tolerate null/unknown leaves
// rather than erroring out. Equivalent SDKv2 helper: schema.TypeMap with
// Optional + Computed semantics.
func ObjectAsOpts() basetypes.ObjectAsOptions {
	return basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true}
}
