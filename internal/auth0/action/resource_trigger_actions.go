package action

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewTriggerActionsResource will return a new auth0_trigger_actions resource.
func NewTriggerActionsResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createTriggerBinding,
		ReadContext:   readTriggerBinding,
		UpdateContext: updateTriggerBinding,
		DeleteContext: deleteTriggerBinding,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can bind actions to a trigger. Once actions are created " +
			"and deployed, they can be attached (i.e. bound) to a trigger so that it will be executed as " +
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
					"password-reset-post-challenge",
					"iga-approval",
					"iga-certification",
					"iga-fulfillment-assignment",
					"iga-fulfillment-execution",
				}, false),
				Description: "The ID of the trigger to bind with. Options include: `post-login`, `credentials-exchange`, " +
					"`pre-user-registration`, `post-user-registration`, `post-change-password`, `send-phone-message`, " +
					"`password-reset-post-challenge`, `iga-approval` , `iga-certification` , `iga-fulfillment-assignment`, " +
					"`iga-fulfillment-execution`.",
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
							Description: "The display name of the action within the flow.",
						},
					},
				},
				Description: "The list of actions bound to this trigger.",
			},
		},
	}
}

func createTriggerBinding(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	id := data.Get("trigger").(string)
	triggerBindings := expandTriggerBindings(data.GetRawConfig().GetAttr("actions"))

	if err := api.Action.UpdateBindings(ctx, id, triggerBindings); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(id)

	return readTriggerBinding(ctx, data, meta)
}

func readTriggerBinding(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	triggerBindings, err := api.Action.Bindings(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenTriggerBinding(data, triggerBindings.Bindings))
}

func updateTriggerBinding(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	triggerBindings := expandTriggerBindings(data.GetRawConfig().GetAttr("actions"))

	if err := api.Action.UpdateBindings(ctx, data.Id(), triggerBindings); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readTriggerBinding(ctx, data, meta)
}

func deleteTriggerBinding(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Action.UpdateBindings(ctx, data.Id(), []*management.ActionBinding{}); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
