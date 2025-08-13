package prompt

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"strings"

	"github.com/auth0/go-auth0/management"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewPromptScreenRendersDataSource creates a new data source to retrieve the prompt screen settings by `prompt_type`.
func NewPromptScreenRendersDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readPromptScreenRendersDataSource,
		Description: "Data source to retrieve a specific Auth0 prompt screen settings by `prompt_type` and `screen_name`",
		Schema:      getPromptScreenRendersDataSourceSchema(),
	}
}

func getPromptScreenRendersDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewPromptScreenRenderResource().Schema)
	internalSchema.SetExistingAttributesAsRequired(dataSourceSchema, "prompt_type")
	dataSourceSchema["prompt_type"].Description = "The type of prompt to customize."
	dataSourceSchema["prompt_type"].ValidateFunc = validation.StringInSlice(allowedPromptsSettingsRenderer, false)
	dataSourceSchema["screen_renderers"] = &schema.Schema{
		Type:        schema.TypeList,
		Description: "The screen name associated with the prompt type",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"screen_name": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "The name of the screen associated with the renderer",
					ValidateFunc: validation.StringInSlice(allowedScreensSettingsRenderer, false),
				},
				"rendering_mode": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      management.RenderingModeStandard,
					ValidateFunc: validation.StringInSlice(supportedRenderingModes, false),
					Description: "Rendering mode" +
						"Options are: `" + strings.Join(supportedRenderingModes, "`, `") + "`.",
				},

				"context_configuration": {
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Description: "Context values to make available",
				},
				"default_head_tags_disabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Override Universal Login default head tags",
				},
				"use_page_template": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Use page template with ACUL",
				},
				"filters": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Description: "Optional filters to apply rendering rules to specific entities. `match_type` and at least one of the entity arrays are required.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"match_type": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice([]string{"includes_any", "excludes_any"}, false),
								Description:  "Type of match to apply. Options: `includes_any`, `excludes_any`.",
							},
							"clients": {
								Type:             schema.TypeString,
								Optional:         true,
								Description:      "An array of clients (applications) identified by id or a metadata key/value pair. Entity Limit: 25.",
								ValidateFunc:     validation.StringIsJSON,
								DiffSuppressFunc: suppressUnorderedJSONDiff,
							},
							"organizations": {
								Type:             schema.TypeString,
								Optional:         true,
								Description:      "An array of organizations identified by id or a metadata key/value pair. Entity Limit: 25.",
								ValidateFunc:     validation.StringIsJSON,
								DiffSuppressFunc: suppressUnorderedJSONDiff,
							},
							"domains": {
								Type:             schema.TypeString,
								Optional:         true,
								Description:      "An array of domains identified by id or a metadata key/value pair. Entity Limit: 25.",
								ValidateFunc:     validation.StringIsJSON,
								DiffSuppressFunc: suppressUnorderedJSONDiff,
							},
						},
					},
				},
				"head_tags": {
					Type:             schema.TypeString,
					Optional:         true,
					Computed:         true,
					ValidateFunc:     validation.StringIsJSON,
					DiffSuppressFunc: structure.SuppressJsonDiff,
					Description:      "An array of head tags",
				},
				"tenant": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Tenant ID",
				},
			},
		},
	}

	return dataSourceSchema
}

func readPromptScreenRendersDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	promptType := data.Get("prompt_type").(string)

	screenSettings, err := api.Prompt.ListRendering(ctx, management.Parameter("prompt", promptType))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(promptType)

	if err := flattenPromptScreenRenderers(data, screenSettings); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(err)
}
