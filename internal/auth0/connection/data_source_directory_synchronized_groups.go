package connection

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDirectorySynchronizedGroupsDataSource will return a new auth0_connection_directory_synchronized_groups data source.
func NewDirectorySynchronizedGroupsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readDirectorySynchronizedGroupsDataSource,
		Description: "Data source to retrieve the selected synchronized group IDs for a connection's directory provisioning configuration.",
		Schema:      getDirectorySynchronizedGroupsDataSourceSchema(),
	}
}

func getDirectorySynchronizedGroupsDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewDirectorySynchronizedGroupsResource().Schema)
	internalSchema.SetExistingAttributesAsRequired(dataSourceSchema, "connection_id")

	return dataSourceSchema
}

func readDirectorySynchronizedGroupsDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()
	connectionID := data.Get("connection_id").(string)

	groupIDs, err := getAllSynchronizedGroups(ctx, apiv2, connectionID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(connectionID)

	result := multierror.Append(
		data.Set("connection_id", connectionID),
		data.Set("group_ids", groupIDs),
	)

	return diag.FromErr(result.ErrorOrNil())
}
