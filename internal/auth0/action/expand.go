package action

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandAction(data *schema.ResourceData) *management.Action {
	config := data.GetRawConfig()

	action := &management.Action{
		Name:              value.String(config.GetAttr("name")),
		Code:              value.String(config.GetAttr("code")),
		Runtime:           value.String(config.GetAttr("runtime")),
		SupportedTriggers: expandActionTriggers(config.GetAttr("supported_triggers")),
	}

	if data.HasChange("dependencies") {
		action.Dependencies = expandActionDependencies(config.GetAttr("dependencies"))
	}

	if data.HasChange("secrets") {
		action.Secrets = expandActionSecrets(config.GetAttr("secrets"))
	}

	// If custom-token-exchange is part of SupportedTriggers for an action,
	// we'd not manipulate it's runtime value.
	// This is done, to support node18 as runtime.
	// TODO: Remove this soon as node18 reaches EOL.
	if action.GetRuntime() == "node18" && !isTokenExchangeInSupportedTriggers(action.SupportedTriggers) {
		action.Runtime = auth0.String("node18-actions")
	}

	return action
}

func isTokenExchangeInSupportedTriggers(actionTriggers []management.ActionTrigger) bool {
	for _, actionTrigger := range actionTriggers {
		if actionTrigger.GetID() == "custom-token-exchange" {
			return true
		}
	}
	return false
}

func expandActionTriggers(triggers cty.Value) []management.ActionTrigger {
	if triggers.IsNull() {
		return nil
	}

	supportedTriggers := make([]management.ActionTrigger, 0)

	triggers.ForEachElement(func(_ cty.Value, triggers cty.Value) (stop bool) {
		supportedTriggers = append(supportedTriggers, management.ActionTrigger{
			ID:      value.String(triggers.GetAttr("id")),
			Version: value.String(triggers.GetAttr("version")),
		})
		return stop
	})

	return supportedTriggers
}

func expandActionDependencies(dependencies cty.Value) *[]management.ActionDependency {
	actionDependencies := make([]management.ActionDependency, 0)

	dependencies.ForEachElement(func(_ cty.Value, dep cty.Value) (stop bool) {
		actionDependencies = append(actionDependencies, management.ActionDependency{
			Name:    value.String(dep.GetAttr("name")),
			Version: value.String(dep.GetAttr("version")),
		})
		return stop
	})

	return &actionDependencies
}

func expandActionSecrets(secrets cty.Value) *[]management.ActionSecret {
	actionSecrets := make([]management.ActionSecret, 0)

	secrets.ForEachElement(func(_ cty.Value, secret cty.Value) (stop bool) {
		actionSecrets = append(actionSecrets, management.ActionSecret{
			Name:  value.String(secret.GetAttr("name")),
			Value: value.String(secret.GetAttr("value")),
		})
		return stop
	})

	return &actionSecrets
}

func expandTriggerBindings(config cty.Value) []*management.ActionBinding {
	var triggerBindings []*management.ActionBinding

	config.ForEachElement(func(_ cty.Value, action cty.Value) (stop bool) {
		t := "action_id"
		triggerBindings = append(triggerBindings, &management.ActionBinding{
			Ref: &management.ActionBindingReference{
				Type:  &t,
				Value: value.String(action.GetAttr("id")),
			},
			DisplayName: value.String(action.GetAttr("display_name")),
		})
		return stop
	})

	return triggerBindings
}

func preventErasingUnmanagedSecrets(ctx context.Context, data *schema.ResourceData, api *management.Management) diag.Diagnostics {
	if !data.HasChange("secrets") {
		return nil
	}

	preUpdateAction, err := api.Action.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	// Extract changes to secrets from the resource data.
	oldSecrets, newSecrets := data.GetChange("secrets")

	// Stores the old and secrets from *schema.Set to slices of interface{}.
	var secretsList []interface{}

	if oldSecrets != nil {
		secretsList = append(secretsList, oldSecrets.(*schema.Set).List()...)
	}

	if newSecrets != nil {
		secretsList = append(secretsList, newSecrets.(*schema.Set).List()...)
	}

	// Pass allSecrets to check for unmanaged action secrets.
	return checkForUnmanagedActionSecrets(secretsList, preUpdateAction.GetSecrets())
}

func checkForUnmanagedActionSecrets(
	secretsFromConfig []interface{},
	secretsFromAPI []management.ActionSecret,
) diag.Diagnostics {
	secretKeysInConfigMap := make(map[string]bool, len(secretsFromConfig))
	for _, secret := range secretsFromConfig {
		// Check if the element can be asserted as a map.
		secretMap, ok := secret.(map[string]interface{})
		if !ok {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid Configuration Format",
					Detail:   "Secrets configuration contains improperly formatted elements. Each secret must be a map with 'name' and 'value'.",
				},
			}
		}

		// Safely extract the "name" field from the secret map.
		secretKeyName, nameOk := secretMap["name"].(string)
		if !nameOk || secretKeyName == "" {
			return diag.Diagnostics{
				diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid Secret Name",
					Detail:   "Each secret in the configuration must have a valid 'name' as a string.",
				},
			}
		}

		secretKeysInConfigMap[secretKeyName] = true
	}

	var diagnostics diag.Diagnostics
	for _, secret := range secretsFromAPI {
		if _, ok := secretKeysInConfigMap[secret.GetName()]; !ok {
			diagnostics = append(diagnostics, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unmanaged Action Secret",
				Detail: fmt.Sprintf("Detected an action secret not managed though Terraform: %s. If you proceed, "+
					"this secret will get deleted. It is required to add this secret to your action configuration "+
					"to prevent unintentionally destructive results.",
					secret.GetName(),
				),
				AttributePath: cty.Path{cty.GetAttrStep{Name: "secrets"}},
			})
		}
	}

	return diagnostics
}
