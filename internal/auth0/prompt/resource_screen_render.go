package prompt

import (
	"context"
	"fmt"
	"strings"

	"github.com/auth0/go-auth0/management"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	allowedPromptsSettingsRenderer = []string{
		string(management.PromptSignupID),
		string(management.PromptSignupPassword),
		string(management.PromptLoginID),
		string(management.PromptLoginPassword),
		string(management.PromptLoginPasswordLess),
		string(management.PromptPhoneIdentifierEnrollment),
		string(management.PromptPhoneIdentifierChallenge),
		string(management.PromptEmailIdentifierChallenge),
		string(management.PromptCommon),
		string(management.PromptCaptcha),
	}
	allowedScreensSettingsRenderer = []string{
		string(management.ScreenSignupID),
		string(management.ScreenSignupPassword),
		string(management.ScreenLoginID),
		string(management.ScreenLoginPassword),
		string(management.ScreenLoginPasswordlessSMSOTP),
		string(management.ScreenLoginPasswordlessEmailCode),
		string(management.ScreenPhoneIdentifierEnrollment),
		string(management.ScreenPhoneIdentifierChallenge),
		string(management.ScreenEmailIdentifierChallenge),
		string(management.ScreenPasskeyEnrollment),
		string(management.ScreenPasskeyEnrollmentLocal),
		string(management.ScreenInterstitialCaptcha),
	}

	supportedRenderingModes = []string{string(management.RenderingModeStandard), string(management.RenderingModeAdvanced)}
)

// NewPromptScreenRenderResource will return a new auth0_prompt_screen_renderer resource.
func NewPromptScreenRenderResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPromptScreenRenderer,
		ReadContext:   readPromptScreenRenderer,
		UpdateContext: updatePromptScreenRenderer,
		DeleteContext: deletePromptScreenRenderer,
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
			"screen_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(allowedScreensSettingsRenderer, false),
				Description: "The screen that you are configuring settings for. " +
					"Options are: `" + strings.Join(allowedScreensSettingsRenderer, "`, `") + "`.",
			},
			"rendering_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      management.RenderingModeStandard,
				ValidateFunc: validation.StringInSlice(supportedRenderingModes, false),
				Description: "Rendering mode" +
					"Options are: `" + strings.Join(supportedRenderingModes, "`, `") + "`.",
			},
			"tenant": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Tenant ID",
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
			"head_tags": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description:      "An array of head tags",
			},
		},
	}
}

func createPromptScreenRenderer(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	promptName := data.Get("prompt_type").(string)
	screenName := data.Get("screen_name").(string)
	data.SetId(fmt.Sprintf("%s:%s", promptName, screenName))
	return updatePromptScreenRenderer(ctx, data, meta)
}

func readPromptScreenRenderer(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	promptScreenSettings, err := api.Prompt.ReadRendering(ctx, management.PromptType(strings.Split(data.Id(), ":")[0]), management.ScreenName(strings.Split(data.Id(), ":")[1]))
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(flattenPromptScreenSettings(data, promptScreenSettings))
}

func updatePromptScreenRenderer(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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

	return readPromptScreenRenderer(ctx, data, meta)
}

func deletePromptScreenRenderer(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
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
