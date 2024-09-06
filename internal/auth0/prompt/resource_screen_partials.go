package prompt

import (
	"context"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

var allowedPromptsWithPartials = []string{
	string(management.PromptLoginID),
	string(management.PromptLogin),
	string(management.PromptLoginPassword),
	string(management.PromptSignup),
	string(management.PromptSignupID),
	string(management.PromptSignupPassword),
	string(management.PromptLoginPasswordLess),
}

// NewScreenPartialsResource will return a new auth0_prompt_screen_partials (1:many) resource.
func NewScreenPartialsResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPromptScreenPartials,
		ReadContext:   readPromptScreenPartials,
		UpdateContext: updatePromptScreenPartials,
		DeleteContext: deletePromptScreenPartials,
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
			"screen_partials": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"screen_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the screen associated with the partials",
						},
						"insertion_points": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
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
				},
			},
		},
	}
}

// createPromptScreenPartials creates a new prompt screen partials resource.
func createPromptScreenPartials(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	prompt := data.Get("prompt_type").(string)
	data.SetId(prompt)
	return updatePromptScreenPartials(ctx, data, meta)
}

// readPromptScreenPartials reads the prompt screen partials resource.
func readPromptScreenPartials(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	promptPartials, err := api.Prompt.GetPartials(ctx, management.PromptType(data.Id()))
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(flattenPromptScreenPartials(data, promptPartials))
}

// updatePromptScreenPartials updates the prompt screen partials resource.
func updatePromptScreenPartials(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	prompt := management.PromptType(data.Get("prompt_type").(string))
	promptPartials := expandPromptScreenPartials(data)
	if err := api.Prompt.SetPartials(ctx, prompt, promptPartials); err != nil {
		return diag.FromErr(err)
	}
	return readPromptScreenPartials(ctx, data, meta)
}

// deletePromptScreenPartials deletes the prompt screen partials resource.
func deletePromptScreenPartials(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	prompt := management.PromptType(data.Id())
	if err := api.Prompt.SetPartials(ctx, prompt, &management.PromptScreenPartials{}); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
