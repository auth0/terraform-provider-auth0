package auth0

import (
	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newDataClient() *schema.Resource {
	return &schema.Resource{
		ReadContext: readDataClient,
		Schema:      newClientSchema(),
	}
}

func newClientSchema() map[string]*schema.Schema {
	clientSchema := datasourceSchemaFromResourceSchema(newClient().Schema)
	delete(clientSchema, "client_secret_rotation_trigger")
	addOptionalFieldsToSchema(clientSchema, "name", "client_id")
	return clientSchema
}

func readDataClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	clientId := auth0.StringValue(String(d, "client_id"))
	if clientId != "" {
		d.SetId(clientId)
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
			clientId = auth0.StringValue(client.ClientID)
			d.SetId(clientId)
			return readClient(ctx, d, m)
		}
	}
	return diag.Errorf("no client found with 'name' = '%s'", name)
}
