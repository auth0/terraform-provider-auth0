package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func newRuleConfig() *schema.Resource {
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
	ruleConfig := expandRuleConfig(d.GetRawConfig())
	key := auth0.StringValue(ruleConfig.Key)
	ruleConfig.Key = nil
	api := m.(*management.Management)
	if err := api.RuleConfig.Upsert(key, ruleConfig); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(ruleConfig.Key))

	return readRuleConfig(ctx, d, m)
}

func readRuleConfig(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	ruleConfig, err := api.RuleConfig.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	return diag.FromErr(d.Set("key", ruleConfig.Key))
}

func updateRuleConfig(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ruleConfig := expandRuleConfig(d.GetRawConfig())
	ruleConfig.Key = nil
	api := m.(*management.Management)
	if err := api.RuleConfig.Upsert(d.Id(), ruleConfig); err != nil {
		return diag.FromErr(err)
	}

	return readRuleConfig(ctx, d, m)
}

func deleteRuleConfig(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.RuleConfig.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func expandRuleConfig(d cty.Value) *management.RuleConfig {
	return &management.RuleConfig{
		Key:   value.String(d.GetAttr("key")),
		Value: value.String(d.GetAttr("value")),
	}
}
