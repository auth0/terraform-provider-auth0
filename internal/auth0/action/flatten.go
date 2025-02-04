package action

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenAction(data *schema.ResourceData, action *management.Action) error {
	result := multierror.Append(
		data.Set("name", action.GetName()),
		data.Set("supported_triggers", flattenActionTriggers(action.SupportedTriggers)),
		data.Set("code", action.GetCode()),
		data.Set("dependencies", flattenActionDependencies(action.GetDependencies())),
		data.Set("runtime", action.GetRuntime()),
	)

	// If custom-token-exchange is part of SupportedTriggers for an action,
	// we'd not manipulate it's runtime value.
	// This is done, to support node18 as runtime.
	// TODO: Remove this soon as node18 reaches EOL.

	if action.GetRuntime() == "node18-actions" {
		result = multierror.Append(result, data.Set("runtime", "node18"))
	}

	if action.GetDeployedVersion() != nil {
		result = multierror.Append(result, data.Set("version_id", action.GetDeployedVersion().GetID()))
	}

	return result.ErrorOrNil()
}

func flattenActionTriggers(triggers []management.ActionTrigger) []interface{} {
	var result []interface{}

	for _, trigger := range triggers {
		result = append(result, map[string]interface{}{
			"id":      trigger.GetID(),
			"version": trigger.GetVersion(),
		})
	}

	return result
}

func flattenActionDependencies(dependencies []management.ActionDependency) []interface{} {
	var result []interface{}

	for _, dependency := range dependencies {
		result = append(result, map[string]interface{}{
			"name":    dependency.GetName(),
			"version": dependency.GetVersion(),
		})
	}

	return result
}

func flattenTriggerBinding(data *schema.ResourceData, bindings []*management.ActionBinding) error {
	result := multierror.Append(
		data.Set("trigger", data.Id()),
		data.Set("actions", flattenTriggerBindingActions(bindings)),
	)

	return result.ErrorOrNil()
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
