package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func newTriggerBinding() *schema.Resource {
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
	id := d.Get("trigger").(string)
	triggerBindings := expandTriggerBindings(d)
	api := m.(*management.Management)
	if err := api.Action.UpdateBindings(id, triggerBindings); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return readTriggerBinding(ctx, d, m)
}

func readTriggerBinding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	triggerBindings, err := api.Action.Bindings(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	return diag.FromErr(d.Set("actions", flattenTriggerBindingActions(triggerBindings.Bindings)))
}

func updateTriggerBinding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	triggerBindings := expandTriggerBindings(d)
	api := m.(*management.Management)
	if err := api.Action.UpdateBindings(d.Id(), triggerBindings); err != nil {
		return diag.FromErr(err)
	}

	return readTriggerBinding(ctx, d, m)
}

func deleteTriggerBinding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
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

func expandTriggerBindings(d *schema.ResourceData) []*management.ActionBinding {
	var triggerBindings []*management.ActionBinding

	List(d, "actions").Elem(func(d ResourceData) {
		triggerBindings = append(triggerBindings, &management.ActionBinding{
			Ref: &management.ActionBindingReference{
				Type:  auth0.String("action_id"),
				Value: String(d, "id"),
			},
			DisplayName: String(d, "display_name"),
		})
	})

	return triggerBindings
}

func flattenTriggerBindingActions(bindings []*management.ActionBinding) []interface{} {
	var triggerBindingActions []interface{}

	for _, binding := range bindings {
		triggerBindingActions = append(
			triggerBindingActions,
			map[string]interface{}{
				"id":           binding.Action.GetID(),
				"display_name": binding.GetDisplayName(),
			},
		)
	}

	return triggerBindingActions
}
