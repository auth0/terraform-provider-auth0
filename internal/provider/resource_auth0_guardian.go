package provider

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func newGuardian() *schema.Resource {
	return &schema.Resource{
		CreateContext: createGuardian,
		ReadContext:   readGuardian,
		UpdateContext: updateGuardian,
		DeleteContext: deleteGuardian,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
			},
			"webauthn_roaming": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						},
						"override_relying_party": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"relying_party_identifier": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							RequiredWith: []string{"webauthn_roaming.0.override_relying_party"},
						},
					},
				},
			},
			"webauthn_platform": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"override_relying_party": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"relying_party_identifier": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							RequiredWith: []string{"webauthn_platform.0.override_relying_party"},
						},
					},
				},
			},
			"phone": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									"auth0",
									"twilio",
									"phone-message-hook",
								},
								false,
							),
						},
						"message_types": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"options": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enrollment_message": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"verification_message": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"from": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"messaging_service_sid": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"auth_token": {
										Type:      schema.TypeString,
										Sensitive: true,
										Optional:  true,
									},
									"sid": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"email": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"otp": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"recovery_code": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"duo": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"integration_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secret_key": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"hostname": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"push": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amazon_sns": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_access_key_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"aws_region": {
										Type:     schema.TypeString,
										Required: true,
									},
									"aws_secret_access_key": {
										Type:      schema.TypeString,
										Required:  true,
										Sensitive: true,
									},
									"sns_apns_platform_application_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
									"sns_gcm_platform_application_arn": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"custom_app": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"app_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"apple_app_link": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsURLWithHTTPS,
									},
									"google_app_link": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsURLWithHTTPS,
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

func createGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(resource.UniqueId())
	return updateGuardian(ctx, d, m)
}

func readGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	flattenedPolicy, err := flattenMultiFactorPolicy(api)
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(d.Set("policy", flattenedPolicy))

	multiFactorList, err := api.Guardian.MultiFactor.List()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, factor := range multiFactorList {
		switch factor.GetName() {
		case "email":
			result = multierror.Append(result, d.Set("email", factor.GetEnabled()))
		case "otp":
			result = multierror.Append(result, d.Set("otp", factor.GetEnabled()))
		case "recovery-code":
			result = multierror.Append(result, d.Set("recovery_code", factor.GetEnabled()))
		case "sms":
			result = multierror.Append(result, d.Set("phone", nil))

			if factor.GetEnabled() {
				phone, err := flattenPhone(api)
				if err != nil {
					return diag.FromErr(err)
				}

				result = multierror.Append(result, d.Set("phone", phone))
			}
		case "webauthn-roaming":
			result = multierror.Append(result, d.Set("webauthn_roaming", nil))

			if factor.GetEnabled() {
				webAuthnRoaming, err := flattenWebAuthnRoaming(api)
				if err != nil {
					return diag.FromErr(err)
				}

				result = multierror.Append(result, d.Set("webauthn_roaming", webAuthnRoaming))
			}
		case "webauthn-platform":
			result = multierror.Append(result, d.Set("webauthn_platform", nil))

			if factor.GetEnabled() {
				webAuthnPlatform, err := flattenWebAuthnPlatform(api)
				if err != nil {
					return diag.FromErr(err)
				}

				result = multierror.Append(result, d.Set("webauthn_platform", webAuthnPlatform))
			}
		case "duo":
			result = multierror.Append(result, d.Set("duo", nil))

			if factor.GetEnabled() {
				duo, err := flattenDUO(api)
				if err != nil {
					return diag.FromErr(err)
				}

				result = multierror.Append(result, d.Set("duo", duo))
			}
		case "push":
			result = multierror.Append(result, d.Set("push", nil))

			if factor.GetEnabled() {
				push, err := flattenPush(api)
				if err != nil {
					return diag.FromErr(err)
				}

				result = multierror.Append(result, d.Set("push", push))
			}
		}
	}

	return diag.FromErr(result.ErrorOrNil())
}

func updateGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	result := multierror.Append(
		updatePolicy(d, api),
		updateEmailFactor(d, api),
		updateOTPFactor(d, api),
		updateRecoveryCodeFactor(d, api),
		updatePhoneFactor(d, api),
		updateWebAuthnRoaming(d, api),
		updateWebAuthnPlatform(d, api),
		updateDUO(d, api),
		updatePush(d, api),
	)
	if err := result.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	return readGuardian(ctx, d, m)
}

func deleteGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	result := multierror.Append(
		api.Guardian.MultiFactor.UpdatePolicy(&management.MultiFactorPolicies{}),
		api.Guardian.MultiFactor.Phone.Enable(false),
		api.Guardian.MultiFactor.Email.Enable(false),
		api.Guardian.MultiFactor.OTP.Enable(false),
		api.Guardian.MultiFactor.RecoveryCode.Enable(false),
		api.Guardian.MultiFactor.WebAuthnRoaming.Enable(false),
		api.Guardian.MultiFactor.WebAuthnPlatform.Enable(false),
		api.Guardian.MultiFactor.DUO.Enable(false),
		api.Guardian.MultiFactor.Push.Enable(false),
	)
	if err := result.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return nil
}
