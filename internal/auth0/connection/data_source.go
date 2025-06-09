package connection

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_connection data source.
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

	dataSourceSchema["enabled_clients"] = &schema.Schema{
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
		Computed:    true,
		Description: "IDs of the clients for which the connection is enabled.",
	}

	return dataSourceSchema
}

func readConnectionForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)
	if connectionID != "" {
		connection, err := api.Connection.Read(ctx, connectionID)
		if err != nil {
			return diag.FromErr(err)
		}

		data.SetId(connectionID)

		existingClientsResp, err := api.Connection.ReadEnabledClients(ctx, connectionID)
		if err != nil {
			return diag.FromErr(err)
		}

		return flattenConnectionForDataSource(data, connection, existingClientsResp)
	}

	name := data.Get("name").(string)
	var from string

	options := []management.RequestOption{
		management.Take(100),
	}

	for {
		if from != "" {
			options = append(options, management.From(from))
		}

		connections, err := api.Connection.List(ctx, options...)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, connection := range connections.Connections {
			if connection.GetName() == name {
				data.SetId(connection.GetID())
				existingClientsResp, err := api.Connection.ReadEnabledClients(ctx, connection.GetID())
				if err != nil {
					return diag.FromErr(err)
				}

				return flattenConnectionForDataSource(data, connection, existingClientsResp)
			}
		}

		if !connections.HasNext() {
			break
		}

		from = connections.Next
	}

	return diag.Errorf("No connection found with \"name\" = %q", name)
}
