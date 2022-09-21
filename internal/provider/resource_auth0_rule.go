package provider

import (
	"context"
	"net/http"
	"regexp"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

var ruleNameRegexp = regexp.MustCompile(`^[^\s-][\w -]+[^\s-]$`)

func newRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: createRule,
		ReadContext:   readRule,
		UpdateContext: updateRule,
		DeleteContext: deleteRule,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With Auth0, you can create custom Javascript snippets that run in a secure, isolated sandbox " +
			"as part of your authentication pipeline, which are otherwise known as rules. This resource allows you " +
			"to create and manage rules. You can create global variable for use with rules by using the " +
			"`auth0_rule_config` resource.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					ruleNameRegexp,
					"Can only contain alphanumeric characters, spaces and '-'. "+
						"Can neither start nor end with '-' or spaces.",
				),
				Description: "Name of the rule. May only contain alphanumeric characters, spaces, and hyphens. " +
					"May neither start nor end with hyphens or spaces.",
			},
			"script": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Code to be executed when the rule runs.",
			},
			"order": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				Description: "Order in which the rule executes relative to other rules. " +
					"Lower-valued rules execute first.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether the rule is enabled.",
			},
		},
	}
}

func createRule(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rule := expandRule(d.GetRawConfig())
	api := m.(*management.Management)
	if err := api.Rule.Create(rule); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(rule.ID))

	return readRule(ctx, d, m)
}

func readRule(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	rule, err := api.Rule.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("name", rule.Name),
		d.Set("script", rule.Script),
		d.Set("order", rule.Order),
		d.Set("enabled", rule.Enabled),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateRule(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rule := expandRule(d.GetRawConfig())
	api := m.(*management.Management)
	if err := api.Rule.Update(d.Id(), rule); err != nil {
		return diag.FromErr(err)
	}

	return readRule(ctx, d, m)
}

func deleteRule(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.Rule.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	return nil
}

func expandRule(d cty.Value) *management.Rule {
	return &management.Rule{
		Name:    value.String(d.GetAttr("name")),
		Script:  value.String(d.GetAttr("script")),
		Order:   value.Int(d.GetAttr("order")),
		Enabled: value.Bool(d.GetAttr("enabled")),
	}
}
