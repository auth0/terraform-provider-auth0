package experimentsegment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource returns a new auth0_experiment_segment data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readSegmentForDataSource,
		Description: "Data source to retrieve a specific Experiment Center segment by `id`. (EA only)",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)

	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the segment.",
	}

	return dataSourceSchema
}

func readSegmentForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	id := data.Get("id").(string)
	data.SetId(id)

	segment, err := api.Experimentation.Segments.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenSegment(data, segment))
}
