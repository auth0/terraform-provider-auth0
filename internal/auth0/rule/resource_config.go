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

func createRuleConfig(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	ruleConfig := expandRuleConfig(d.GetRawConfig())
	key := ruleConfig.GetKey()
	ruleConfig.Key = nil

	if err := api.RuleConfig.Upsert(ctx, key, ruleConfig); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ruleConfig.GetKey())

	return readRuleConfig(ctx, d, m)
}

func readRuleConfig(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	ruleConfig, err := api.RuleConfig.Read(ctx, d.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return diag.FromErr(d.Set("key", ruleConfig.GetKey()))
}

func updateRuleConfig(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	ruleConfig := expandRuleConfig(d.GetRawConfig())
	ruleConfig.Key = nil

	if err := api.RuleConfig.Upsert(ctx, d.Id(), ruleConfig); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return readRuleConfig(ctx, d, m)
}

func deleteRuleConfig(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.RuleConfig.Delete(ctx, d.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return nil
}
