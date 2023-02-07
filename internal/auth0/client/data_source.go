package client

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_client data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readClientForDataSource,
		Description: "Data source to retrieve a specific Auth0 application client by `client_id` or `name`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)

	delete(dataSourceSchema, "client_secret_rotation_trigger")

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "name", "client_id")

	dataSourceSchema["name"].Description = "The name of the client. If not provided, `client_id` must be set."
	dataSourceSchema["client_id"].Description = "The ID of the client. If not provided, `name` must be set."

	return dataSourceSchema
}

func readClientForDataSource(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID := d.Get("client_id").(string)
	if clientID != "" {
		d.SetId(clientID)
		return readClient(ctx, d, m)
	}

	name := d.Get("name").(string)
	if name == "" {
		return diag.Errorf("One of 'client_id' or 'name' is required.")
	}

	api := m.(*management.Management)

	var page int
	for {
		clients, err := api.Client.List(
			management.IncludeFields("client_id", "name"),
			management.Page(page),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, client := range clients.Clients {
			if client.GetName() == name {
				d.SetId(client.GetClientID())
				return readClient(ctx, d, m)
			}
		}

		if !clients.HasNext() {
			break
		}

		page++
	}

	return diag.Errorf("No client found with \"name\" = %q", name)
}
