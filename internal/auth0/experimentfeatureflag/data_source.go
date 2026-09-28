package experimentfeatureflag

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource returns a new auth0_experiment_feature_flag data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readFeatureFlagForDataSource,
		Description: "Data source to retrieve a specific Experiment Center feature flag by `id`. (EA only)",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)

	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the feature flag.",
	}

	return dataSourceSchema
}

func readFeatureFlagForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	id := data.Get("id").(string)
	data.SetId(id)

	flag, err := api.Experimentation.FeatureFlags.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	// A data source reports every variation on the flag, so list them all rather than the
	// state-tracked subset the resource read uses.
	variations, err := api.Experimentation.FeatureFlags.Variations.List(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenFeatureFlag(data, flag, variations.GetVariations()))
}
