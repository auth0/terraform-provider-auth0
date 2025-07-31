package prompt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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
		string(management.PromptPasskeys),
		string(management.PromptCaptcha),
		string(management.PromptLogin),
		string(management.PromptSignup),
		string(management.PromptResetPassword),
		string(management.PromptMFA),
		string(management.PromptMFASMS),
		string(management.PromptMFAEmail),
		string(management.PromptMFAPush),
		string(management.PromptInvitation),
		string(management.PromptOrganizations),
		string(management.PromptMFAOTP),
		string(management.PromptDeviceFlow),
		string(management.PromptMFAPhone),
		string(management.PromptMFAVoice),
		string(management.PromptMFARecoveryCode),
		string(management.PromptCommon),
		string(management.PromptEmailVerification),
		string(management.PromptLoginEmailVerification),
		string(management.PromptLogout),
		string(management.PromptMFAWebAuthn),
		string(management.PromptConsent),
		string(management.PromptCustomizedConsent),
		string(management.PromptEmailOTPChallenge),
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
		string(management.ScreenLogin),
		string(management.ScreenSignup),
		string(management.ScreenResetPasswordRequest),
		string(management.ScreenResetPasswordEmail),
		string(management.ScreenResetPassword),
		string(management.ScreenResetPasswordSuccess),
		string(management.ScreenResetPasswordError),
		string(management.ScreenResetPasswordMFAEmailChallenge),
		string(management.ScreenResetPasswordMFAOTPChallenge),
		string(management.ScreenResetPasswordMFAPushChallengePush),
		string(management.ScreenResetPasswordMFASMSChallenge),
		string(management.ScreenMFADetectBrowserCapabilities),
		string(management.ScreenMFAEnrollResult),
		string(management.ScreenMFABeginEnrollOptions),
		string(management.ScreenMFALoginOptions),
		string(management.ScreenMFACountryCodes),
		string(management.ScreenMFASMSChallenge),
		string(management.ScreenMFASMSEnrollment),
		string(management.ScreenMFASMSList),
		string(management.ScreenMFAEmailChallenge),
		string(management.ScreenMFAEmailList),
		string(management.ScreenMFAPushChallengePush),
		string(management.ScreenMFAPushEnrollmentQR),
		string(management.ScreenMFAPushList),
		string(management.ScreenMFAPushWelcome),
		string(management.ScreenAcceptInvitation),
		string(management.ScreenOrganizationSelection),
		string(management.ScreenOrganizationPicker),
		string(management.ScreenMFAOTPChallenge),
		string(management.ScreenMFAOTPEnrollmentCode),
		string(management.ScreenMFAOTPEnrollmentQR),
		string(management.ScreenDeviceCodeActivation),
		string(management.ScreenDeviceCodeActivationAllowed),
		string(management.ScreenDeviceCodeActivationDenied),
		string(management.ScreenDeviceCodeConfirmation),
		string(management.ScreenMFAPhoneChallenge),
		string(management.ScreenMFAPhoneEnrollment),
		string(management.ScreenMFAVoiceChallenge),
		string(management.ScreenMFAVoiceEnrollment),
		string(management.ScreenResetPasswordMFAPhoneChallenge),
		string(management.ScreenResetPasswordMFAVoiceChallenge),
		string(management.ScreenMFARecoveryCodeChallenge),
		string(management.ScreenMFARecoveryCodeEnrollment),
		string(management.ScreenResetPasswordMFARecoveryCodeChallenge),
		string(management.ScreenRedeemTicket),
		"mfa-recovery-code-challenge-new-code",
		string(management.ScreenEmailVerificationResult),
		string(management.ScreenLoginEmailVerification),
		string(management.ScreenLogout),
		string(management.ScreenLogoutAborted),
		string(management.ScreenLogoutComplete),
		string(management.ScreenMFAWebAuthnChangeKeyNickname),
		string(management.ScreenMFAWebAuthnEnrollmentSuccess),
		string(management.ScreenMFAWebAuthnError),
		string(management.ScreenMFAWebAuthnPlatformChallenge),
		string(management.ScreenMFAWebAuthnPlatformEnrollment),
		string(management.ScreenMFAWebAuthnRoamingChallenge),
		string(management.ScreenMFAWebAuthnRoamingEnrollment),
		string(management.ScreenResetPasswordMFAWebAuthnPlatformChallenge),
		string(management.ScreenResetPasswordMFAWebAuthnRoamingChallenge),
		string(management.ScreenConsent),
		string(management.ScreenCustomizedConsent),
		string(management.ScreenEmailOTPChallenge),
		string(management.ScreenMFAWebAuthnNotAvailableError),
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
			"use_page_template": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Use page template with ACUL",
			},
			"filters": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Optional filters to apply rendering rules to specific entities. `match_type` and at least one of the entity arrays are required.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"match_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"includes_any", "excludes_any"}, false),
							Description:  "Type of match to apply. Options: `includes_any`, `excludes_any`.",
						},
						"clients": {
							Type:             schema.TypeString,
							Optional:         true,
							Description:      "An array of clients (applications) identified by id or a metadata key/value pair. Entity Limit: 25.",
							ValidateFunc:     validation.StringIsJSON,
							DiffSuppressFunc: suppressUnorderedJSONDiff,
						},
						"organizations": {
							Type:             schema.TypeString,
							Optional:         true,
							Description:      "An array of organizations identified by id or a metadata key/value pair. Entity Limit: 25.",
							ValidateFunc:     validation.StringIsJSON,
							DiffSuppressFunc: suppressUnorderedJSONDiff,
						},
						"domains": {
							Type:             schema.TypeString,
							Optional:         true,
							Description:      "An array of domains identified by id or a metadata key/value pair. Entity Limit: 25.",
							ValidateFunc:     validation.StringIsJSON,
							DiffSuppressFunc: suppressUnorderedJSONDiff,
						},
					},
				},
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

	if isFiltersNull(data) {
		if err := api.Request(ctx, http.MethodPatch, api.URI("prompts", string(prompt), "screen", string(screen), "rendering"), map[string]interface{}{"filters": nil}); err != nil {
			return diag.FromErr(err)
		}
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

func suppressUnorderedJSONDiff(_, older, newer string, _ *schema.ResourceData) bool {
	var oldArray, newArray []map[string]interface{}

	if err := json.Unmarshal([]byte(older), &oldArray); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(newer), &newArray); err != nil {
		return false
	}

	sort.Slice(oldArray, func(i, j int) bool {
		return fmt.Sprint(oldArray[i]) < fmt.Sprint(oldArray[j])
	})
	sort.Slice(newArray, func(i, j int) bool {
		return fmt.Sprint(newArray[i]) < fmt.Sprint(newArray[j])
	})

	oldBytes, _ := json.Marshal(oldArray)
	newBytes, _ := json.Marshal(newArray)

	return bytes.Equal(oldBytes, newBytes)
}
