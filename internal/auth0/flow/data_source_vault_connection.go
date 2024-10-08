package flow

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewVaultConnectionDataSource will return a new auth0_flow_vault_connection data source.
func NewVaultConnectionDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readVaultConnectionForDataSource,
		Description: "Data source to retrieve a specific Auth0 Flow Vault Connection by `id`",
		Schema:      getVaultConnectionDataSourceSchema(),
	}
}

func getVaultConnectionDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewVaultConnectionResource().Schema)
	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Flow Vault Connection.",
	}
	dataSourceSchema["id"].Description = "The id of the Flow Vault Connection."
	return dataSourceSchema
}

func readVaultConnectionForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	id := data.Get("id").(string)
	data.SetId(id)
	vaultConnection, err := api.Flow.Vault.GetConnection(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenVaultConnection(data, vaultConnection)
	return diag.FromErr(err)
}
