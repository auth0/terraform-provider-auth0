package provider

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

func newDataClient() *schema.Resource {
	clientDataSource := &schema.Resource{
		ReadContext: readDataClient,
		Schema:      newClientSchema(),
		Description: "Data source to retrieve a specific Auth0 Application client by 'client_id' or 'name'.",
	}

	internalSchema.SetExistingAttributesAsOptional(clientDataSource.Schema, "name", "client_id")
	clientDataSource.Schema["name"].Description = "The name of the client. If not provided, `client_id` must be set."
	clientDataSource.Schema["client_id"].Description = "The ID of the client. If not provided, `name` must be set."

	return clientDataSource
}

func newClientSchema() map[string]*schema.Schema {
	clientSchema := internalSchema.TransformResourceToDataSource(newClient().Schema)
	delete(clientSchema, "client_secret_rotation_trigger")
	return clientSchema
}

func readDataClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
