package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	availablePrompts = []string{
		"login", "login-id", "login-password", "login-email-verification", "signup", "signup-id", "signup-password",
		"reset-password", "consent", "mfa-push", "mfa-otp", "mfa-voice", "mfa-phone", "mfa-webauthn", "mfa-sms",
		"mfa-email", "mfa-recovery-code", "mfa", "status", "device-flow", "email-verification", "email-otp-challenge",
		"organizations", "invitation", "common",
	}
	availableLanguages = []string{
		"ar", "bg", "bs", "cs", "da", "de", "el", "en", "es", "et", "fi", "fr", "fr-CA", "fr-FR", "he", "hi", "hr",
		"hu", "id", "is", "it", "ja", "ko", "lt", "lv", "nb", "nl", "pl", "pt", "pt-BR", "pt-PT", "ro", "ru", "sk",
		"sl", "sr", "sv", "th", "tr", "uk", "vi", "zh-CN", "zh-TW",
	}
	errEmptyPromptCustomTextID         = fmt.Errorf("ID cannot be empty")
	errInvalidPromptCustomTextIDFormat = fmt.Errorf("ID must be formated as prompt:language")
)

func newPromptCustomText() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPromptCustomText,
		ReadContext:   readPromptCustomText,
		UpdateContext: updatePromptCustomText,
		DeleteContext: deletePromptCustomText,
		Importer: &schema.ResourceImporter{
			StateContext: importPromptCustomText,
		},
		Description: "With this resource, you can manage custom text on your Auth0 prompts. You can read more about " +
			"custom texts [here](https://auth0.com/docs/customize/universal-login-pages/customize-login-text-prompts).",
		Schema: map[string]*schema.Schema{
			"prompt": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(availablePrompts, false),
				Description: "The term `prompt` is used to refer to a specific step in the login flow. " +
					"Options include `login`, `login-id`, `login-password`, `login-email-verification`, `signup`, " +
					"`signup-id`, `signup-password`, `reset-password`, `consent`, `mfa-push`, `mfa-otp`, `mfa-voice`," +
					" `mfa-phone`, `mfa-webauthn`, `mfa-sms`, `mfa-email`, `mfa-recovery-code`, `mfa`, `status`, " +
					"`device-flow`, `email-verification`, `email-otp-challenge`, `organizations`, " +
					"`invitation`, `common`.",
			},
			"language": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(availableLanguages, false),
				Description: "Language of the custom text. Options include `ar`, `bg`, `bs`, `cs`, `da`, `de`, `el`, " +
					"`en`, `es`, `et`, `fi`, `fr`, `fr-CA`, `fr-FR`, `he`, `hi`, `hr`, `hu`, `id`, `is`, `it`, `ja`, " +
					"`ko`, `lt`, `lv`, `nb`, `nl`, `pl`, `pt`, `pt-BR`, `pt-PT`, `ro`, `ru`, `sk`, `sl`, `sr`, `sv`, " +
					"`th`, `tr`, `uk`, `vi`, `zh-CN`, `zh-TW`",
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

func importPromptCustomText(ctx context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	prompt, language, err := getPromptAndLanguage(d)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	d.SetId(d.Id())

	result := multierror.Append(
		d.Set("prompt", prompt),
		d.Set("language", language),
	)

	return []*schema.ResourceData{d}, result.ErrorOrNil()
}

func createPromptCustomText(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(d.Get("prompt").(string) + ":" + d.Get("language").(string))
	return updatePromptCustomText(ctx, d, m)
}

func readPromptCustomText(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
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
	api := m.(*management.Management)
	prompt, language, err := getPromptAndLanguage(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("body").(string) != "" {
		var body map[string]interface{}
		if err := json.Unmarshal([]byte(d.Get("body").(string)), &body); err != nil {
			return diag.FromErr(err)
		}

		if err := api.Prompt.SetCustomText(prompt, language, body); err != nil {
			return diag.FromErr(err)
		}
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

func getPromptAndLanguage(d *schema.ResourceData) (string, string, error) {
	rawID := d.Id()
	if rawID == "" {
		return "", "", errEmptyPromptCustomTextID
	}

	if !strings.Contains(rawID, ":") {
		return "", "", errInvalidPromptCustomTextIDFormat
	}

	idPair := strings.Split(rawID, ":")
	if len(idPair) != 2 {
		return "", "", errInvalidPromptCustomTextIDFormat
	}

	return idPair[0], idPair[1], nil
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
