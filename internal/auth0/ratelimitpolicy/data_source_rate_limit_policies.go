package ratelimitpolicy

import (
	"context"

	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewRateLimitPoliciesDataSource will return a new auth0_rate_limit_policies data source.
func NewRateLimitPoliciesDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readRateLimitPoliciesForDataSource,
		Description: "Data source to retrieve Rate Limit Policies, optionally filtered by " +
			"resource, consumer, or consumer selector. (EA only)",
		Schema: map[string]*schema.Schema{
			"resource": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter policies by resource. (EA only)",
			},
			"consumer": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"resource"},
				Description:  "Filter policies by consumer. Requires `resource` to also be set. (EA only)",
			},
			"consumer_selector": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"consumer"},
				Description:  "Filter policies by consumer selector. Requires `consumer` to also be set. (EA only)",
			},
			"rate_limit_policies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of Rate Limit Policies matching the filter criteria. (EA only)",
				Elem: &schema.Resource{
					Schema: rateLimitPolicyListElemSchema(),
				},
			},
		},
	}
}

func rateLimitPolicyListElemSchema() map[string]*schema.Schema {
	elem := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	elem["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ID of the Rate Limit Policy. (EA only)",
	}
	return elem
}

func readRateLimitPoliciesForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	params := &management.ListRateLimitPoliciesRequestParameters{}
	if v, ok := data.GetOk("resource"); ok {
		params.SetResource(management.RateLimitPolicyResourceEnum(v.(string)).Ptr())
	}
	if v, ok := data.GetOk("consumer"); ok {
		params.SetConsumer(management.RateLimitPolicyConsumerEnum(v.(string)).Ptr())
	}
	if v, ok := data.GetOk("consumer_selector"); ok {
		selector := v.(string)
		params.SetConsumerSelector(&selector)
	}

	page, err := apiv2.RateLimitPolicies.List(ctx, params)
	if err != nil {
		return diag.FromErr(err)
	}

	var policies []*management.RateLimitPolicy
	iterator := page.Iterator()
	for iterator.Next(ctx) {
		policies = append(policies, iterator.Current())
	}
	if err := iterator.Err(); err != nil {
		return diag.FromErr(err)
	}

	data.SetId("rate-limit-policies")

	return diag.FromErr(flattenRateLimitPolicyList(data, policies))
}
