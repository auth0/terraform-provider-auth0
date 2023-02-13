package role

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

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

	return dataSourceSchema
}

func readRoleForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	roleID := data.Get("role_id").(string)
	if roleID != "" {
		data.SetId(roleID)
		return readRole(ctx, data, meta)
	}

	api := meta.(*management.Management)
	name := data.Get("name").(string)
	page := 0
	for {
		roles, err := api.Role.List(
			management.Page(page),
			management.PerPage(100),
			management.Parameter("name_filter", name),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, role := range roles.Roles {
			if role.GetName() == name {
				data.SetId(role.GetID())
				return readRole(ctx, data, meta)
			}
		}

		if !roles.HasNext() {
			break
		}

		page++
	}

	return diag.Errorf("No role found with \"name\" = %q", name)
}
