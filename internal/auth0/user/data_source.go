package user

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_user data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readUserForDataSource,
		Description: "Data source to retrieve a specific Auth0 user by `user_id` or by `lucene query`. " +
			"If filtered by Lucene Query, it should include sufficient filters to retrieve a unique user.",
		Schema: dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(internalSchema.Clone(NewResource().Schema))

	dataSourceSchema["user_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "ID of the user.",
		AtLeastOneOf: []string{"user_id", "query"},
	}

	dataSourceSchema["query"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "Lucene Query for retrieving a user.",
		AtLeastOneOf: []string{"user_id", "query"},
	}

	dataSourceSchema["permissions"] = &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "List of API permissions granted to the user.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of the permission.",
				},
				"description": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Description of the permission.",
				},
				"resource_server_identifier": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Resource server identifier associated with the permission.",
				},
				"resource_server_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of resource server that the permission is associated with.",
				},
			},
		},
	}

	dataSourceSchema["roles"] = &schema.Schema{
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Description: "Set of IDs of roles assigned to the user.",
	}

	return dataSourceSchema
}

func readUserForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	var user *management.User

	userID := data.Get("user_id").(string)
	if userID != "" {
		u, err := api.User.Read(ctx, userID)
		if err != nil {
			return diag.FromErr(err)
		}
		user = u
		data.SetId(user.GetID())
	} else {
		query := data.Get("query").(string)
		users, err := api.User.List(ctx, management.Parameter("q", query))
		if err != nil {
			return diag.FromErr(err)
		}

		// The data-source retrieves the roles and permissions for a user.
		// Hence, it is important the search bottoms out to a single user.
		// If multiple users are retrieved via Lucene Query, we prompt the user to add further filters.
		if users.Length == 1 {
			user = users.Users[0]
			data.SetId(users.Users[0].GetID())
		} else {
			return diag.Errorf("Further improve the query to retrieve a single user from the query")
		}
	}

	// Populate Roles for the retrieved User.
	var roles []*management.Role
	var rolesPage int
	for {
		roleList, err := api.User.Roles(ctx, user.GetID(), management.Page(rolesPage), management.PerPage(100))
		if err != nil {
			return diag.FromErr(err)
		}

		roles = append(roles, roleList.Roles...)

		if !roleList.HasNext() {
			break
		}

		rolesPage++
	}

	// Populate Permissions for the retrieved User.
	var permissions []*management.Permission
	var permissionsPage int
	for {
		permissionList, err := api.User.Permissions(ctx, user.GetID(), management.Page(permissionsPage), management.PerPage(100))
		if err != nil {
			return diag.FromErr(err)
		}

		permissions = append(permissions, permissionList.Permissions...)

		if !permissionList.HasNext() {
			break
		}

		permissionsPage++
	}

	return diag.FromErr(flattenUserForDataSource(data, user, roles, permissions))
}
