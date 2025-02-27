package action

import (
	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
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
					"password-reset-post-challenge",
					"custom-token-exchange",
					"custom-email-provider",
				}, false),
				Description: "The ID of the trigger to bind with. Available options: `post-login`, `credentials-exchange`, `pre-user-registration`, `post-user-registration`, `post-change-password`, `send-phone-message`, `password-reset-post-challenge`, `custom-token-exchange`, `custom-email-provider`.",
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

func createTriggerAction(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	trigger := data.Get("trigger").(string)
	actionID := data.Get("action_id").(string)
	displayName := data.Get("display_name").(string)

	currentBindings, err := api.Action.Bindings(ctx, trigger)
	if err != nil {
		return diag.FromErr(err)
	}

	var updatedBindings []*management.ActionBinding
	for _, binding := range currentBindings.Bindings {
		if binding.Action.GetID() == actionID {
			internalSchema.SetResourceGroupID(data, trigger, actionID)
			return readTriggerAction(ctx, data, meta)
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

	internalSchema.SetResourceGroupID(data, trigger, actionID)

	return readTriggerAction(ctx, data, meta)
}

func updateTriggerAction(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	trigger := data.Get("trigger").(string)
	actionID := data.Get("action_id").(string)
	displayName := data.Get("display_name").(string)

	var currentBindings []*management.ActionBinding
	var page int
	for {
		triggerBindingList, err := api.Action.Bindings(
			ctx,
			trigger,
			management.Page(page),
			management.PerPage(100),
		)
		if err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}

		currentBindings = append(currentBindings, triggerBindingList.Bindings...)

		if !triggerBindingList.HasNext() {
			break
		}

		page++
	}

	found := false
	var updatedBindings []*management.ActionBinding
	for _, binding := range currentBindings {
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
		data.SetId("")
		return nil
	}

	if err := api.Action.UpdateBindings(ctx, trigger, updatedBindings); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readTriggerAction(ctx, data, meta)
}

func readTriggerAction(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	trigger := data.Get("trigger").(string)
	actionID := data.Get("action_id").(string)

	var triggerBindings []*management.ActionBinding
	var page int
	for {
		triggerBindingList, err := api.Action.Bindings(
			ctx,
			trigger,
			management.Page(page),
			management.PerPage(100),
		)
		if err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}

		triggerBindings = append(triggerBindings, triggerBindingList.Bindings...)

		if !triggerBindingList.HasNext() {
			break
		}

		page++
	}

	for _, binding := range triggerBindings {
		if binding.Action.GetID() == actionID {
			return diag.FromErr(data.Set("display_name", binding.GetDisplayName()))
		}
	}

	data.SetId("")
	return nil
}

func deleteTriggerAction(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	trigger := data.Get("trigger").(string)
	actionID := data.Get("action_id").(string)

	triggerBindings, err := api.Action.Bindings(ctx, trigger)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
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

	if err = api.Action.UpdateBindings(ctx, trigger, updatedBindings); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
