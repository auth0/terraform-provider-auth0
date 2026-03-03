package branding

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewPhoneNotificationTemplateResource returns a new auth0_branding_phone_notification_template resource.
func NewPhoneNotificationTemplateResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPhoneNotificationTemplate,
		ReadContext:   readPhoneNotificationTemplate,
		UpdateContext: updatePhoneNotificationTemplate,
		DeleteContext: deletePhoneNotificationTemplate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages phone notification templates used for SMS and voice communications in Auth0.",
		Schema: map[string]*schema.Schema{
			"template_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the phone notification template.",
			},
			"channel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The channel of the phone notification template (e.g., `sms`, `voice`).",
			},
			"tenant": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The tenant of the phone notification template.",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"otp_verify",
					"otp_enroll",
					"change_password",
					"blocked_account",
					"password_breach"},
					false),
				Description: "The type of the phone notification template.",
			},
			"customizable": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether the phone notification template is customizable.",
			},
			"disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether the phone notification template is disabled.",
			},
			"content": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "The content of the phone notification template.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"syntax": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The syntax of the phone notification template.",
						},
						"from": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "The sender phone number for SMS or voice notifications.",
						},
						"body": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Description: "The body content of the phone notification template.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"text": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The text content for SMS notifications.",
									},
									"voice": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The voice content for voice notifications.",
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

func createPhoneNotificationTemplate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	templateConfig := expandPhoneNotificationTemplate(data.GetRawConfig())

	if err := api.Branding.CreatePhoneNotificationTemplate(ctx, templateConfig); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(templateConfig.GetID())

	return readPhoneNotificationTemplate(ctx, data, meta)
}

func readPhoneNotificationTemplate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	template, err := api.Branding.ReadPhoneNotificationTemplate(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenPhoneNotificationTemplate(data, template))
}

func updatePhoneNotificationTemplate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	templateConfig := expandPhoneNotificationTemplate(data.GetRawConfig())

	if err := api.Branding.UpdatePhoneNotificationTemplate(ctx, data.Id(), templateConfig); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readPhoneNotificationTemplate(ctx, data, meta)
}

func deletePhoneNotificationTemplate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Branding.DeletePhoneNotificationTemplate(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
