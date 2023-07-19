package client

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
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
	dataSourceSchema := internalSchema.TransformResourceToDataSource(internalSchema.Clone(NewResource().Schema))

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "name", "client_id")

	dataSourceSchema["name"].Description = "The name of the client. If not provided, `client_id` must be set."
	dataSourceSchema["client_id"].Description = "The ID of the client. If not provided, `name` must be set."

	dataSourceSchema["client_secret"] = &schema.Schema{
		Type:      schema.TypeString,
		Computed:  true,
		Sensitive: true,
		Description: "Secret for the client. Keep this private. To access this attribute you need to add the " +
			"`read:client_keys` scope to the Terraform client. Otherwise, the attribute will contain an empty string.",
	}

	dataSourceSchema["token_endpoint_auth_method"] = &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
		Description: "The authentication method for the token endpoint. " +
			"Results include `none` (public client without a client secret), " +
			"`client_secret_post` (client uses HTTP POST parameters), " +
			"`client_secret_basic` (client uses HTTP Basic). " +
			"Managing a client's authentication method can be done via the " +
			"`auth0_client_credentials` resource.",
	}

	return dataSourceSchema
}

func readClientForDataSource(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	clientID := d.Get("client_id").(string)
	if clientID != "" {
		d.SetId(clientID)

		client, err := api.Client.Read(ctx, d.Id())
		if err != nil {
			return diag.FromErr(err)
		}

		err = flattenClientForDataSource(d, client)

		return diag.FromErr(err)
	}

	name := d.Get("name").(string)
	if name == "" {
		return diag.Errorf("One of 'client_id' or 'name' is required.")
	}

	var page int
	for {
		clients, err := api.Client.List(
			ctx,
			management.Page(page),
		)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, client := range clients.Clients {
			if client.GetName() == name {
				d.SetId(client.GetClientID())
				err = flattenClientForDataSource(d, client)
				return diag.FromErr(err)
			}
		}

		if !clients.HasNext() {
			break
		}

		page++
	}

	return diag.Errorf("No client found with \"name\" = %q", name)
}
