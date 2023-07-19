package user

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func flattenUser(data *schema.ResourceData, user *management.User) (err error) {
	userMetadata, err := structure.FlattenJsonToString(user.GetUserMetadata())
	if err != nil {
		return err
	}

	appMetadata, err := structure.FlattenJsonToString(user.GetAppMetadata())
	if err != nil {
		return err
	}

	result := multierror.Append(
		data.Set("user_id", user.GetID()),
		data.Set("username", user.GetUsername()),
		data.Set("name", user.GetName()),
		data.Set("family_name", user.GetFamilyName()),
		data.Set("given_name", user.GetGivenName()),
		data.Set("nickname", user.GetNickname()),
		data.Set("email", user.GetEmail()),
		data.Set("email_verified", user.GetEmailVerified()),
		data.Set("verify_email", user.GetVerifyEmail()),
		data.Set("phone_number", user.GetPhoneNumber()),
		data.Set("phone_verified", user.GetPhoneVerified()),
		data.Set("blocked", user.GetBlocked()),
		data.Set("picture", user.GetPicture()),
		data.Set("user_metadata", userMetadata),
		data.Set("app_metadata", appMetadata),
	)

	return result.ErrorOrNil()
}

func flattenUserForDataSource(
	data *schema.ResourceData,
	user *management.User,
	roles []*management.Role,
	permissions []*management.Permission,
) error {
	result := multierror.Append(
		flattenUser(data, user),
		data.Set("roles", flattenUserRolesSlice(roles)),
		data.Set("permissions", flattenUserPermissionsSlice(permissions)),
	)

	return result.ErrorOrNil()
}

func flattenUserPermissions(data *schema.ResourceData, permissions []*management.Permission) error {
	result := multierror.Append(
		data.Set("user_id", data.Id()),
		data.Set("permissions", flattenUserPermissionsSlice(permissions)),
	)

	return result.ErrorOrNil()
}

func flattenUserPermissionsSlice(permissions []*management.Permission) []interface{} {
	var userPermissions []interface{}
	for _, permission := range permissions {
		userPermissions = append(userPermissions, map[string]string{
			"name":                       permission.GetName(),
			"resource_server_identifier": permission.GetResourceServerIdentifier(),
			"description":                permission.GetDescription(),
			"resource_server_name":       permission.GetResourceServerName(),
		})
	}
	return userPermissions
}

func flattenUserPermission(data *schema.ResourceData, permission *management.Permission) error {
	result := multierror.Append(
		data.Set("description", permission.GetDescription()),
		data.Set("resource_server_name", permission.GetResourceServerName()),
	)

	return result.ErrorOrNil()
}

func flattenUserRole(data *schema.ResourceData, role *management.Role) error {
	result := multierror.Append(
		data.Set("role_name", role.GetName()),
		data.Set("role_description", role.GetDescription()),
	)

	return result.ErrorOrNil()
}

func flattenUserRoles(data *schema.ResourceData, roles []*management.Role) error {
	result := multierror.Append(
		data.Set("user_id", data.Id()),
		data.Set("roles", flattenUserRolesSlice(roles)),
	)

	return result.ErrorOrNil()
}

func flattenUserRolesSlice(roles []*management.Role) []string {
	var userRoles []string
	for _, role := range roles {
		userRoles = append(userRoles, role.GetID())
	}
	return userRoles
}
