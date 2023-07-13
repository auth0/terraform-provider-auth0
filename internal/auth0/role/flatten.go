package role

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenRole(data *schema.ResourceData, role *management.Role, permissions []*management.Permission) error {
	result := multierror.Append(
		data.Set("name", role.GetName()),
		data.Set("description", role.GetDescription()),
		data.Set("permissions", flattenRolePermissions(permissions)),
	)

	return result.ErrorOrNil()
}

func flattenRolePermissions(permissions []*management.Permission) []interface{} {
	var result []interface{}
	for _, permission := range permissions {
		result = append(result, map[string]interface{}{
			"name":                       permission.GetName(),
			"description":                permission.GetDescription(),
			"resource_server_identifier": permission.GetResourceServerIdentifier(),
			"resource_server_name":       permission.GetResourceServerName(),
		})
	}
	return result
}
