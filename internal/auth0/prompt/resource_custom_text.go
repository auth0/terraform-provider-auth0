package prompt

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

var (
	availablePrompts = []string{
		"capcha", "common", "consent", "custom-form", "customized-consent", "device-flow", "email-otp-challenge",
		"email-verification", "invitation", "login", "login-id", "login-password", "login-passwordless",
		"login-email-verification", "logout", "mfa", "mfa-email", "mfa-otp", "mfa-phone", "mfa-push", "mfa-recovery-code",
		"mfa-sms", "mfa-voice", "mfa-webauthn", "organizations", "passkeys", "phone-identifier-challenge",
		"phone-identifier-enrollment", "reset-password", "signup", "signup-id", "signup-password", "status",
	}
	availableLanguages = []string{
		"ar", "bg", "bs", "ca-ES", "cs", "cy", "da", "de", "el", "en", "es", "et", "eu-ES", "fi", "fr", "fr-CA", "fr-FR", "gl-ES", "he", "hi", "hr",
		"hu", "id", "is", "it", "ja", "ko", "lt", "lv", "nb", "nl", "nn", "no", "pl", "pt", "pt-BR", "pt-PT", "ro", "ru", "sk",
		"sl", "sr", "sv", "th", "tr", "uk", "vi", "zh-CN", "zh-TW",
	}
)

// NewCustomTextResource will return a new auth0_prompt_custom_text resource.
func NewCustomTextResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPromptCustomText,
		ReadContext:   readPromptCustomText,
		UpdateContext: updatePromptCustomText,
		DeleteContext: deletePromptCustomText,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("prompt", "language"),
		},
		Description: "With this resource, you can manage custom text on your Auth0 prompts. You can read more about " +
			"custom texts [here](https://auth0.com/docs/customize/universal-login-pages/customize-login-text-prompts).",
		Schema: map[string]*schema.Schema{
			"prompt": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(availablePrompts, false),
				Description: "The term `prompt` is used to refer to a specific step in the login flow. " +
					"Options include: `" + strings.Join(availablePrompts, "`, `") + "`.",
			},
			"language": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(availableLanguages, false),
				Description: "Language of the custom text. Options include: `" +
					strings.Join(availableLanguages, "`, `") + "`.",
			},
			"body": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description: "JSON containing the custom texts. You can check the options for each prompt " +
					"[here](https://auth0.com/docs/customize/universal-login-pages/customize-login-text-prompts#prompt-values).",
			},
		},
	}
}

func createPromptCustomText(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	prompt := data.Get("prompt").(string)
	language := data.Get("language").(string)

	internalSchema.SetResourceGroupID(data, prompt, language)

	return updatePromptCustomText(ctx, data, meta)
}

func readPromptCustomText(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	customText, err := api.Prompt.CustomText(ctx, data.Get("prompt").(string), data.Get("language").(string))
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenPromptCustomText(data, customText))
}

func updatePromptCustomText(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	prompt := data.Get("prompt").(string)
	language := data.Get("language").(string)
	body := data.Get("body").(string)

	if body == "" {
		return nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return diag.FromErr(err)
	}

	if err := api.Prompt.SetCustomText(ctx, prompt, language, payload); err != nil {
		return diag.FromErr(err)
	}

	return readPromptCustomText(ctx, data, meta)
}

func deletePromptCustomText(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := data.Set("body", "{}"); err != nil {
		return diag.FromErr(err)
	}

	return updatePromptCustomText(ctx, data, meta)
}
