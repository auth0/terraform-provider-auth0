package prompt

import (
	"context"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0/management"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// NewPromptScreenRendersResource will return a new auth0_prompt_screen_renderers resource.
func NewPromptScreenRendersResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPromptScreenRenderers,
		ReadContext:   readPromptScreenRenderers,
		UpdateContext: updatePromptScreenRenderers,
		DeleteContext: deletePromptScreenRenderers,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can Configure the render settings for a specific screen." +
			"You can read more about this [here](https://auth0.com/docs/customize/login-pages/advanced-customizations/getting-started/configure-acul-screens).",
		Schema: map[string]*schema.Schema{
			"prompt_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(allowedPromptsSettingsRenderer, false),
				Description: "The prompt that you are configuring settings for. " +
					"Options are: `" + strings.Join(allowedPromptsSettingsRenderer, "`, `") + "`.",
			},
			"screen_renderers": {
				Type:     schema.TypeList,
				Optional: true,
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
			},
		},
	}
}

func createPromptScreenRenderers(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	promptName := data.Get("prompt_type").(string)
	data.SetId(promptName)
	return updatePromptScreenRenderers(ctx, data, meta)
}

func readPromptScreenRenderers(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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

func updatePromptScreenRenderers(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	prompt := management.PromptType(data.Get("prompt_type").(string))
	screen := management.ScreenName(data.Get("screen_name").(string))

	promptSettings, err := expandPromptSettings(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.Prompt.UpdateRendering(ctx, prompt, screen, promptSettings); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	if isFiltersNull(data) {
		if err := api.Request(ctx, http.MethodPatch, api.URI("prompts", string(prompt), "screen", string(screen), "rendering"), map[string]interface{}{"filters": nil}); err != nil {
			return diag.FromErr(err)
		}
	}

	return readPromptScreenRenderers(ctx, data, meta)
}

func deletePromptScreenRenderers(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	idComponents := strings.Split(data.Id(), ":")
	promptName, screenName := idComponents[0], idComponents[1]

	prompt := management.PromptType(promptName)
	screen := management.ScreenName(screenName)

	promptSettings := &management.PromptRendering{RenderingMode: &management.RenderingModeStandard}
	if err := api.Prompt.UpdateRendering(ctx, prompt, screen, promptSettings); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
