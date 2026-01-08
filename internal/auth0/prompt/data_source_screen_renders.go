package prompt

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewPromptRenderingsDataSource creates a new data source to retrieve all prompt rendering settings with filtering.
func NewPromptRenderingsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readPromptRenderingsDataSource,
		Description: "Data source to retrieve Auth0 prompt rendering settings with optional filtering by prompt, screen, and rendering_mode. Supports pagination.",
		Schema:      getPromptRenderingsDataSourceSchema(),
	}
}

func getPromptRenderingsDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := map[string]*schema.Schema{
		"prompt": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filter by prompt type name (supports wildcards). Leave empty to retrieve all prompts.",
		},
		"screen": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filter by screen name (supports wildcards). Leave empty to retrieve all screens.",
		},
		"rendering_mode": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "Filter by rendering mode. Options are: `standard`, `advanced`.",
			ValidateFunc: validation.StringInSlice(supportedRenderingModes, false),
		},
		"renderings": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of prompt rendering settings matching the filter criteria.",
			Elem: &schema.Resource{
				Schema: internalSchema.TransformResourceToDataSource(getRenderingItemSchema()),
			},
		},
	}

	return dataSourceSchema
}

func readPromptRenderingsDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	var options []management.RequestOption

	if prompt, ok := data.GetOk("prompt"); ok {
		options = append(options, management.Parameter("prompt", prompt.(string)))
	}

	if screen, ok := data.GetOk("screen"); ok {
		options = append(options, management.Parameter("screen", screen.(string)))
	}

	if renderingMode, ok := data.GetOk("rendering_mode"); ok {
		options = append(options, management.Parameter("rendering_mode", renderingMode.(string)))
	}

	renderingList, err := api.Prompt.ListRendering(ctx, options...)
	if err != nil {
		return diag.FromErr(err)
	}

	// Generate a unique ID for the data source.
	data.SetId("prompt-renderings-bulk-" + id.UniqueId())

	if err := flattenPromptRenderingsList(data, renderingList); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
