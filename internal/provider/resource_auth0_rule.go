package provider

import (
	"context"
	"net/http"
	"regexp"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					ruleNameRegexp,
					"Can only contain alphanumeric characters, spaces and '-'. "+
						"Can neither start nor end with '-' or spaces."),
			},
			"script": {
				Type:     schema.TypeString,
				Required: true,
			},
			"order": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func createRule(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rule := buildRule(d)
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
	rule := buildRule(d)
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

func buildRule(d *schema.ResourceData) *management.Rule {
	return &management.Rule{
		Name:    String(d, "name"),
		Script:  String(d, "script"),
		Order:   Int(d, "order"),
		Enabled: Bool(d, "enabled"),
	}
}
