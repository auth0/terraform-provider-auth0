package client

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewGlobalResource will return a new auth0_global_client resource.
func NewGlobalResource() *schema.Resource {
	client := NewResource()
	client.Description = "Use a tenant's global Auth0 Application client."
	client.CreateContext = createGlobalClient
	client.DeleteContext = deleteGlobalClient

	exclude := []string{"client_secret_rotation_trigger"}

	// Mark all values computed and optional,
	// because the global client has already
	// been created for all tenants.
	for key := range client.Schema {
		// Exclude certain fields from
		// being marked as computed.
		if in(key, exclude) {
			continue
		}

		client.Schema[key].Required = false
		client.Schema[key].Optional = true
		client.Schema[key].Computed = true
	}

	return client
}

func in(needle string, haystack []string) bool {
	for i := 0; i < len(haystack); i++ {
		if needle == haystack[i] {
			return true
		}
	}
	return false
}

func createGlobalClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := readGlobalClientID(ctx, d, m); err != nil {
		return err
	}
	return updateClient(ctx, d, m)
}

func readGlobalClientID(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	clients, err := api.Client.List(
		management.Parameter("is_global", "true"),
		management.IncludeFields("client_id"),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(clients.Clients) == 0 {
		return diag.Errorf("No auth0 global client found.")
	}

	d.SetId(clients.Clients[0].GetClientID())
	return nil
}

func deleteGlobalClient(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}
