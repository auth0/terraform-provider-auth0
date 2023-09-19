package rule

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewConfigResource will return a new auth0_rule_config resource.
func NewConfigResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createRuleConfig,
		ReadContext:   readRuleConfig,
		UpdateContext: updateRuleConfig,
		DeleteContext: deleteRuleConfig,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With Auth0, you can create custom Javascript snippets that run in a secure, isolated sandbox " +
			"as part of your authentication pipeline, which are otherwise known as rules. This resource allows you " +
			"to create and manage variables that are available to all rules via Auth0's global configuration object. " +
			"Used in conjunction with configured rules.",
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Key for a rules configuration variable.",
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Value for a rules configuration variable.",
			},
		},
	}
}

func createRuleConfig(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	ruleConfig := expandRuleConfig(data.GetRawConfig())
	key := ruleConfig.GetKey()
	ruleConfig.Key = nil

	if err := api.RuleConfig.Upsert(ctx, key, ruleConfig); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(ruleConfig.GetKey())

	return readRuleConfig(ctx, data, meta)
}

func readRuleConfig(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	ruleConfig, err := api.RuleConfig.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(data.Set("key", ruleConfig.GetKey()))
}

func updateRuleConfig(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	ruleConfig := expandRuleConfig(data.GetRawConfig())
	ruleConfig.Key = nil

	if err := api.RuleConfig.Upsert(ctx, data.Id(), ruleConfig); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readRuleConfig(ctx, data, meta)
}

func deleteRuleConfig(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.RuleConfig.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
