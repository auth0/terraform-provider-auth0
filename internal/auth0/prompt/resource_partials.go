package prompt

import (
	"context"
	"fmt"
	"github.com/auth0/go-auth0/management"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

var availablePartialsPrompts = []string{
	string(management.PartialsPromptLoginID),
	string(management.PartialsPromptLogin),
	string(management.PartialsPromptLoginPassword),
	string(management.PartialsPromptSignup),
	string(management.PartialsPromptSignupID),
	string(management.PartialsPromptSignupPassword),
}

// NewPartialsResource creates a new resource for partial prompts.
func NewPartialsResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPartialsPrompt,
		ReadContext:   readPartialsPrompt,
		UpdateContext: updatePartialsPrompt,
		DeleteContext: deletePartialsPrompt,
		Description:   "With Auth0, you can use a custom sign in to maintain a consistent user experience. This resource allows you to create and manage a custom sign in within your Auth0 tenant.",
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
			"prompt": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(availablePartialsPrompts, false),
				Description: "The prompt that you are adding partials for. " +
					"Options are: `" + strings.Join(availablePartialsPrompts, "`, `") + "`.",
				Required: true,
			},
		},
	}
}
func createPartialsPrompt(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	return updatePartialsPrompt(ctx, data, meta)
}

func readPartialsPrompt(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	segment, ok := data.GetOk("prompt")
	if !ok {
		return diag.FromErr(fmt.Errorf("failed, missing prompt"))
	}

	prompt, err := api.Prompt.ReadPartials(ctx, management.PartialsPromptSegment(segment.(string)))
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("form_content_start", prompt.FormContentStart),
		data.Set("form_content_end", prompt.FormContentEnd),
		data.Set("form_footer_start", prompt.FormFooterStart),
		data.Set("form_footer_end", prompt.FormFooterEnd),
		data.Set("secondary_actions_start", prompt.SecondaryActionsStart),
		data.Set("secondary_actions_end", prompt.SecondaryActionsEnd),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updatePartialsPrompt(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	segment, ok := data.GetOk("prompt")
	if !ok {
		return diag.FromErr(fmt.Errorf("failed, missing prompt"))
	}

	prompt := &management.PartialsPrompt{
		Segment:               management.PartialsPromptSegment(segment.(string)),
		FormContentStart:      data.Get("form_content_start").(string),
		FormContentEnd:        data.Get("form_content_end").(string),
		FormFooterStart:       data.Get("form_footer_start").(string),
		FormFooterEnd:         data.Get("form_footer_end").(string),
		SecondaryActionsStart: data.Get("secondary_actions_start").(string),
		SecondaryActionsEnd:   data.Get("secondary_actions_end").(string),
	}

	if err := api.Prompt.UpdatePartials(ctx, prompt); err != nil {
		return diag.FromErr(err)
	}

	return readPartialsPrompt(ctx, data, meta)
}

func deletePartialsPrompt(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	segment, ok := data.GetOk("prompt")
	if !ok {
		return diag.FromErr(fmt.Errorf("failed, missing prompt"))
	}

	prompt := &management.PartialsPrompt{
		Segment:               management.PartialsPromptSegment(segment.(string)),
		FormContentStart:      data.Get("form_content_start").(string),
		FormContentEnd:        data.Get("form_content_end").(string),
		FormFooterStart:       data.Get("form_footer_start").(string),
		FormFooterEnd:         data.Get("form_footer_end").(string),
		SecondaryActionsStart: data.Get("secondary_actions_start").(string),
		SecondaryActionsEnd:   data.Get("secondary_actions_end").(string),
	}

	if err := api.Prompt.DeletePartials(ctx, prompt); err != nil {
		return diag.FromErr(err)
	}

	return readPartialsPrompt(ctx, data, meta)
}
