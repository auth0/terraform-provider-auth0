package action

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewTriggerActionResource will return a new auth0_trigger_action resource.
func NewTriggerActionResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createTriggerAction,
		ReadContext:   readTriggerAction,
		DeleteContext: deleteTriggerAction,
		Importer: &schema.ResourceImporter{
			StateContext: importTriggerAction,
		},
		Description: "With this resource, you can bind an action to a trigger. Once an action is created and deployed, it can be attached (i.e. bound) to a trigger so that it will be executed as part of a flow. The list of actions reflects the order in which they will be executed during the appropriate flow.\n\nOrdering of an action within a specific flow is not currently supported when using this resource; the action will get appended to the end of the flow. To precisely manage ordering, it is advised to either do so with the dashboard UI or with the `auth0_trigger_bindings` resource.",
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
		},
	}
}

func createTriggerAction(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	trigger := d.Get("trigger").(string)
	actionID := d.Get("action_id").(string)

	currentBindings, err := api.Action.Bindings(trigger)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	t := "action_id"

	var updatedBindings []*management.ActionBinding
	for _, binding := range currentBindings.Bindings {
		if binding.Action.ID == &actionID {
			d.SetId(trigger + "::" + actionID)
			return nil
		}
		updatedBindings = append(updatedBindings, &management.ActionBinding{
			Ref: &management.ActionBindingReference{
				Type:  &t,
				Value: binding.Action.ID,
			},
		})
	}

	action, err := api.Action.Read(actionID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	updatedBindings = append(updatedBindings, &management.ActionBinding{
		Ref: &management.ActionBindingReference{
			Type:  &t,
			Value: &actionID,
		},
		DisplayName: action.Name,
	})

	if err := api.Action.UpdateBindings(trigger, updatedBindings); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(trigger + "::" + actionID)
	return nil
}

func readTriggerAction(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	trigger := d.Get("trigger").(string)
	actionID := d.Get("action_id").(string)

	triggerBindings, err := api.Action.Bindings(trigger)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	for _, binding := range triggerBindings.Bindings {
		if binding.Action.GetID() == actionID {
			d.SetId(trigger + "::" + actionID)
			return nil
		}
	}

	d.SetId(trigger + "::" + actionID)
	return nil
}

func deleteTriggerAction(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	trigger := d.Get("trigger").(string)
	actionID := d.Get("action_id").(string)

	triggerBindings, err := api.Action.Bindings(trigger)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var updatedBindings []*management.ActionBinding
	for _, binding := range triggerBindings.Bindings {
		if binding.Action.GetID() == actionID {
			continue
		}

		t := "action_id"
		id := binding.Action.GetID()
		displayName := binding.GetDisplayName()
		updatedBindings = append(updatedBindings, &management.ActionBinding{
			Ref: &management.ActionBindingReference{
				Type:  &t,
				Value: &id,
			},
			DisplayName: &displayName,
		})
	}

	if err := api.Action.UpdateBindings(trigger, updatedBindings); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func importTriggerAction(
	_ context.Context,
	data *schema.ResourceData,
	_ interface{},
) ([]*schema.ResourceData, error) {
	rawID := data.Id()
	if rawID == "" {
		return nil, fmt.Errorf("ID cannot be empty")
	}

	if !strings.Contains(rawID, "::") {
		return nil, fmt.Errorf("ID must be formatted as <trigger>::<actionID>")
	}

	idPair := strings.Split(rawID, "::")
	if len(idPair) != 2 {
		return nil, fmt.Errorf("ID must be formatted as <trigger>::<actionID>")
	}

	result := multierror.Append(
		data.Set("trigger", idPair[0]),
		data.Set("action_id", idPair[1]),
	)

	return []*schema.ResourceData{data}, result.ErrorOrNil()
}
