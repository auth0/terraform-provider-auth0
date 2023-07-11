package page

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_page data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readPagesForDataSource,
		Description: "Use this data source to access the HTML for the login, reset password, multi-factor authentication and error pages.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	return dataSourceSchema
}

func readPagesForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// This resource is not identified by an id in the Auth0 management API.
	data.SetId(id.UniqueId())
	return readPages(ctx, data, meta)
}
