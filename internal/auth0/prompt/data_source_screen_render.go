package prompt

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewPromptScreenRenderDataSource creates a new data source to retrieve the prompt and screen settings`.
func NewPromptScreenRenderDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readPromptScreenRenderDataSource,
		Description: "Data source to retrieve a specific Auth0 prompt screen settings by `prompt_type` and `screen_name`",
		Schema:      getPromptScreenRenderDataSourceSchema(),
	}
}

func getPromptScreenRenderDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewPromptScreenRenderResource().Schema)
	internalSchema.SetExistingAttributesAsRequired(dataSourceSchema, "prompt_type", "screen_name")
	dataSourceSchema["prompt_type"].Description = "The type of prompt to customize."
	dataSourceSchema["prompt_type"].ValidateFunc = validation.StringInSlice(allowedPromptsSettingsRenderer, false)
	dataSourceSchema["screen_name"].Description = "The screen name associated with the prompt type."
	dataSourceSchema["screen_name"].ValidateFunc = validation.StringInSlice(allowedScreensSettingsRenderer, false)
	return dataSourceSchema
}

func readPromptScreenRenderDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	prompt := management.PromptType(data.Get("prompt_type").(string))
	screen := management.ScreenName(data.Get("screen_name").(string))

	screenSettings, err := api.Prompt.ReadRendering(ctx, prompt, screen)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(fmt.Sprintf("%s:%s", prompt, screen))

	if err := flattenPromptScreenSettings(data, screenSettings); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(err)
}
