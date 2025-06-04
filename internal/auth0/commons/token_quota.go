package commons

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/value"
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

// FlattenTokenQuota flattens the token quota configuration used for tenants, clients and organization.
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

// ExpandTokenQuota expands the token quota configuration for clients and organization.
func ExpandTokenQuota(raw cty.Value) *management.TokenQuota {
	if raw.IsNull() {
		return nil
	}

	var quota *management.TokenQuota

	raw.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		clientCredsValue := config.GetAttr("client_credentials")
		if clientCredsValue.IsNull() {
			return false
		}

		clientCredsValue.ForEachElement(func(_ cty.Value, credsConfig cty.Value) (stop bool) {
			enforce := value.Bool(credsConfig.GetAttr("enforce"))
			perHour := value.Int(credsConfig.GetAttr("per_hour"))
			perDay := value.Int(credsConfig.GetAttr("per_day"))

			quota = &management.TokenQuota{
				ClientCredentials: &management.TokenQuotaClientCredentials{
					Enforce: enforce,
				},
			}

			if perHour != nil {
				quota.ClientCredentials.PerHour = perHour
			}

			if perDay != nil {
				quota.ClientCredentials.PerDay = perDay
			}

			return false
		})

		return false
	})

	return quota
}

// IsTokenQuotaNull checks if the token quota configuration is null or empty used for clients and organization..
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
