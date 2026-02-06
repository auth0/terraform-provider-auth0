package action

import (
	"context"

	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewModuleActionsDataSource will return a new auth0_action_module_actions data source.
func NewModuleActionsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readActionModuleActionsForDataSource,
		Description: "Data source to retrieve all actions that are using a specific Auth0 action module.",
		Schema: map[string]*schema.Schema{
			"module_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the action module.",
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The total number of actions using this module.",
			},
			"actions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of actions using this module.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the action.",
						},
						"action_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the action.",
						},
						"module_version_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the module version this action is using.",
						},
						"module_version_number": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The version number of the module this action is using.",
						},
						"supported_triggers": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "The triggers that this action supports.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The trigger ID.",
									},
									"version": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The trigger version.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func readActionModuleActionsForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()
	moduleID := data.Get("module_id").(string)

	actionsPage, err := apiv2.Actions.Modules.ListActions(ctx, moduleID, &management.GetActionModuleActionsRequestParameters{})
	if err != nil {
		return diag.FromErr(err)
	}

	// Collect all actions using the iterator.
	var allActions []*management.ActionModuleAction
	iterator := actionsPage.Iterator()
	for iterator.Next(ctx) {
		allActions = append(allActions, iterator.Current())
	}
	if err := iterator.Err(); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(moduleID)

	return diag.FromErr(flattenActionModuleActions(data, allActions, actionsPage.Response.GetTotal()))
}

func flattenActionModuleActions(data *schema.ResourceData, actions []*management.ActionModuleAction, total int) error {
	result := multierror.Append(
		data.Set("total", total),
		data.Set("actions", flattenActionModuleActionsList(actions)),
	)

	return result.ErrorOrNil()
}

func flattenActionModuleActionsList(actions []*management.ActionModuleAction) []interface{} {
	var result []interface{}

	for _, action := range actions {
		actionMap := map[string]interface{}{
			"action_id":             action.GetActionID(),
			"action_name":           action.GetActionName(),
			"module_version_id":     action.GetModuleVersionID(),
			"module_version_number": action.GetModuleVersionNumber(),
			"supported_triggers":    flattenActionModuleActionTriggers(action.GetSupportedTriggers()),
		}

		result = append(result, actionMap)
	}

	return result
}

func flattenActionModuleActionTriggers(triggers []*management.ActionTrigger) []interface{} {
	var result []interface{}

	for _, trigger := range triggers {
		triggerMap := map[string]interface{}{
			"id":      trigger.GetID(),
			"version": trigger.GetVersion(),
		}

		result = append(result, triggerMap)
	}

	return result
}
