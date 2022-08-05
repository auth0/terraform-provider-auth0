package provider

import (
	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newDataClient() *schema.Resource {
	clientDataSource := &schema.Resource{
		ReadContext: readDataClient,
		Schema:      newClientSchema(),
		Description: "Data source to retrieve a specific Auth0 Application client by 'client_id' or 'name'.",
	}

	addOptionalFieldsToSchema(clientDataSource.Schema, "name", "client_id")
	clientDataSource.Schema["name"].Description = "The name of the client. If not provided, `client_id` must be set."
	clientDataSource.Schema["client_id"].Description = "The ID of the client. If not provided, `name` must be set."

	return clientDataSource
}

func newClientSchema() map[string]*schema.Schema {
	clientSchema := dataSourceSchemaFromResourceSchema(newClient().Schema)
	delete(clientSchema, "client_secret_rotation_trigger")
	return clientSchema
}

func readDataClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientID := auth0.StringValue(String(d, "client_id"))
	if clientID != "" {
		d.SetId(clientID)
		return readClient(ctx, d, m)
	}

	// If not provided ID, perform looking of client by name
	name := auth0.StringValue(String(d, "name"))
	if name == "" {
		return diag.Errorf("no 'client_id' or 'name' was specified")
	}

	api := m.(*management.Management)
	clients, err := api.Client.List(management.IncludeFields("client_id", "name"))
	if err != nil {
		return diag.FromErr(err)
	}
	for _, client := range clients.Clients {
		if auth0.StringValue(client.Name) == name {
			clientID = auth0.StringValue(client.ClientID)
			d.SetId(clientID)
			return readClient(ctx, d, m)
		}
	}
	return diag.Errorf("no client found with 'name' = '%s'", name)
}
