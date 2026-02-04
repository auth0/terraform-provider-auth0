package action

import (
	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandActionModule(data *schema.ResourceData) *management.CreateActionModuleRequestContent {
	config := data.GetRawConfig()

	return &management.CreateActionModuleRequestContent{
		Name:         *value.String(config.GetAttr("name")),
		Code:         *value.String(config.GetAttr("code")),
		Dependencies: expandActionModuleDependencies(config.GetAttr("dependencies")),
		Secrets:      expandActionModuleSecrets(config.GetAttr("secrets")),
	}
}

func expandActionModuleUpdate(data *schema.ResourceData) *management.UpdateActionModuleRequestContent {
	config := data.GetRawConfig()

	module := &management.UpdateActionModuleRequestContent{
		Code: value.String(config.GetAttr("code")),
	}

	if data.HasChange("dependencies") {
		module.Dependencies = expandActionModuleDependencies(config.GetAttr("dependencies"))
	}

	if data.HasChange("secrets") {
		module.Secrets = expandActionModuleSecrets(config.GetAttr("secrets"))
	}

	return module
}

func expandActionModuleDependencies(dependencies cty.Value) []*management.ActionModuleDependencyRequest {
	if dependencies.IsNull() {
		return nil
	}

	actionModuleDependencies := make([]*management.ActionModuleDependencyRequest, 0)

	dependencies.ForEachElement(func(_ cty.Value, dep cty.Value) (stop bool) {
		actionModuleDependencies = append(actionModuleDependencies, &management.ActionModuleDependencyRequest{
			Name:    *value.String(dep.GetAttr("name")),
			Version: *value.String(dep.GetAttr("version")),
		})
		return stop
	})

	return actionModuleDependencies
}

func expandActionModuleSecrets(secrets cty.Value) []*management.ActionModuleSecretRequest {
	if secrets.IsNull() {
		return nil
	}

	actionModuleSecrets := make([]*management.ActionModuleSecretRequest, 0)

	secrets.ForEachElement(func(_ cty.Value, secret cty.Value) (stop bool) {
		actionModuleSecrets = append(actionModuleSecrets, &management.ActionModuleSecretRequest{
			Name:  *value.String(secret.GetAttr("name")),
			Value: *value.String(secret.GetAttr("value")),
		})
		return stop
	})

	return actionModuleSecrets
}
