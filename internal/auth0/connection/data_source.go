package connection

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/auth0"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_connection_client data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readConnectionForDataSource,
		Description: "Data source to retrieve a specific Auth0 connection by `connection_id` or `name`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(internalSchema.Clone(NewResource().Schema))
	dataSourceSchema["connection_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the connection. If not provided, `name` must be set.",
		AtLeastOneOf: []string{"connection_id", "name"},
	}

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "name")
	dataSourceSchema["name"].Description = "The name of the connection. If not provided, `connection_id` must be set."
	dataSourceSchema["name"].AtLeastOneOf = []string{"connection_id", "name"}

	dataSourceSchema["enabled_clients"].Deprecated = ""
	dataSourceSchema["enabled_clients"].Description = "IDs of the clients for which the connection is enabled."

	return dataSourceSchema
}

func readConnectionForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionID := data.Get("connection_id").(string)
	if connectionID != "" {
		data.SetId(connectionID)
		return auth0.CheckFor404Error(ctx, readConnection, data, meta)
	}

	api := meta.(*config.Config).GetAPI()
	name := data.Get("name").(string)
	page := 0
	for {
		connections, err := api.Connection.List(
			ctx,
			management.IncludeFields("id", "name"),
			management.Page(page),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, connection := range connections.Connections {
			if connection.GetName() == name {
				data.SetId(connection.GetID())
				return auth0.CheckFor404Error(ctx, readConnection, data, meta)
			}
		}

		if !connections.HasNext() {
			break
		}

		page++
	}

	return diag.Errorf("No connection found with \"name\" = %q", name)
}
