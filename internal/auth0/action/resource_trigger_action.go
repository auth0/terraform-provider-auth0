package action

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewTriggerActionResource will return a new auth0_trigger_action resource.
func NewTriggerActionResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createTriggerAction,
		ReadContext:   readTriggerAction,
		UpdateContext: updateTriggerAction,
		DeleteContext: deleteTriggerAction,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("trigger", "action_id"),
		},
		Description: "With this resource, you can bind an action to a trigger. Once an action is created and deployed, it can be attached (i.e. bound) to a trigger so that it will be executed as part of a flow.\n\nOrdering of an action within a specific flow is not currently supported when using this resource; the action will get appended to the end of the flow. To precisely manage ordering, it is advised to either do so with the dashboard UI or with the `auth0_trigger_bindings` resource.",
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
				Description: "The ID of the trigger to bind with. Available options: `post-login`, `credentials-exchange`, `pre-user-registration`, `post-user-registration`, `post-change-password`, `send-phone-message`, `iga-approval`, `iga-certification`, `iga-fulfillment-assignment`, `iga-fulfillment-execution`,",
			},
			"action_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the action to bind to the trigger.",
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name for this action within the trigger. This can be useful for distinguishing between multiple instances of the same action bound to a trigger. Defaults to action name when not provided.",
			},
		},
	}
}

func createTriggerAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	trigger := d.Get("trigger").(string)
	actionID := d.Get("action_id").(string)
	displayName := d.Get("display_name").(string)

	currentBindings, err := api.Action.Bindings(ctx, trigger)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var updatedBindings []*management.ActionBinding
	for _, binding := range currentBindings.Bindings {
		if binding.Action.GetID() == actionID {
			internalSchema.SetResourceGroupID(d, trigger, actionID)
			return nil
		}

		updatedBindings = append(updatedBindings, &management.ActionBinding{
			Ref: &management.ActionBindingReference{
				Type:  auth0.String("action_id"),
				Value: binding.Action.ID,
			},
			DisplayName: binding.DisplayName,
		})
	}

	if displayName == "" {
		action, err := api.Action.Read(ctx, actionID)

		if err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
		displayName = action.GetName()
	}

	updatedBindings = append(updatedBindings, &management.ActionBinding{
		Ref: &management.ActionBindingReference{
			Type:  auth0.String("action_id"),
			Value: &actionID,
		},
		DisplayName: &displayName,
	})

	if err := api.Action.UpdateBindings(ctx, trigger, updatedBindings); err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(d, trigger, actionID)
	return readTriggerAction(ctx, d, m)
}

func updateTriggerAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	trigger := d.Get("trigger").(string)
	actionID := d.Get("action_id").(string)
	displayName := d.Get("display_name").(string)

	currentBindings, err := api.Action.Bindings(ctx, trigger)
	if err != nil {
		return diag.FromErr(err)
	}

	found := false
	var updatedBindings []*management.ActionBinding
	for _, binding := range currentBindings.Bindings {
		if binding.Action.GetID() == actionID {
			updatedBindings = append(updatedBindings, &management.ActionBinding{
				Ref: &management.ActionBindingReference{
					Type:  auth0.String("action_id"),
					Value: &actionID,
				},
				DisplayName: &displayName,
			})
			found = true
			continue
		}
		updatedBindings = append(updatedBindings, &management.ActionBinding{
			Ref: &management.ActionBindingReference{
				Type:  auth0.String("action_id"),
				Value: binding.Action.ID,
			},
			DisplayName: binding.DisplayName,
		})
	}

	if !found {
		d.SetId("")
		return nil
	}

	if err := api.Action.UpdateBindings(ctx, trigger, updatedBindings); err != nil {
		return diag.FromErr(err)
	}

	return readTriggerAction(ctx, d, m)
}

func readTriggerAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	trigger := d.Get("trigger").(string)
	actionID := d.Get("action_id").(string)

	triggerBindings, err := api.Action.Bindings(ctx, trigger)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, binding := range triggerBindings.Bindings {
		if binding.Action.GetID() == actionID {
			return diag.FromErr(d.Set("display_name", binding.GetDisplayName()))
		}
	}

	d.SetId("")
	return nil
}

func deleteTriggerAction(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	trigger := d.Get("trigger").(string)
	actionID := d.Get("action_id").(string)

	triggerBindings, err := api.Action.Bindings(ctx, trigger)
	if err != nil {
		return diag.FromErr(err)
	}

	updatedBindings := make([]*management.ActionBinding, 0)
	for _, binding := range triggerBindings.Bindings {
		if binding.Action.GetID() == actionID {
			continue
		}

		updatedBindings = append(updatedBindings, &management.ActionBinding{
			Ref: &management.ActionBindingReference{
				Type:  auth0.String("action_id"),
				Value: binding.Action.ID,
			},
			DisplayName: binding.DisplayName,
		})
	}

	return diag.FromErr(api.Action.UpdateBindings(ctx, trigger, updatedBindings))
}
