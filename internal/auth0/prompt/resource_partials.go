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

// NewPartialsResource creates a new resource for partial prompts.
func NewPartialsResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPromptPartials,
		ReadContext:   readPromptPartials,
		UpdateContext: updatePromptPartials,
		DeleteContext: deletePromptPartials,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		DeprecationMessage: "This resource is deprecated and will be removed in the next major version. " +
			"Please use `auth0_prompt_screen_partials` for managing multiple prompt screens or `auth0_prompt_screen_partial` " +
			"for managing a single prompt screen, depending on your use case.",
		Description: "With this resource, you can manage a customized sign up and login experience by adding custom content, form elements and css/javascript. " +
			"You can read more about this [here](https://auth0.com/docs/customize/universal-login-pages/customize-signup-and-login-prompts).",
		Schema: map[string]*schema.Schema{
			"form_content_start": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Content that goes at the start of the form.",
				Deprecated: "This resource is deprecated and will be removed in the next major version. " +
					"Please use `auth0_prompt_screen_partials` for managing multiple prompt screens or `auth0_prompt_screen_partial` " +
					"for managing a single prompt screen, depending on your use case.",
			},
			"form_content_end": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Content that goes at the end of the form.",
				Deprecated: "This resource is deprecated and will be removed in the next major version. " +
					"Please use `auth0_prompt_screen_partials` for managing multiple prompt screens or `auth0_prompt_screen_partial` " +
					"for managing a single prompt screen, depending on your use case.",
			},
			"form_footer_start": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Footer content for the start of the footer.",
				Deprecated: "This resource is deprecated and will be removed in the next major version. " +
					"Please use `auth0_prompt_screen_partials` for managing multiple prompt screens or `auth0_prompt_screen_partial` " +
					"for managing a single prompt screen, depending on your use case.",
			},
			"form_footer_end": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Footer content for the end of the footer.",
				Deprecated: "This resource is deprecated and will be removed in the next major version. " +
					"Please use `auth0_prompt_screen_partials` for managing multiple prompt screens or `auth0_prompt_screen_partial` " +
					"for managing a single prompt screen, depending on your use case.",
			},
			"secondary_actions_start": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Actions that go at the start of secondary actions.",
				Deprecated: "This resource is deprecated and will be removed in the next major version. " +
					"Please use `auth0_prompt_screen_partials` for managing multiple prompt screens or `auth0_prompt_screen_partial` " +
					"for managing a single prompt screen, depending on your use case.",
			},
			"secondary_actions_end": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Actions that go at the end of secondary actions.",
				Deprecated: "This resource is deprecated and will be removed in the next major version. " +
					"Please use `auth0_prompt_screen_partials` for managing multiple prompt screens or `auth0_prompt_screen_partial` " +
					"for managing a single prompt screen, depending on your use case.",
			},
			"prompt": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(allowedPromptsWithPartials, false),
				Description: "The prompt that you are adding partials for. " +
					"Options are: `" + strings.Join(allowedPromptsWithPartials, "`, `") + "`.",
				Required: true,
				Deprecated: "This resource is deprecated and will be removed in the next major version. " +
					"Please use `auth0_prompt_screen_partials` for managing multiple prompt screens or `auth0_prompt_screen_partial` " +
					"for managing a single prompt screen, depending on your use case.",
			},
		},
	}
}

// Deprecated: createPromptPartials is deprecated and will be removed in the next major version.
func createPromptPartials(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	prompt := data.Get("prompt").(string)
	data.SetId(prompt)
	return updatePromptPartials(ctx, data, meta)
}

// Deprecated: readPromptPartials is deprecated and will be removed in the next major version.
func readPromptPartials(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	promptPartials, err := api.Prompt.ReadPartials(ctx, management.PromptType(data.Id()))
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenPromptPartials(data, promptPartials))
}

// Deprecated: updatePromptPartials is deprecated and will be removed in the next major version.
func updatePromptPartials(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	promptPartials := expandPromptPartials(data)

	if err := api.Prompt.UpdatePartials(ctx, promptPartials); err != nil {
		return diag.FromErr(err)
	}

	return readPromptPartials(ctx, data, meta)
}

// Deprecated: deletePromptPartials is deprecated and will be removed in the next major version.
func deletePromptPartials(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	prompt := data.Get("prompt").(string)

	if err := api.Prompt.DeletePartials(ctx, management.PromptType(prompt)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
