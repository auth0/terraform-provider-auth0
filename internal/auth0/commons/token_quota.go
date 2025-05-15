package commons

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	maxTokenQuotaLimit = 2147483647
)

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

func DefaultTokenQuotaSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "Token Quota configuration...",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"clients":       TokenQuotaSchema(),
				"organizations": TokenQuotaSchema(),
			},
		},
	}
}
