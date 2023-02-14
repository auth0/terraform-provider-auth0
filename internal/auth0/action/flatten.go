package action

import (
	"github.com/auth0/go-auth0/management"
)

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
