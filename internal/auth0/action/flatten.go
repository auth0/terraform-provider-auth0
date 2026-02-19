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
		data.Set("modules", flattenActionModulesForAction(data, action.GetModules())),
	)

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

// flattenActionModulesForAction flattens modules from API response while preserving
// the module_id and module_version_id from the config (since they are required inputs).
func flattenActionModulesForAction(data *schema.ResourceData, modulesFromAPI []management.ActionModules) []interface{} {
	// Build a map of module config values.
	configModules := make(map[string]map[string]string)
	if modulesSet, ok := data.GetOk("modules"); ok {
		for _, m := range modulesSet.(*schema.Set).List() {
			moduleMap := m.(map[string]interface{})
			moduleID := moduleMap["module_id"].(string)
			moduleVersionID := moduleMap["module_version_id"].(string)
			configModules[moduleID] = map[string]string{
				"module_id":         moduleID,
				"module_version_id": moduleVersionID,
			}
		}
	}

	var result []interface{}
	for _, module := range modulesFromAPI {
		moduleMap := map[string]interface{}{
			"module_id":             module.GetModuleID(),
			"module_name":           module.GetModuleName(),
			"module_version_id":     module.GetModuleVersionID(),
			"module_version_number": module.GetModuleVersionNumber(),
		}

		// Preserve module_version_id from config if available.
		if configModule, ok := configModules[module.GetModuleID()]; ok {
			moduleMap["module_version_id"] = configModule["module_version_id"]
		}

		result = append(result, moduleMap)
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
