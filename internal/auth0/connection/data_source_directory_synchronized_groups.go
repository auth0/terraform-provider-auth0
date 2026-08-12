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
		Description: "Data source to retrieve the selected synchronized groups for a connection's directory provisioning configuration.",
		Schema:      getDirectorySynchronizedGroupsDataSourceSchema(),
	}
}

func getDirectorySynchronizedGroupsDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewDirectorySynchronizedGroupsResource().Schema)
	internalSchema.SetExistingAttributesAsRequired(dataSourceSchema, "connection_id")

	dataSourceSchema["query"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: "Filter the synchronized groups by a prefix search on a single field. " +
			"Only `name` and `email` are searchable, and only as a prefix, so the term must " +
			"take the form `name:<value>*` or `email:<value>*` (for example `name:engineering*`). " +
			"Returns all synchronized groups when omitted.",
	}

	const filteredBy = " Limited to the groups matching `query`, when one is given."
	dataSourceSchema["group_ids"].Description = "IDs of the synchronized Google Workspace Directory groups." + filteredBy
	dataSourceSchema["groups"].Description = "Details of the synchronized Google Workspace Directory groups." + filteredBy

	return dataSourceSchema
}

func readDirectorySynchronizedGroupsDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()
	connectionID := data.Get("connection_id").(string)

	groups, err := getAllGroups(ctx, apiv3, connectionID, data.Get("query").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(connectionID)

	result := multierror.Append(
		data.Set("connection_id", connectionID),
		data.Set("group_ids", flattenGroupIDs(groups)),
		data.Set("groups", flattenGroups(groups)),
	)

	return diag.FromErr(result.ErrorOrNil())
}
