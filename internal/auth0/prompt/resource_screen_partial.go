package prompt

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewScreenPartialResource creates a new resource for prompt screen partial.
func NewScreenPartialResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPromptScreenPartial,
		ReadContext:   readPromptScreenPartial,
		UpdateContext: updatePromptScreenPartial,
		DeleteContext: deletePromptScreenPartial,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage a customized sign up and login experience by adding custom content, form elements and css/javascript. " +
			"You can read more about this [here](https://auth0.com/docs/customize/universal-login-pages/customize-signup-and-login-prompts).",
		Schema: map[string]*schema.Schema{
			"prompt_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(allowedPromptsWithPartials, false),
				Description: "The prompt that you are adding partials for. " +
					"Options are: `" + strings.Join(allowedPromptsWithPartials, "`, `") + "`.",
			},
			"screen_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the screen associated with the partials",
			},
			"insertion_points": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "The insertion points for the partials.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"form_content_start": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Content that goes at the start of the form.",
						},
						"form_content_end": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Content that goes at the end of the form.",
						},
						"form_footer_start": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Footer content for the start of the footer.",
						},
						"form_footer_end": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Footer content for the end of the footer.",
						},
						"secondary_actions_start": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Actions that go at the start of secondary actions.",
						},
						"secondary_actions_end": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Actions that go at the end of secondary actions.",
						},
					},
				},
			},
		},
	}
}

func createPromptScreenPartial(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	promptName := data.Get("prompt_type").(string)
	screenName := data.Get("screen_name").(string)
	data.SetId(fmt.Sprintf("%s:%s", promptName, screenName))
	return updatePromptScreenPartial(ctx, data, meta)
}

func readPromptScreenPartial(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	promptPartial, err := api.Prompt.GetPartials(ctx, management.PromptType(strings.Split(data.Id(), ":")[0]))
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(flattenPromptScreenPartial(data, promptPartial))
}

func updatePromptScreenPartial(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	prompt := management.PromptType(data.Get("prompt_type").(string))

	existingPromptScreenPartial, err := api.Prompt.GetPartials(ctx, prompt)
	if err != nil {
		return diag.FromErr(err)
	}
	if existingPromptScreenPartial == nil {
		existingPromptScreenPartial = &management.PromptScreenPartials{}
	}

	promptPartial := expandPromptScreenPartial(data)
	for screenName, insertionPoints := range *promptPartial {
		if existingInsertionPoints, exists := (*existingPromptScreenPartial)[screenName]; exists {
			for insertionPoint, content := range insertionPoints {
				existingInsertionPoints[insertionPoint] = content
			}
		} else {
			(*existingPromptScreenPartial)[screenName] = insertionPoints
		}
	}

	if err := api.Prompt.SetPartials(ctx, prompt, existingPromptScreenPartial); err != nil {
		return diag.FromErr(err)
	}
	return readPromptScreenPartial(ctx, data, meta)
}

func deletePromptScreenPartial(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	promptName, screenName := strings.Split(data.Id(), ":")[0], strings.Split(data.Id(), ":")[1]

	prompt := management.PromptType(promptName)

	existingPromptScreenPartial, err := api.Prompt.GetPartials(ctx, prompt)
	if err != nil {
		return diag.FromErr(err)
	}
	if existingPromptScreenPartial == nil {
		existingPromptScreenPartial = &management.PromptScreenPartials{}
	}

	delete(*existingPromptScreenPartial, management.ScreenName(screenName))

	if err := api.Prompt.SetPartials(ctx, prompt, existingPromptScreenPartial); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
