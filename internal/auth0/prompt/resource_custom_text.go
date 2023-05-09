package prompt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

var (
	availablePrompts = []string{
		"common", "consent", "device-flow", "email-otp-challenge", "email-verification", "invitation", "login",
		"login-id", "login-password", "login-passwordless", "login-email-verification", "logout", "mfa", "mfa-email",
		"mfa-otp", "mfa-phone", "mfa-push", "mfa-recovery-code", "mfa-sms", "mfa-voice", "mfa-webauthn",
		"organizations", "reset-password", "signup", "signup-id", "signup-password", "status",
	}
	availableLanguages = []string{
		"ar", "bg", "bs", "cs", "da", "de", "el", "en", "es", "et", "fi", "fr", "fr-CA", "fr-FR", "he", "hi", "hr",
		"hu", "id", "is", "it", "ja", "ko", "lt", "lv", "nb", "nl", "pl", "pt", "pt-BR", "pt-PT", "ro", "ru", "sk",
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
			StateContext: internalSchema.ImportResourcePairID("prompt", "language"),
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

func createPromptCustomText(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(d.Get("prompt").(string) + ":" + d.Get("language").(string))
	return updatePromptCustomText(ctx, d, m)
}

func readPromptCustomText(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
	customText, err := api.Prompt.CustomText(d.Get("prompt").(string), d.Get("language").(string))
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	body, err := marshalCustomTextBody(customText)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(d.Set("body", body))
}

func updatePromptCustomText(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	prompt := d.Get("prompt").(string)
	language := d.Get("language").(string)
	body := d.Get("body").(string)

	if body == "" {
		return nil
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return diag.FromErr(err)
	}

	if err := api.Prompt.SetCustomText(prompt, language, payload); err != nil {
		return diag.FromErr(err)
	}

	return readPromptCustomText(ctx, d, m)
}

func deletePromptCustomText(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := d.Set("body", "{}"); err != nil {
		return diag.FromErr(err)
	}
	if err := updatePromptCustomText(ctx, d, m); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func marshalCustomTextBody(b map[string]interface{}) (string, error) {
	bodyBytes, err := json.Marshal(b)
	if err != nil {
		return "", fmt.Errorf("failed to serialize the custom texts to JSON: %w", err)
	}

	var buffer bytes.Buffer
	const jsonIndentation = "    "
	if err := json.Indent(&buffer, bodyBytes, "", jsonIndentation); err != nil {
		return "", fmt.Errorf("failed to format the custom texts JSON: %w", err)
	}

	return buffer.String(), nil
}
