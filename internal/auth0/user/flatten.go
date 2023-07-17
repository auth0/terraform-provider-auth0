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

func flattenUserRoles(roleList *management.RoleList) []interface{} {
	var roles []interface{}
	for _, role := range roleList.Roles {
		roles = append(roles, role.GetID())
	}
	return roles
}

func flattenUserPermissions(permissionList *management.PermissionList) []interface{} {
	var permissions []interface{}
	for _, permission := range permissionList.Permissions {
		permissions = append(permissions, map[string]string{
			"name":                       permission.GetName(),
			"resource_server_identifier": permission.GetResourceServerIdentifier(),
			"description":                permission.GetDescription(),
			"resource_server_name":       permission.GetResourceServerName(),
		})
	}
	return permissions
}
