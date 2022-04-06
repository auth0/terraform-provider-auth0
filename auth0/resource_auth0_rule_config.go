package auth0

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func createRuleConfig(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ruleConfig := buildRuleConfig(d)
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
	ruleConfig := buildRuleConfig(d)
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

func buildRuleConfig(d *schema.ResourceData) *management.RuleConfig {
	return &management.RuleConfig{
		Key:   String(d, "key"),
		Value: String(d, "value"),
	}
}
