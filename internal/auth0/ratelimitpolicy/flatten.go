package ratelimitpolicy

import (
	"time"

	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// rateLimitPolicyResponse is satisfied by Create/Get/Update response content types,
// which share the same getter surface.
type rateLimitPolicyResponse interface {
	GetID() string
	GetResource() management.RateLimitPolicyResourceEnum
	GetConsumer() management.RateLimitPolicyConsumerEnum
	GetConsumerSelector() string
	GetConfiguration() *management.RateLimitPolicyConfiguration
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

func flattenRateLimitPolicy(data *schema.ResourceData, policy rateLimitPolicyResponse) diag.Diagnostics {
	result := multierror.Append(
		data.Set("resource", string(policy.GetResource())),
		data.Set("consumer", string(policy.GetConsumer())),
		data.Set("consumer_selector", policy.GetConsumerSelector()),
		data.Set("configuration", flattenRateLimitPolicyConfiguration(policy.GetConfiguration())),
		data.Set("created_at", policy.GetCreatedAt().Format(time.RFC3339)),
		data.Set("updated_at", policy.GetUpdatedAt().Format(time.RFC3339)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func flattenRateLimitPolicyConfiguration(cfg *management.RateLimitPolicyConfiguration) []interface{} {
	if cfg == nil {
		return nil
	}

	m := map[string]interface{}{}
	switch {
	case cfg.RateLimitPolicyConfigurationZero != nil:
		m["action"] = string(cfg.RateLimitPolicyConfigurationZero.GetAction())
	case cfg.RateLimitPolicyConfigurationOne != nil:
		m["action"] = string(cfg.RateLimitPolicyConfigurationOne.GetAction())
		m["limit"] = cfg.RateLimitPolicyConfigurationOne.GetLimit()
	case cfg.RateLimitPolicyConfigurationAction != nil:
		m["action"] = string(cfg.RateLimitPolicyConfigurationAction.GetAction())
		m["limit"] = cfg.RateLimitPolicyConfigurationAction.GetLimit()
		m["redirect_uri"] = cfg.RateLimitPolicyConfigurationAction.GetRedirectURI()
	default:
		return nil
	}
	return []interface{}{m}
}

// flattenRateLimitPolicyList converts a list of policies (from the List endpoint) into the
// computed `rate_limit_policies` attribute used by the plural data source.
func flattenRateLimitPolicyList(data *schema.ResourceData, policies []*management.RateLimitPolicy) error {
	if policies == nil {
		return data.Set("rate_limit_policies", make([]map[string]interface{}, 0))
	}

	list := make([]interface{}, 0, len(policies))
	for _, p := range policies {
		if p == nil {
			continue
		}
		list = append(list, map[string]interface{}{
			"id":                p.GetID(),
			"resource":          string(p.GetResource()),
			"consumer":          string(p.GetConsumer()),
			"consumer_selector": p.GetConsumerSelector(),
			"configuration":     flattenRateLimitPolicyConfiguration(p.GetConfiguration()),
			"created_at":        p.GetCreatedAt().Format(time.RFC3339),
			"updated_at":        p.GetUpdatedAt().Format(time.RFC3339),
		})
	}
	return data.Set("rate_limit_policies", list)
}
