package eventstream

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource returns a new auth0_event_stream data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readEventStreamForDataSource,
		Description: "Data source to retrieve a specific Auth0 Event Stream by `id`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)

	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Event Stream.",
	}

	return dataSourceSchema
}

func readEventStreamForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	id := data.Get("id").(string)
	data.SetId(id)

	stream, err := api.EventStream.Read(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenEventStream(data, stream)
	return diag.FromErr(err)
}
