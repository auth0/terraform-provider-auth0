package prompt

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewPromptScreenPartialsDataSource creates a new data source to retrieve a specific Auth0 prompt screen partials by `prompt_type`.
func NewPromptScreenPartialsDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readPromptScreenPartialsDataSource,
		Description: "Data source to retrieve a specific Auth0 prompt screen partials by `prompt_type`.",
		Schema:      getScreenPartialsDataSourceSchema(),
	}
}

func getScreenPartialsDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewScreenPartialsResource().Schema)
	internalSchema.SetExistingAttributesAsRequired(dataSourceSchema, "prompt_type")
	dataSourceSchema["prompt_type"].Description = "The type of prompt to customize."
	dataSourceSchema["prompt_type"].ValidateFunc = validation.StringInSlice(allowedPromptsWithPartials, false)
	dataSourceSchema["screen_partials"].Description = "The screen partials associated with the prompt type."
	dataSourceSchema["screen_partials"].Optional = true
	return dataSourceSchema
}

func readPromptScreenPartialsDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	promptType := data.Get("prompt_type").(string)
	screenPartials, err := api.Prompt.GetPartials(ctx, management.PromptType(promptType))
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(promptType)
	if err := flattenPromptScreenPartials(data, screenPartials); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(err)
}
