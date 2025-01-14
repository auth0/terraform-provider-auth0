package organization

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenOrganization(data *schema.ResourceData, organization *management.Organization) error {
	result := multierror.Append(
		data.Set("name", organization.GetName()),
		data.Set("display_name", organization.GetDisplayName()),
		data.Set("branding", flattenOrganizationBranding(organization.GetBranding())),
		data.Set("metadata", organization.GetMetadata()),
	)

	return result.ErrorOrNil()
}

func flattenOrganizationForDataSource(
	data *schema.ResourceData,
	organization *management.Organization,
	connections []*management.OrganizationConnection,
	members []management.OrganizationMember,
	clientGrants []*management.ClientGrant,
) error {
	result := multierror.Append(
		flattenOrganization(data, organization),
		data.Set("connections", flattenOrganizationEnabledConnections(connections)),
		data.Set("members", flattenOrganizationMembersSlice(members)),
		data.Set("client_grants", flattenOrganizationClientGrantsSlice(clientGrants)),
	)

	return result.ErrorOrNil()
}

func flattenOrganizationBranding(organizationBranding *management.OrganizationBranding) []interface{} {
	if organizationBranding == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"logo_url": organizationBranding.GetLogoURL(),
			"colors":   organizationBranding.GetColors(),
		},
	}
}

func flattenOrganizationConnection(data *schema.ResourceData, orgConn *management.OrganizationConnection) error {
	result := multierror.Append(
		data.Set("assign_membership_on_login", orgConn.GetAssignMembershipOnLogin()),
		data.Set("is_signup_enabled", orgConn.GetIsSignupEnabled()),
		data.Set("show_as_button", orgConn.GetShowAsButton()),
		data.Set("name", orgConn.GetConnection().GetName()),
		data.Set("strategy", orgConn.GetConnection().GetStrategy()),
	)

	return result.ErrorOrNil()
}

func flattenOrganizationConnections(data *schema.ResourceData, connections []*management.OrganizationConnection) error {
	result := multierror.Append(
		data.Set("organization_id", data.Id()),
		data.Set("enabled_connections", flattenOrganizationEnabledConnections(connections)),
	)

	return result.ErrorOrNil()
}

func flattenOrganizationEnabledConnections(connections []*management.OrganizationConnection) []interface{} {
	if connections == nil {
		return nil
	}

	result := make([]interface{}, len(connections))
	for index, connection := range connections {
		result[index] = map[string]interface{}{
			"connection_id":              connection.GetConnectionID(),
			"assign_membership_on_login": connection.GetAssignMembershipOnLogin(),
			"is_signup_enabled":          connection.GetIsSignupEnabled(),
			"show_as_button":             connection.GetShowAsButton(),
		}
	}

	return result
}

func flattenOrganizationMemberRole(data *schema.ResourceData, role management.OrganizationMemberRole) error {
	result := multierror.Append(
		data.Set("role_name", role.GetName()),
		data.Set("role_description", role.GetDescription()),
	)

	return result.ErrorOrNil()
}

func flattenOrganizationMembers(data *schema.ResourceData, members []management.OrganizationMember) error {
	result := multierror.Append(
		data.Set("organization_id", data.Id()),
		data.Set("members", flattenOrganizationMembersSlice(members)),
	)

	return result.ErrorOrNil()
}

func flattenOrganizationMembersSlice(members []management.OrganizationMember) []string {
	if len(members) == 0 {
		return nil
	}
	flattenedMembers := make([]string, 0)
	for _, member := range members {
		flattenedMembers = append(flattenedMembers, member.GetUserID())
	}

	return flattenedMembers
}

func flattenOrganizationClientGrantsSlice(clientGrants []*management.ClientGrant) []string {
	if len(clientGrants) == 0 {
		return nil
	}
	flattenedClientGrants := make([]string, 0)
	for _, grant := range clientGrants {
		flattenedClientGrants = append(flattenedClientGrants, grant.GetID())
	}
	return flattenedClientGrants
}
