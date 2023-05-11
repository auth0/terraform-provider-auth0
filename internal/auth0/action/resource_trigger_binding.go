package action

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewTriggerBindingResource will return a new auth0_trigger_binding resource.
func NewTriggerBindingResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createTriggerBinding,
		ReadContext:   readTriggerBinding,
		UpdateContext: updateTriggerBinding,
		DeleteContext: deleteTriggerBinding,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can bind an action to a trigger. Once an action is created " +
			"and deployed, it can be attached (i.e. bound) to a trigger so that it will be executed as " +
			"part of a flow. The list of actions reflects the order in which they will be executed during " +
			"the appropriate flow.",
		Schema: map[string]*schema.Schema{
			"trigger": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"post-login",
					"credentials-exchange",
					"pre-user-registration",
					"post-user-registration",
					"post-change-password",
					"send-phone-message",
					"iga-approval",
					"iga-certification",
					"iga-fulfillment-assignment",
					"iga-fulfillment-execution",
				}, false),
				Description: "The ID of the trigger to bind with.",
			},
			"actions": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Action ID.",
						},
						"display_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of an action.",
						},
					},
				},
				Description: "The actions bound to this trigger",
			},
		},
	}
}

func createTriggerBinding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	id := d.Get("trigger").(string)
	triggerBindings := expandTriggerBindings(d.GetRawConfig().GetAttr("actions"))

	if err := api.Action.UpdateBindings(id, triggerBindings); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return readTriggerBinding(ctx, d, m)
}

func readTriggerBinding(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	triggerBindings, err := api.Action.Bindings(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("trigger", d.Id()),
		d.Set("actions", flattenTriggerBindingActions(triggerBindings.Bindings)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateTriggerBinding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	triggerBindings := expandTriggerBindings(d.GetRawConfig().GetAttr("actions"))
	api := m.(*config.Config).GetAPI()
	if err := api.Action.UpdateBindings(d.Id(), triggerBindings); err != nil {
		return diag.FromErr(err)
	}

	return readTriggerBinding(ctx, d, m)
}

func deleteTriggerBinding(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
	if err := api.Action.UpdateBindings(d.Id(), []*management.ActionBinding{}); err != nil {
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
