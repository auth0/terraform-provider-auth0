package flow

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_flow data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readSelfServiceProfileForDataSource,
		Description: "Data source to retrieve a specific Auth0 Flow by `id`",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Flow.",
	}
	dataSourceSchema["id"].Description = "The id of the Flow."
	return dataSourceSchema
}

func readSelfServiceProfileForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	id := data.Get("id").(string)
	data.SetId(id)
	flow, err := api.Flow.Read(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenFlow(data, flow)
	return diag.FromErr(err)
}
