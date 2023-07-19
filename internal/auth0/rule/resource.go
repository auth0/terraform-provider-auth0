package rule

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

var ruleNameRegexp = regexp.MustCompile(`^[^\s-][\w -]+[^\s-]$`)

// NewResource will return a new auth0_rule resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: "This resource is deprecated. Refer to the [guide on how to migrate from rules to actions](https://auth0.com/docs/customize/actions/migrate/migrate-from-rules-to-actions) " +
			"and manage your actions using the `auth0_action` resource.",
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
			"`auth0_rule_config` resource.\n\n!> This resource is deprecated. Refer to the [guide on how to migrate from rules to actions](https://auth0.com/docs/customize/actions/migrate/migrate-from-rules-to-actions) " +
			"and manage your actions using the `auth0_action` resource.",
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
				Computed:    true,
				Description: "Indicates whether the rule is enabled.",
			},
		},
	}
}

func createRule(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	rule := expandRule(d.GetRawConfig())

	if err := api.Rule.Create(ctx, rule); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(rule.GetID())

	return readRule(ctx, d, m)
}

func readRule(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
	rule, err := api.Rule.Read(ctx, d.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return diag.FromErr(flattenRule(d, rule))
}

func updateRule(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	rule := expandRule(d.GetRawConfig())

	if err := api.Rule.Update(ctx, d.Id(), rule); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return readRule(ctx, d, m)
}

func deleteRule(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.Rule.Delete(ctx, d.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return nil
}
