package experimentfeatureflag

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewVariationDataSource returns a new auth0_experiment_feature_flag_variation data source.
func NewVariationDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readVariationForDataSource,
		Description: "Data source to retrieve a specific Experiment Center feature flag variation " +
			"by `feature_flag_id` and `variation_id`. (EA only)",
		Schema: variationDataSourceSchema(),
	}
}

func variationDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewVariationResource().Schema)
	internalSchema.SetExistingAttributesAsRequired(dataSourceSchema, "feature_flag_id", "variation_id")

	return dataSourceSchema
}

func readVariationForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	featureFlagID := data.Get("feature_flag_id").(string)
	variationID := data.Get("variation_id").(string)
	internalSchema.SetResourceGroupID(data, featureFlagID, variationID)

	variation, err := api.Experimentation.FeatureFlags.Variations.Get(ctx, featureFlagID, variationID)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenVariation(data, variation))
}
