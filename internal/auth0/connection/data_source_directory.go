package connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDirectoryDataSource will return a new auth0_connection_directory data source.
func NewDirectoryDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readDirectoryDataSource,
		Description: "Data source to retrieve directory provisioning configuration for an Auth0 connection by `connection_id`.",
		Schema:      getDirectoryDataSourceSchema(),
	}
}

func getDirectoryDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewDirectoryResource().Schema)
	internalSchema.SetExistingAttributesAsRequired(dataSourceSchema, "connection_id")

	return dataSourceSchema
}

func readDirectoryDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	connectionID := data.Get("connection_id").(string)
	directoryConfig, err := apiv2.Connections.DirectoryProvisioning.Get(ctx, connectionID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(connectionID)

	return flattenDirectory(data, directoryConfig)
}
