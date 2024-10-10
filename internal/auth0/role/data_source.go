package role

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_role data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readRoleForDataSource,
		Description: "Data source to retrieve a specific Auth0 role by `role_id` or `name`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["role_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the role. If not provided, `name` must be set.",
		AtLeastOneOf: []string{"role_id", "name"},
	}

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "name")
	dataSourceSchema["name"].Description = "The name of the role. If not provided, `role_id` must be set."
	dataSourceSchema["name"].AtLeastOneOf = []string{"role_id", "name"}

	dataSourceSchema["permissions"] = &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "Configuration settings for permissions (scopes) attached to the role.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of the permission (scope) configured on the resource server (API).",
				},
				"resource_server_identifier": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Unique identifier for the resource server (API).",
				},
				"description": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Description of the permission.",
				},
				"resource_server_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of resource server (API) that the permission is associated with.",
				},
			},
		},
	}

	dataSourceSchema["users"] = &schema.Schema{
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Computed:    true,
		Description: "List of user IDs assigned to this role. Retrieves a maximum of 1000 user IDs.",
	}

	return dataSourceSchema
}

func readRoleForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	roleID := data.Get("role_id").(string)
	if roleID != "" {
		data.SetId(roleID)
		return readRoleByID(ctx, data, api, roleID)
	}

	roleName := data.Get("name").(string)
	return readRoleByName(ctx, data, api, roleName)
}

func readRoleByID(
	ctx context.Context,
	data *schema.ResourceData,
	api *management.Management,
	roleID string,
) diag.Diagnostics {
	role, err := api.Role.Read(ctx, roleID)
	if err != nil {
		return diag.FromErr(err)
	}

	permissions, err := getAllRolePermissions(ctx, api, roleID)
	if err != nil {
		return diag.FromErr(err)
	}

	users, err := getAllRoleUsers(ctx, api, role.GetID())
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenRoleForDataSource(data, role, permissions, users))
}

func readRoleByName(
	ctx context.Context,
	data *schema.ResourceData,
	api *management.Management,
	roleName string,
) diag.Diagnostics {
	page := 0
	for {
		roles, err := api.Role.List(
			ctx,
			management.Page(page),
			management.PerPage(100),
			management.Parameter("name_filter", roleName),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, role := range roles.Roles {
			if role.GetName() == roleName {
				data.SetId(role.GetID())

				permissions, err := getAllRolePermissions(ctx, api, role.GetID())
				if err != nil {
					return diag.FromErr(err)
				}

				users, err := getAllRoleUsers(ctx, api, role.GetID())
				if err != nil {
					return diag.FromErr(err)
				}

				return diag.FromErr(flattenRoleForDataSource(data, role, permissions, users))
			}
		}

		if !roles.HasNext() {
			break
		}

		page++
	}

	return diag.Errorf("No role found with \"name\" = %q", roleName)
}

func getAllRolePermissions(
	ctx context.Context,
	api *management.Management,
	roleID string,
) ([]*management.Permission, error) {
	var permissions []*management.Permission
	var page int
	for {
		permissionList, err := api.Role.Permissions(ctx, roleID, management.Page(page), management.PerPage(100))
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permissionList.Permissions...)

		if !permissionList.HasNext() {
			break
		}

		page++
	}

	return permissions, nil
}

func getAllRoleUsers(
	ctx context.Context,
	api *management.Management,
	roleID string,
) ([]*management.User, error) {
	var users []*management.User
	var from string

	options := []management.RequestOption{
		management.Take(100),
	}

	for {
		if from != "" {
			options = append(options, management.From(from))
		}

		userList, err := api.Role.Users(ctx, roleID, options...)
		if err != nil {
			return nil, err
		}

		users = append(users, userList.Users...)
		if !userList.HasNext() {
			break
		}

		from = userList.Next
	}

	return users, nil
}
