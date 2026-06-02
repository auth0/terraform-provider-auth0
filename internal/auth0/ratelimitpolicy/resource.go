package ratelimitpolicy

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_rate_limit_policy resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createRateLimitPolicy,
		ReadContext:   readRateLimitPolicy,
		UpdateContext: updateRateLimitPolicy,
		DeleteContext: deleteRateLimitPolicy,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage Rate Limit Policies for the Authentication API, " +
			"applying per-application (client-ID-based) throttling so a single application's traffic spike " +
			"is isolated at the edge and does not exhaust the tenant's global rate limit. (EA only)",
		Schema: map[string]*schema.Schema{
			"resource": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The resource the policy applies to. Currently only `oauth_authentication_api` is supported. (EA only)",
			},
			"consumer": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The consumer category the policy applies to. Currently only `client` is supported. (EA only)",
			},
			"consumer_selector": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "Identifier or category within the consumer to which the policy applies. Supported values: " +
					"`client_id:<client_id>`, `client_id:<cimd_uri>`, `cimd_clients`, `third_party_clients`, or `default`. (EA only)",
			},
			"configuration": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "The throttling configuration applied when the rate limit is reached. (EA only)",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Action to take when the rate limit is exceeded. One of `allow`, `block`, " +
								"`log`, or `redirect`. (EA only)",
						},
						"limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Description: "Maximum number of requests allowed in a single window (0-10000). " +
								"Required for `block`, `log`, and `redirect` actions. (EA only)",
						},
						"redirect_uri": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "HTTPS URI to redirect to when the rate limit is exceeded. " +
								"Required (and only valid) for the `redirect` action. (EA only)",
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time when the rate limit policy was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time when the rate limit policy was last updated.",
			},
		},
	}
}

func createRateLimitPolicy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	policy, err := apiv2.RateLimitPolicies.Create(ctx, expandRateLimitPolicyCreate(data))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(policy.GetID())

	return readRateLimitPolicy(ctx, data, meta)
}

func readRateLimitPolicy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	policy, err := apiv2.RateLimitPolicies.Get(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return flattenRateLimitPolicy(data, policy)
}

func updateRateLimitPolicy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	if _, err := apiv2.RateLimitPolicies.Update(ctx, data.Id(), expandRateLimitPolicyPatch(data)); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readRateLimitPolicy(ctx, data, meta)
}

func deleteRateLimitPolicy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	if err := apiv2.RateLimitPolicies.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
