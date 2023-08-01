package guardian

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalValidation "github.com/auth0/terraform-provider-auth0/internal/validation"
)

// NewResource will return a new auth0_guardian resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createGuardian,
		ReadContext:   readGuardian,
		UpdateContext: updateGuardian,
		DeleteContext: deleteGuardian,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Multi-Factor Authentication works by requiring additional factors during the login process " +
			"to prevent unauthorized access. With this resource you can configure some options available for MFA.",
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(
					[]string{
						"all-applications",
						"confidence-score",
						"never",
					},
					false,
				),
				Description: "Policy to use. Available options are `never`, `all-applications` and `confidence-score`.",
			},
			"webauthn_roaming": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Configuration settings for the WebAuthn with FIDO Security Keys MFA. " +
					"If this block is present, WebAuthn with FIDO Security Keys MFA will be enabled, " +
					"and disabled otherwise.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether WebAuthn with FIDO Security Keys MFA is enabled.",
						},
						"user_verification": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									"discouraged",
									"preferred",
									"required",
								},
								false,
							),
							Description: "User verification, one of `discouraged`, `preferred` or `required`.",
						},
						"override_relying_party": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							Description: "The Relying Party is the domain for which the WebAuthn keys will be issued," +
								" set to `true` if you are customizing the identifier.",
						},
						"relying_party_identifier": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							RequiredWith: []string{"webauthn_roaming.0.override_relying_party"},
							Description:  "The Relying Party should be a suffix of the custom domain.",
						},
					},
				},
			},
			"webauthn_platform": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Configuration settings for the WebAuthn with FIDO Device Biometrics MFA. " +
					"If this block is present, WebAuthn with FIDO Device Biometrics MFA will be enabled, " +
					"and disabled otherwise.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether WebAuthn with FIDO Device Biometrics MFA is enabled.",
						},
						"override_relying_party": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							Description: "The Relying Party is the domain for which the WebAuthn keys will be issued," +
								" set to `true` if you are customizing the identifier.",
						},
						"relying_party_identifier": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							RequiredWith: []string{"webauthn_platform.0.override_relying_party"},
							Description:  "The Relying Party should be a suffix of the custom domain.",
						},
					},
				},
			},
			"phone": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Configuration settings for the phone MFA. If this block is present, " +
					"Phone MFA will be enabled, and disabled otherwise.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether Phone MFA is enabled.",
						},
						"provider": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									"auth0",
									"twilio",
									"phone-message-hook",
								},
								false,
							),
							RequiredWith: []string{"phone.0.message_types"},
							Description: "Provider to use, one of `auth0`, `twilio` or `phone-message-hook`. " +
								"Selecting `phone-message-hook` will require a " +
								"Phone Message Action to be created before. " +
								"[Learn how](https://auth0.com/docs/customize/actions/flows-and-triggers/send-phone-message-flow).",
						},
						"message_types": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							RequiredWith: []string{"phone.0.provider"},
							Description: "Message types to use, array of `sms` and/or `voice`. " +
								"Adding both to the array should enable the user to choose.",
						},
						"options": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "Options for the various providers.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enrollment_message": {
										Type:     schema.TypeString,
										Optional: true,
										Description: "This message will be sent whenever a user enrolls a new device " +
											"for the first time using MFA. Supports Liquid syntax, see " +
											"[Auth0 docs](https://auth0.com/docs/customize/customize-sms-or-voice-messages).",
									},
									"verification_message": {
										Type:     schema.TypeString,
										Optional: true,
										Description: "This message will be sent whenever a user logs in after the " +
											"enrollment. Supports Liquid syntax, see " +
											"[Auth0 docs](https://auth0.com/docs/customize/customize-sms-or-voice-messages).",
									},
									"from": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Phone number to use as the sender.",
									},
									"messaging_service_sid": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Messaging service SID.",
									},
									"auth_token": {
										Type:        schema.TypeString,
										Sensitive:   true,
										Optional:    true,
										Description: "AuthToken for your Twilio account.",
									},
									"sid": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "SID for your Twilio account.",
									},
								},
							},
						},
					},
				},
			},
			"email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicates whether email MFA is enabled.",
			},
			"otp": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicates whether one time password MFA is enabled.",
			},
			"recovery_code": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Indicates whether recovery code MFA is enabled.",
			},
			"duo": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Configuration settings for the Duo MFA. If this block is present, " +
					"Duo MFA will be enabled, and disabled otherwise.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether Duo MFA is enabled.",
						},
						"integration_key": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"duo.0.secret_key", "duo.0.hostname"},
							Description:  "Duo client ID, see the Duo documentation for more details on Duo setup.",
						},
						"secret_key": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							RequiredWith: []string{"duo.0.integration_key", "duo.0.hostname"},
							Description:  "Duo client secret, see the Duo documentation for more details on Duo setup.",
						},
						"hostname": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"duo.0.integration_key", "duo.0.secret_key"},
							Description:  "Duo API Hostname, see the Duo documentation for more details on Duo setup.",
						},
					},
				},
			},
			"push": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Configuration settings for the Push MFA. If this block is present, " +
					"Push MFA will be enabled, and disabled otherwise.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether Push MFA is enabled.",
						},
						"provider": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"direct", "guardian", "sns"}, false),
							Description:  "Provider to use, one of `direct`, `guardian`, `sns`.",
						},
						"amazon_sns": {
							Type:         schema.TypeList,
							Optional:     true,
							Computed:     true,
							MaxItems:     1,
							RequiredWith: []string{"push.0.provider"},
							Description:  "Configuration for Amazon SNS.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_access_key_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Your AWS Access Key ID.",
									},
									"aws_region": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Your AWS application's region.",
									},
									"aws_secret_access_key": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "Your AWS Secret Access Key.",
									},
									"sns_apns_platform_application_arn": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The Amazon Resource Name for your Apple Push Notification Service.",
									},
									"sns_gcm_platform_application_arn": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The Amazon Resource Name for your Firebase Cloud Messaging Service.",
									},
								},
							},
						},
						"custom_app": {
							Type:         schema.TypeList,
							Optional:     true,
							Computed:     true,
							MaxItems:     1,
							RequiredWith: []string{"push.0.provider"},
							Description:  "Configuration for the Guardian Custom App.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"app_name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Custom Application Name.",
									},
									"apple_app_link": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
										Description:  "Apple App Store URL. Must be HTTPS or an empty string.",
									},
									"google_app_link": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
										Description:  "Google Store URL. Must be HTTPS or an empty string.",
									},
								},
							},
						},
						"direct_apns": {
							Type:         schema.TypeList,
							Optional:     true,
							Computed:     true,
							MaxItems:     1,
							RequiredWith: []string{"push.0.provider"},
							Description:  "Configuration for the Apple Push Notification service (APNs) settings.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sandbox": {
										Type:     schema.TypeBool,
										Required: true,
										Description: "Set to true to use the sandbox iOS app environment, " +
											"otherwise set to false to use the production iOS app environment.",
									},
									"bundle_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "The Apple Push Notification service Bundle ID.",
									},
									"p12": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "The base64 encoded certificate in P12 format.",
									},
									"enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Computed:    true,
										Description: "Indicates whether the Apple Push Notification service is enabled.",
									},
								},
							},
						},
						"direct_fcm": {
							Type:         schema.TypeList,
							Optional:     true,
							Computed:     true,
							MaxItems:     1,
							RequiredWith: []string{"push.0.provider"},
							Description:  "Configuration for Firebase Cloud Messaging (FCM) settings.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"server_key": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
										Description: "The Firebase Cloud Messaging Server Key. " +
											"For security purposes, we don’t retrieve your existing FCM server key " +
											"to check for drift.",
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

func createGuardian(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())
	return updateGuardian(ctx, data, meta)
}

func readGuardian(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	flattenedPolicy, err := flattenMultiFactorPolicy(ctx, api)
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(data.Set("policy", flattenedPolicy))

	multiFactorList, err := api.Guardian.MultiFactor.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, factor := range multiFactorList {
		switch factor.GetName() {
		case "email":
			result = multierror.Append(result, data.Set("email", factor.GetEnabled()))
		case "otp":
			result = multierror.Append(result, data.Set("otp", factor.GetEnabled()))
		case "recovery-code":
			result = multierror.Append(result, data.Set("recovery_code", factor.GetEnabled()))
		case "sms":
			phone, err := flattenPhone(ctx, factor.GetEnabled(), api)
			if err != nil {
				return diag.FromErr(err)
			}

			result = multierror.Append(result, data.Set("phone", phone))
		case "webauthn-roaming":
			webAuthnRoaming, err := flattenWebAuthnRoaming(ctx, factor.GetEnabled(), api)
			if err != nil {
				return diag.FromErr(err)
			}

			result = multierror.Append(result, data.Set("webauthn_roaming", webAuthnRoaming))
		case "webauthn-platform":
			webAuthnPlatform, err := flattenWebAuthnPlatform(ctx, factor.GetEnabled(), api)
			if err != nil {
				return diag.FromErr(err)
			}

			result = multierror.Append(result, data.Set("webauthn_platform", webAuthnPlatform))
		case "duo":
			duo, err := flattenDUO(ctx, factor.GetEnabled(), api)
			if err != nil {
				return diag.FromErr(err)
			}

			result = multierror.Append(result, data.Set("duo", duo))
		case "push-notification":
			push, err := flattenPush(ctx, data, factor.GetEnabled(), api)
			if err != nil {
				return diag.FromErr(err)
			}

			result = multierror.Append(result, data.Set("push", push))
		}
	}

	return diag.FromErr(result.ErrorOrNil())
}

func updateGuardian(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	result := multierror.Append(
		updatePolicy(ctx, data, api),
		updateEmailFactor(ctx, data, api),
		updateOTPFactor(ctx, data, api),
		updateRecoveryCodeFactor(ctx, data, api),
		updatePhoneFactor(ctx, data, api),
		updateWebAuthnRoaming(ctx, data, api),
		updateWebAuthnPlatform(ctx, data, api),
		updateDUO(ctx, data, api),
		updatePush(ctx, data, api),
	)
	if err := result.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return readGuardian(ctx, data, meta)
}

func deleteGuardian(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	result := multierror.Append(
		api.Guardian.MultiFactor.UpdatePolicy(ctx, &management.MultiFactorPolicies{}),
		api.Guardian.MultiFactor.Phone.Enable(ctx, false),
		api.Guardian.MultiFactor.Email.Enable(ctx, false),
		api.Guardian.MultiFactor.OTP.Enable(ctx, false),
		api.Guardian.MultiFactor.RecoveryCode.Enable(ctx, false),
		api.Guardian.MultiFactor.WebAuthnRoaming.Enable(ctx, false),
		api.Guardian.MultiFactor.WebAuthnPlatform.Enable(ctx, false),
		api.Guardian.MultiFactor.DUO.Enable(ctx, false),
		api.Guardian.MultiFactor.Push.Enable(ctx, false),
	)

	return diag.FromErr(result.ErrorOrNil())
}
