package resourceserver

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_resource_server data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readResourceServerForDataSource,
		Description: "Data source to retrieve a specific Auth0 resource server by `resource_server_id` or `identifier`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["resource_server_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the resource server. If not provided, `identifier` must be set.",
		AtLeastOneOf: []string{"resource_server_id", "identifier"},
	}

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "identifier")
	dataSourceSchema["identifier"].Description = "The unique identifier for the resource server. " +
		"If not provided, `resource_server_id` must be set."
	dataSourceSchema["identifier"].AtLeastOneOf = []string{"resource_server_id", "identifier"}

	return dataSourceSchema
}

func readResourceServerForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceServerID := data.Get("resource_server_id").(string)
	if resourceServerID != "" {
		data.SetId(resourceServerID)
		return readResourceServer(ctx, data, meta)
	}

	data.SetId(url.PathEscape(data.Get("identifier").(string)))
	return readResourceServer(ctx, data, meta)
}
