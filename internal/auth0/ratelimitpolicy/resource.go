package ratelimitpolicy

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
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
		CustomizeDiff: validateRateLimitPolicyConfiguration,
		Description: "With this resource, you can manage Rate Limit Policies for the Authentication API, " +
			"applying per-application (client-ID-based) throttling so a single application's traffic spike " +
			"is isolated at the edge and does not exhaust the tenant's global rate limit. (EA only)",
		Schema: map[string]*schema.Schema{
			"resource": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(internalSchema.ToStrSlice(management.RateLimitPolicyResourceEnumOauthAuthenticationAPI), false),
				Description:  fmt.Sprintf("The resource the policy applies to. Valid values are: %v (EA Only)", internalSchema.ToStrSlice(management.RateLimitPolicyResourceEnumOauthAuthenticationAPI)),
			},
			"consumer": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(internalSchema.ToStrSlice(management.RateLimitPolicyConsumerEnumClient), false),
				Description:  fmt.Sprintf("The consumer category the policy applies to. Valid values are: %v (EA Only)", internalSchema.ToStrSlice(management.RateLimitPolicyConsumerEnumClient)),
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(rateLimitPolicyActions(), false),
							Description:  fmt.Sprintf("Action to take when the rate limit is exceeded. Valid values are: %v (EA only)", rateLimitPolicyActions()),
						},
						"limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 10000),
							Description: "Maximum number of requests allowed in a single window (0-10000). " +
								"Required and only valid for the `block`, `log`, and `redirect` actions. (EA only)",
						},
						"redirect_uri": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
							Description: "HTTPS URI to redirect to when the rate limit is exceeded. " +
								"Required and only valid for the `redirect` action. (EA only)",
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

// rateLimitPolicyActions returns the valid `action` values, which span multiple SDK union types.
func rateLimitPolicyActions() []string {
	actions := internalSchema.ToStrSlice(management.RateLimitPolicyConfigurationZeroActionAllow)
	actions = append(actions, internalSchema.ToStrSlice(management.RateLimitPolicyConfigurationOneActionBlock, management.RateLimitPolicyConfigurationOneActionLog)...)
	actions = append(actions, internalSchema.ToStrSlice(management.RateLimitPolicyConfigurationActionActionRedirect)...)

	return actions
}

func validateRateLimitPolicyConfiguration(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	action, limit, redirectURI := readConfiguration(diff.GetRawConfig().GetAttr("configuration"))

	return checkRateLimitPolicyConfiguration(action, limit, redirectURI)
}

// checkRateLimitPolicyConfiguration enforces only the fields each action *requires*. Inapplicable
// fields are ignored, not rejected, so generated config (which always emits `limit = 0`) round-trips.
func checkRateLimitPolicyConfiguration(action string, limit *int, redirectURI *string) error {
	switch action {
	case string(management.RateLimitPolicyConfigurationOneActionBlock),
		string(management.RateLimitPolicyConfigurationOneActionLog):
		if limit == nil {
			return fmt.Errorf("`limit` is required when `action` is %q", action)
		}
	case string(management.RateLimitPolicyConfigurationActionActionRedirect):
		if limit == nil {
			return fmt.Errorf("`limit` is required when `action` is %q", action)
		}
		if redirectURI == nil {
			return fmt.Errorf("`redirect_uri` is required when `action` is %q", action)
		}
	}

	return nil
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
		return internalError.HandleReadAPIError("auth0_rate_limit_policy", data, err)
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
