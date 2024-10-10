package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// GetRequestModels gets the config, plan, and state data as models.
func GetRequestModels[T any](ctx context.Context, requestConfig *tfsdk.Config, requestPlan *tfsdk.Plan, requestState *tfsdk.State) (configData T, planData T, stateData T, diagnostics diag.Diagnostics) {
	diagnostics.Append(requestConfig.Get(ctx, &configData)...)
	diagnostics.Append(requestPlan.Get(ctx, &planData)...)
	if requestState != nil && !requestState.Raw.IsNull() {
		diagnostics.Append(requestState.Get(ctx, &stateData)...)
	}

	return
}
