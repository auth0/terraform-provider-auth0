package action

import (
	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenActionModule(data *schema.ResourceData, module *management.GetActionModuleResponseContent) error {
	result := multierror.Append(
		data.Set("name", module.GetName()),
		data.Set("code", module.GetCode()),
		data.Set("dependencies", flattenActionModuleDependencies(module.GetDependencies())),
		data.Set("actions_using_module_total", module.GetActionsUsingModuleTotal()),
		data.Set("all_changes_published", module.GetAllChangesPublished()),
		data.Set("latest_version_number", module.GetLatestVersionNumber()),
		data.Set("latest_version", flattenActionModuleLatestVersion(module.LatestVersion)),
		data.Set("secrets", flattenActionModuleSecretsWithValue(data, module.GetSecrets())),
	)

	return result.ErrorOrNil()
}

// flattenActionModuleSecretsWithValue flattens secrets from API response while preserving
// the value from the config (since the API doesn't return secret values).
func flattenActionModuleSecretsWithValue(data *schema.ResourceData, secretsFromAPI []*management.ActionModuleSecret) []interface{} {
	// Build a map of secret values from config.
	configSecrets := make(map[string]string)
	if secretsSet, ok := data.GetOk("secrets"); ok {
		for _, s := range secretsSet.(*schema.Set).List() {
			secretMap := s.(map[string]interface{})
			name := secretMap["name"].(string)
			if value, ok := secretMap["value"].(string); ok {
				configSecrets[name] = value
			}
		}
	}

	var result []interface{}
	for _, secret := range secretsFromAPI {
		secretMap := map[string]interface{}{
			"name": secret.GetName(),
		}

		// Preserve value from config.
		if value, ok := configSecrets[secret.GetName()]; ok {
			secretMap["value"] = value
		}

		if secret.UpdatedAt != nil {
			secretMap["updated_at"] = secret.UpdatedAt.String()
		}

		result = append(result, secretMap)
	}

	return result
}

func flattenActionModuleLatestVersion(latestVersion *management.ActionModuleVersionReference) []interface{} {
	if latestVersion == nil {
		return nil
	}

	result := map[string]interface{}{
		"id":             latestVersion.GetID(),
		"version_number": latestVersion.GetVersionNumber(),
		"code":           latestVersion.GetCode(),
		"dependencies":   flattenActionModuleDependencies(latestVersion.GetDependencies()),
		"secrets":        flattenActionModuleSecrets(latestVersion.GetSecrets()),
	}

	if latestVersion.CreatedAt != nil {
		result["created_at"] = latestVersion.CreatedAt.String()
	}

	return []interface{}{result}
}

func flattenActionModuleSecrets(secrets []*management.ActionModuleSecret) []interface{} {
	var result []interface{}

	for _, secret := range secrets {
		secretMap := map[string]interface{}{
			"name": secret.GetName(),
		}

		if secret.UpdatedAt != nil {
			secretMap["updated_at"] = secret.UpdatedAt.String()
		}

		result = append(result, secretMap)
	}

	return result
}

func flattenActionModuleDependencies(dependencies []*management.ActionModuleDependency) []interface{} {
	var result []interface{}

	for _, dependency := range dependencies {
		result = append(result, map[string]interface{}{
			"name":    dependency.GetName(),
			"version": dependency.GetVersion(),
		})
	}

	return result
}
