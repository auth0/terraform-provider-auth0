package user

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_user data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readUserForDataSource,
		Description: "Data source to retrieve a specific Auth0 user by `user_id`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(internalSchema.Clone(NewResource().Schema))

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "user_id")
	dataSourceSchema["user_id"].Required = true
	dataSourceSchema["user_id"].Computed = false
	dataSourceSchema["user_id"].Optional = false

	dataSourceSchema["permissions"].Deprecated = ""
	dataSourceSchema["permissions"].Description = "List of API permissions granted to the user."
	dataSourceSchema["roles"].Deprecated = ""
	dataSourceSchema["roles"].Description = "Set of IDs of roles assigned to the user."

	return dataSourceSchema
}

func readUserForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	userID := data.Get("user_id").(string)
	data.SetId(userID)
	return readUser(ctx, data, meta)
}
