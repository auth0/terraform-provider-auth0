package commons

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	maxTokenQuotaLimit = 2147483647
)

// TokenQuotaSchema returns the common schema used for tenants, clients and organization.
func TokenQuotaSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The token quota configuration.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"client_credentials": {
					Type:        schema.TypeList,
					Required:    true,
					MaxItems:    1,
					Description: "The token quota configuration for client credentials.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"enforce": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "If enabled, the quota will be enforced and requests in excess of the quota will fail. If disabled, the quota will not be enforced, but notifications for requests exceeding the quota will be available in logs.",
								Default:     true,
							},
							"per_day": {
								Type:         schema.TypeInt,
								Optional:     true,
								Description:  "Maximum number of issued tokens per day",
								ValidateFunc: validation.IntBetween(1, maxTokenQuotaLimit),
							},
							"per_hour": {
								Type:         schema.TypeInt,
								Optional:     true,
								Description:  "Maximum number of issued tokens per hour",
								ValidateFunc: validation.IntBetween(1, maxTokenQuotaLimit),
							},
						},
					},
				},
			},
		},
	}
}

// DefaultTokenQuotaSchema returns the common schema used for tenants, clients and organisation.
func DefaultTokenQuotaSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Token Quota configuration.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"clients":       TokenQuotaSchema(),
				"organizations": TokenQuotaSchema(),
			},
		},
	}
}

// FlattenTokenQuota flattens the token quota configuration for a client.
func FlattenTokenQuota(tokenQuota *management.TokenQuota) []interface{} {
	if tokenQuota == nil || tokenQuota.ClientCredentials == nil {
		return nil
	}

	result := make(map[string]interface{})

	if tokenQuota.ClientCredentials != nil {
		clientCreds := map[string]interface{}{
			"enforce": tokenQuota.ClientCredentials.GetEnforce(),
		}

		if tokenQuota.ClientCredentials.PerHour != nil {
			clientCreds["per_hour"] = tokenQuota.ClientCredentials.GetPerHour()
		}

		if tokenQuota.ClientCredentials.PerDay != nil {
			clientCreds["per_day"] = tokenQuota.ClientCredentials.GetPerDay()
		}

		result["client_credentials"] = []interface{}{clientCreds}
	}

	return []interface{}{result}
}

// IsTokenQuotaNull checks if the token quota configuration is null or empty.
func IsTokenQuotaNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("token_quota") {
		return false
	}

	rawConfig := data.GetRawConfig()
	if rawConfig.IsNull() {
		return true
	}

	tokenQuotaConfig := rawConfig.GetAttr("token_quota")
	if tokenQuotaConfig.IsNull() {
		return true
	}

	empty := true
	tokenQuotaConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		clientCreds := cfg.GetAttr("client_credentials")
		if !clientCreds.IsNull() {
			clientCreds.ForEachElement(func(_ cty.Value, creds cty.Value) (stop bool) {
				enforce := creds.GetAttr("enforce")
				perHour := creds.GetAttr("per_hour")
				perDay := creds.GetAttr("per_day")

				if (!enforce.IsNull() && enforce.True()) ||
					(!perHour.IsNull()) ||
					(!perDay.IsNull()) {
					empty = false
					stop = true
				}
				return stop
			})
		}
		return false
	})

	return empty
}
