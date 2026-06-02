package ratelimitpolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_rate_limit_policy data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readRateLimitPolicyForDataSource,
		Description: "Data source to retrieve a single Rate Limit Policy by ID. (EA only)",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["policy_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Rate Limit Policy. (EA only)",
	}
	return dataSourceSchema
}

func readRateLimitPolicyForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	policyID := data.Get("policy_id").(string)

	policy, err := apiv2.RateLimitPolicies.Get(ctx, policyID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(policy.GetID())

	return flattenRateLimitPolicy(data, policy)
}
