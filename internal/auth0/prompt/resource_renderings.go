package prompt

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewPromptRenderingsResource creates a new resource for bulk managing prompt rendering settings.
func NewPromptRenderingsResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPromptRenderings,
		ReadContext:   readPromptRenderings,
		UpdateContext: updatePromptRenderings,
		DeleteContext: deletePromptRenderings,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage multiple prompt rendering settings in bulk using the PATCH /api/v2/prompts/rendering endpoint. " +
			"This allows you to configure rendering settings for multiple prompt screens efficiently. " +
			"You can read more about this [here](https://auth0.com/docs/customize/login-pages/advanced-customizations/getting-started/configure-acul-screens).",
		Schema: map[string]*schema.Schema{
			"renderings": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of prompt rendering configurations to manage in bulk.",
				Elem: &schema.Resource{
					Schema: getRenderingItemSchema(),
				},
			},
		},
	}
}

func getRenderingItemSchema() map[string]*schema.Schema {
	baseSchema := NewPromptScreenRenderResource().Schema

	itemSchema := make(map[string]*schema.Schema)

	for key, value := range baseSchema {
		switch key {
		case "prompt_type":
			schemaCopy := *value
			schemaCopy.Description = "The prompt type."
			itemSchema["prompt"] = &schemaCopy
		case "screen_name":
			schemaCopy := *value
			schemaCopy.Description = "The screen name."
			itemSchema["screen"] = &schemaCopy
		default:
			itemSchema[key] = value
		}
	}

	return itemSchema
}

func createPromptRenderings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id := generateBulkRenderingsID(data)
	data.SetId(id)
	return updatePromptRenderings(ctx, data, meta)
}

func readPromptRenderings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	renderings := data.Get("renderings").([]interface{})
	if len(renderings) == 0 {
		return nil
	}

	// Fetch all renderings in a single API call.
	existingRenderings, err := api.Prompt.ListRendering(ctx)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	// Flatten and set the renderings based on configured items.
	if err := flattenPromptRenderingsFromList(data, existingRenderings, renderings); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updatePromptRenderings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	renderings := data.Get("renderings").([]interface{})
	if len(renderings) == 0 {
		return nil
	}

	bulkUpdates := expandPromptRenderingsForBulkUpdate(renderings)

	if err := api.Request(ctx, "PATCH", api.URI("prompts", "rendering"), map[string]interface{}{"configs": bulkUpdates}); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	for _, renderingRaw := range renderings {
		renderingMap := renderingRaw.(map[string]interface{})
		prompt := renderingMap["prompt"].(string)
		screen := renderingMap["screen"].(string)

		filters, hasFilters := renderingMap["filters"]
		shouldClearFilters := !hasFilters
		if hasFilters {
			if filtersList, ok := filters.([]interface{}); ok {
				shouldClearFilters = len(filtersList) == 0 || filtersList[0] == nil
			}
		}

		if shouldClearFilters {
			if err := api.Request(ctx, http.MethodPatch, api.URI("prompts", prompt, "screen", screen, "rendering"), map[string]interface{}{"filters": nil}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return readPromptRenderings(ctx, data, meta)
}

func deletePromptRenderings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	renderings := data.Get("renderings").([]interface{})
	if len(renderings) == 0 {
		return nil
	}

	// Build bulk reset configs - set all to standard mode.
	var bulkResets []map[string]interface{}
	for _, renderingRaw := range renderings {
		renderingMap := renderingRaw.(map[string]interface{})
		bulkResets = append(bulkResets, map[string]interface{}{
			"prompt":         renderingMap["prompt"].(string),
			"screen":         renderingMap["screen"].(string),
			"rendering_mode": "standard",
		})
	}

	// Use bulk PATCH to reset all renderings to standard mode.
	if err := api.Request(ctx, "PATCH", api.URI("prompts", "rendering"), map[string]interface{}{"configs": bulkResets}); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func generateBulkRenderingsID(_ *schema.ResourceData) string {
	// Use a static ID for bulk renderings resource since the ID should remain
	// consistent regardless of which prompt/screen combinations are included.
	// This allows users to add/remove renderings without forcing resource recreation.
	return "prompt-renderings-bulk"
}
