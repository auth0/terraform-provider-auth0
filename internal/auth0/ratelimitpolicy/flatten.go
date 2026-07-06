package ratelimitpolicy

import (
	"encoding/json"

	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func flattenRateLimitPolicy(data *schema.ResourceData, policy *management.GetRateLimitPolicyResponseContent) diag.Diagnostics {
	result := multierror.Append(
		data.Set("resource", string(policy.GetResource())),
		data.Set("consumer", string(policy.GetConsumer())),
		data.Set("consumer_selector", policy.GetConsumerSelector()),
		data.Set("configuration", flattenRateLimitPolicyConfiguration(policy.GetConfiguration())),
		data.Set("created_at", value.FormatTime(policy.GetCreatedAt())),
		data.Set("updated_at", value.FormatTime(policy.GetUpdatedAt())),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func flattenRateLimitPolicyConfiguration(cfg *management.RateLimitPolicyConfiguration) []interface{} {
	// The SDK's oneOf deserialization always matches the allow (Zero) variant,
	// so block/log/redirect land in GetRateLimitPolicyConfigurationZero with their
	// limit/redirect_uri in the variant's extra properties.
	zero := cfg.GetRateLimitPolicyConfigurationZero()
	if zero == nil {
		return nil
	}

	m := map[string]interface{}{
		"action": string(zero.GetAction()),
	}

	extra := zero.GetExtraProperties()
	if limit, ok := extraInt(extra, "limit"); ok {
		m["limit"] = limit
	}
	if uri, ok := extra["redirect_uri"].(string); ok {
		m["redirect_uri"] = uri
	}

	return []interface{}{m}
}

// extraInt reads an int from extra properties, where encoding/json yields float64 or json.Number.
func extraInt(extra map[string]interface{}, key string) (int, bool) {
	switch n := extra[key].(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i), true
		}
	}

	return 0, false
}

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
			"created_at":        value.FormatTime(p.GetCreatedAt()),
			"updated_at":        value.FormatTime(p.GetUpdatedAt()),
		})
	}
	return data.Set("rate_limit_policies", list)
}
