package role

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenRole(data *schema.ResourceData, role *management.Role) error {
	result := multierror.Append(
		data.Set("name", role.GetName()),
		data.Set("description", role.GetDescription()),
	)

	return result.ErrorOrNil()
}

func flattenRoleForDataSource(data *schema.ResourceData, role *management.Role, permissions []*management.Permission) error {
	result := multierror.Append(
		flattenRole(data, role),
		data.Set("permissions", flattenRolePermissionsSlice(permissions)),
	)

	return result.ErrorOrNil()
}

func flattenRolePermissions(data *schema.ResourceData, permissions []*management.Permission) error {
	result := multierror.Append(
		data.Set("role_id", data.Id()),
		data.Set("permissions", flattenRolePermissionsSlice(permissions)),
	)

	return result.ErrorOrNil()
}

func flattenRolePermissionsSlice(permissions []*management.Permission) []interface{} {
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

func flattenRolePermission(data *schema.ResourceData, permission *management.Permission) error {
	result := multierror.Append(
		data.Set("description", permission.GetDescription()),
		data.Set("resource_server_name", permission.GetResourceServerName()),
	)

	return result.ErrorOrNil()
}
