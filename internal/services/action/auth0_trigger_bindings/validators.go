package auth0triggerbindings

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// triggerValidator ensures the `trigger` value is one of the trigger types
// supported by the Auth0 Management API.
type triggerValidator struct{}

func (triggerValidator) Description(_ context.Context) string {
	return "value must be a valid Auth0 action trigger"
}

func (v triggerValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (triggerValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := mgmt.NewActionTriggerTypeEnumFromString(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid trigger",
			err.Error(),
		)
	}
}
