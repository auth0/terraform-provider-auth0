package auth0

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
				ValidateFunc: validation.StringInSlice([]string{
					"all-applications",
					"confidence-score",
					"never",
				}, false),
			},
			"phone": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 0,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"provider": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"auth0",
								"twilio",
								"phone-message-hook",
							}, false),
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
							MaxItems: 1,
							MinItems: 1,
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
		},
	}
}

func createGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(resource.UniqueId())
	return updateGuardian(ctx, d, m)
}

func deleteGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.Guardian.MultiFactor.Phone.Enable(false); err != nil {
		return diag.FromErr(err)
	}
	if err := api.Guardian.MultiFactor.Email.Enable(false); err != nil {
		return diag.FromErr(err)
	}
	if err := api.Guardian.MultiFactor.OTP.Enable(false); err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func updateGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	if d.HasChange("policy") {
		policy := d.Get("policy").(string)
		if policy == "never" {
			// Passing empty array to set it to the "never" policy.
			if err := api.Guardian.MultiFactor.UpdatePolicy(&management.MultiFactorPolicies{}); err != nil {
				return diag.FromErr(err)
			}
		} else {
			if err := api.Guardian.MultiFactor.UpdatePolicy(&management.MultiFactorPolicies{policy}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err := updatePhoneFactor(d, api); err != nil {
		return diag.FromErr(err)
	}
	if err := updateEmailFactor(d, api); err != nil {
		return diag.FromErr(err)
	}
	if err := updateOTPFactor(d, api); err != nil {
		return diag.FromErr(err)
	}

	return readGuardian(ctx, d, m)
}

func updatePhoneFactor(d *schema.ResourceData, api *management.Management) error {
	ok, err := factorShouldBeUpdated(d, "phone")
	if err != nil {
		return err
	}
	if ok {
		if err := api.Guardian.MultiFactor.Phone.Enable(true); err != nil {
			return err
		}

		return configurePhone(d, api)
	}

	return api.Guardian.MultiFactor.Phone.Enable(false)
}

func updateEmailFactor(d *schema.ResourceData, api *management.Management) error {
	if changed := d.HasChange("email"); changed {
		enabled := d.Get("email").(bool)
		return api.Guardian.MultiFactor.Email.Enable(enabled)
	}
	return nil
}

func updateOTPFactor(d *schema.ResourceData, api *management.Management) error {
	if changed := d.HasChange("otp"); changed {
		enabled := d.Get("otp").(bool)
		return api.Guardian.MultiFactor.OTP.Enable(enabled)
	}

	return nil
}

func configurePhone(d *schema.ResourceData, api *management.Management) error {
	var err error

	m := make(map[string]interface{})
	List(d, "phone").Elem(func(d ResourceData) {
		m["provider"] = String(d, "provider", HasChange())
		m["message_types"] = Slice(d, "message_types", HasChange())
		m["options"] = List(d, "options")

		switch *String(d, "provider") {
		case "twilio":
			err = updateTwilioOptions(m["options"].(Iterator), api)
		case "auth0":
			err = updateAuth0Options(m["options"].(Iterator), api)
		}
	})
	if err != nil {
		return err
	}

	if provider, ok := m["provider"]; ok {
		if err := api.Guardian.MultiFactor.Phone.UpdateProvider(
			&management.MultiFactorProvider{
				Provider: provider.(*string),
			},
		); err != nil {
			return err
		}
	}

	messageTypes := typeAssertToStringArray(m["message_types"].([]interface{}))
	if messageTypes != nil {
		if err := api.Guardian.MultiFactor.Phone.UpdateMessageTypes(
			&management.PhoneMessageTypes{
				MessageTypes: messageTypes,
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func updateAuth0Options(opts Iterator, api *management.Management) error {
	var err error
	opts.Elem(func(d ResourceData) {
		err = api.Guardian.MultiFactor.SMS.UpdateTemplate(
			&management.MultiFactorSMSTemplate{
				EnrollmentMessage:   String(d, "enrollment_message"),
				VerificationMessage: String(d, "verification_message"),
			},
		)
	})

	return err
}

func updateTwilioOptions(opts Iterator, api *management.Management) error {
	m := make(map[string]*string)

	opts.Elem(func(d ResourceData) {
		m["sid"] = String(d, "sid")
		m["auth_token"] = String(d, "auth_token")
		m["from"] = String(d, "from")
		m["messaging_service_sid"] = String(d, "messaging_service_sid")
		m["enrollment_message"] = String(d, "enrollment_message")
		m["verification_message"] = String(d, "verification_message")
	})

	err := api.Guardian.MultiFactor.SMS.UpdateTwilio(
		&management.MultiFactorProviderTwilio{
			From:                m["from"],
			MessagingServiceSid: m["messaging_service_sid"],
			AuthToken:           m["auth_token"],
			SID:                 m["sid"],
		},
	)
	if err != nil {
		return err
	}

	return api.Guardian.MultiFactor.SMS.UpdateTemplate(
		&management.MultiFactorSMSTemplate{
			EnrollmentMessage:   m["enrollment_message"],
			VerificationMessage: m["verification_message"],
		},
	)
}

func readGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var result *multierror.Error

	api := m.(*management.Management)
	messageTypes, err := api.Guardian.MultiFactor.Phone.MessageTypes()
	if err != nil {
		return diag.FromErr(err)
	}

	phoneData := make(map[string]interface{})
	phoneData["message_types"] = messageTypes.MessageTypes

	phoneProvider, err := api.Guardian.MultiFactor.Phone.Provider()
	if err != nil {
		return diag.FromErr(err)
	}
	phoneData["provider"] = phoneProvider.Provider

	policy, err := api.Guardian.MultiFactor.Policy()
	if err != nil {
		return diag.FromErr(err)
	}

	if len(*policy) == 0 {
		result = multierror.Append(result, d.Set("policy", "never"))
	} else {
		result = multierror.Append(result, d.Set("policy", (*policy)[0]))
	}

	var phoneProviderFlattenedOptions map[string]interface{}
	switch *phoneProvider.Provider {
	case "twilio":
		phoneProviderFlattenedOptions, err = flattenTwilioOptions(api)
	case "auth0":
		phoneProviderFlattenedOptions, err = flattenAuth0Options(api)
	}
	if err != nil {
		return diag.FromErr(err)
	}

	ok, err := factorShouldBeUpdated(d, "phone")
	if err != nil {
		return diag.FromErr(err)
	}
	if ok {
		phoneData["options"] = []interface{}{phoneProviderFlattenedOptions}
		result = multierror.Append(result, d.Set("phone", []interface{}{phoneData}))
	} else {
		result = multierror.Append(result, d.Set("phone", nil))
	}

	factors, err := api.Guardian.MultiFactor.List()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, factor := range factors {
		if factor.Name != nil {
			if *factor.Name == "email" {
				result = multierror.Append(result, d.Set("email", factor.Enabled))
			}
			if *factor.Name == "otp" {
				result = multierror.Append(result, d.Set("otp", factor.Enabled))
			}
		}
	}

	return diag.FromErr(result.ErrorOrNil())
}

func hasBlockPresentInNewState(d *schema.ResourceData, factor string) bool {
	if ok := d.HasChange(factor); ok {
		_, n := d.GetChange(factor)
		newState := n.([]interface{})
		return len(newState) > 0
	}

	return false
}

func flattenAuth0Options(api *management.Management) (map[string]interface{}, error) {
	md := make(map[string]interface{})

	template, err := api.Guardian.MultiFactor.SMS.Template()
	if err != nil {
		return nil, err
	}

	md["enrollment_message"] = template.EnrollmentMessage
	md["verification_message"] = template.VerificationMessage

	return md, nil
}

func flattenTwilioOptions(api *management.Management) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	template, err := api.Guardian.MultiFactor.SMS.Template()
	if err != nil {
		return nil, err
	}

	m["enrollment_message"] = template.EnrollmentMessage
	m["verification_message"] = template.VerificationMessage

	twilio, err := api.Guardian.MultiFactor.SMS.Twilio()
	if err != nil {
		return nil, err
	}

	m["auth_token"] = twilio.AuthToken
	m["from"] = twilio.From
	m["messaging_service_sid"] = twilio.MessagingServiceSid
	m["sid"] = twilio.SID

	return m, nil
}

func typeAssertToStringArray(from []interface{}) *[]string {
	length := len(from)
	if length < 1 {
		return nil
	}
	stringArray := make([]string, length)
	for i, v := range from {
		stringArray[i] = v.(string)
	}
	return &stringArray
}

// Determines if the factor should be updated.
// This depends on if it is in the state,
// if it is about to be added to the state.
func factorShouldBeUpdated(d *schema.ResourceData, factor string) (bool, error) {
	_, ok := d.GetOk(factor)
	return ok || hasBlockPresentInNewState(d, factor), nil
}
