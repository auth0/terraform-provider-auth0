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
							Computed: true,
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

func readGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	multiFactorPolicies, err := api.Guardian.MultiFactor.Policy()
	if err != nil {
		return diag.FromErr(err)
	}

	result := &multierror.Error{}
	if len(*multiFactorPolicies) == 0 {
		result = multierror.Append(result, d.Set("policy", "never"))
	} else {
		result = multierror.Append(result, d.Set("policy", (*multiFactorPolicies)[0]))
	}

	multiFactorList, err := api.Guardian.MultiFactor.List()
	if err != nil {
		return diag.FromErr(err)
	}

	var phoneEnabled bool
	for _, factor := range multiFactorList {
		switch factor.GetName() {
		case "email":
			result = multierror.Append(result, d.Set("email", factor.GetEnabled()))
		case "otp":
			result = multierror.Append(result, d.Set("otp", factor.GetEnabled()))
		case "sms":
			phoneEnabled = factor.GetEnabled()
		}
	}

	if !phoneEnabled {
		result = multierror.Append(result, d.Set("phone", nil))
		return diag.FromErr(result.ErrorOrNil())
	}

	phoneMessageTypes, err := api.Guardian.MultiFactor.Phone.MessageTypes()
	if err != nil {
		return diag.FromErr(err)
	}

	phoneData := make(map[string]interface{})
	phoneData["message_types"] = phoneMessageTypes.GetMessageTypes()

	phoneProvider, err := api.Guardian.MultiFactor.Phone.Provider()
	if err != nil {
		return diag.FromErr(err)
	}
	phoneData["provider"] = phoneProvider.GetProvider()

	var phoneProviderOptions []interface{}
	switch phoneProvider.GetProvider() {
	case "twilio":
		phoneProviderOptions, err = flattenTwilioOptions(api)
		if err != nil {
			return diag.FromErr(err)
		}
	case "auth0":
		phoneProviderOptions, err = flattenAuth0Options(api)
		if err != nil {
			return diag.FromErr(err)
		}
	case "phone-message-hook":
		phoneProviderOptions = []interface{}{nil}
	}

	phoneData["options"] = phoneProviderOptions
	result = multierror.Append(result, d.Set("phone", []interface{}{phoneData}))

	return diag.FromErr(result.ErrorOrNil())
}

func updateGuardian(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	if err := updatePolicy(d, api); err != nil {
		return diag.FromErr(err)
	}

	if err := updateEmailFactor(d, api); err != nil {
		return diag.FromErr(err)
	}
	if err := updateOTPFactor(d, api); err != nil {
		return diag.FromErr(err)
	}

	if err := updatePhoneFactor(d, api); err != nil {
		return diag.FromErr(err)
	}

	return readGuardian(ctx, d, m)
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

func flattenAuth0Options(api *management.Management) ([]interface{}, error) {
	m := make(map[string]interface{})

	template, err := api.Guardian.MultiFactor.SMS.Template()
	if err != nil {
		return nil, err
	}

	m["enrollment_message"] = template.GetEnrollmentMessage()
	m["verification_message"] = template.GetVerificationMessage()

	return []interface{}{m}, nil
}

func flattenTwilioOptions(api *management.Management) ([]interface{}, error) {
	m := make(map[string]interface{})

	template, err := api.Guardian.MultiFactor.SMS.Template()
	if err != nil {
		return nil, err
	}

	m["enrollment_message"] = template.GetEnrollmentMessage()
	m["verification_message"] = template.GetVerificationMessage()

	twilio, err := api.Guardian.MultiFactor.SMS.Twilio()
	if err != nil {
		return nil, err
	}

	m["auth_token"] = twilio.GetAuthToken()
	m["from"] = twilio.GetFrom()
	m["messaging_service_sid"] = twilio.GetMessagingServiceSid()
	m["sid"] = twilio.GetSID()

	return []interface{}{m}, nil
}

func updatePolicy(d *schema.ResourceData, api *management.Management) error {
	if d.HasChange("policy") {
		multiFactorPolicies := management.MultiFactorPolicies{}

		policy := d.Get("policy").(string)
		if policy != "never" {
			multiFactorPolicies = append(multiFactorPolicies, policy)
		}

		// If the policy is "never" then the slice needs to be empty.
		return api.Guardian.MultiFactor.UpdatePolicy(&multiFactorPolicies)
	}
	return nil
}

func updateEmailFactor(d *schema.ResourceData, api *management.Management) error {
	if d.HasChange("email") {
		enabled := d.Get("email").(bool)
		return api.Guardian.MultiFactor.Email.Enable(enabled)
	}
	return nil
}

func updateOTPFactor(d *schema.ResourceData, api *management.Management) error {
	if d.HasChange("otp") {
		enabled := d.Get("otp").(bool)
		return api.Guardian.MultiFactor.OTP.Enable(enabled)
	}
	return nil
}

func updatePhoneFactor(d *schema.ResourceData, api *management.Management) error {
	if factorShouldBeUpdated(d, "phone") {
		// Always enable phone factor before configuring it.
		// Otherwise, we encounter an error with message_types.
		if err := api.Guardian.MultiFactor.Phone.Enable(true); err != nil {
			return err
		}

		return configurePhone(d, api)
	}

	return api.Guardian.MultiFactor.Phone.Enable(false)
}

// Determines if the factor should be updated.
// This depends on if it is in the state,
// if it is about to be added to the state.
func factorShouldBeUpdated(d *schema.ResourceData, factor string) bool {
	_, ok := d.GetOk(factor)
	return ok || hasBlockPresentInNewState(d, factor)
}

func hasBlockPresentInNewState(d *schema.ResourceData, factor string) bool {
	if d.HasChange(factor) {
		_, n := d.GetChange(factor)
		newState := n.([]interface{})
		return len(newState) > 0
	}

	return false
}

func configurePhone(d *schema.ResourceData, api *management.Management) error {
	m := make(map[string]interface{})
	List(d, "phone").Elem(func(d ResourceData) {
		m["provider"] = String(d, "provider")
		m["message_types"] = Slice(d, "message_types")
		m["options"] = List(d, "options")
	})

	if p, ok := m["provider"]; ok && p != nil {
		provider := p.(*string)
		switch *provider {
		case "twilio":
			if err := updateTwilioOptions(m["options"].(Iterator), api); err != nil {
				return err
			}
		case "auth0":
			if err := updateAuth0Options(m["options"].(Iterator), api); err != nil {
				return err
			}
		}

		multiFactorProvider := &management.MultiFactorProvider{Provider: provider}
		if err := api.Guardian.MultiFactor.Phone.UpdateProvider(multiFactorProvider); err != nil {
			return err
		}
	}

	messageTypes := fromInterfaceSliceToStringSlice(m["message_types"].([]interface{}))
	if len(messageTypes) == 0 {
		return nil
	}

	return api.Guardian.MultiFactor.Phone.UpdateMessageTypes(
		&management.PhoneMessageTypes{MessageTypes: &messageTypes},
	)
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

func fromInterfaceSliceToStringSlice(from []interface{}) []string {
	length := len(from)
	if length == 0 {
		return nil
	}

	stringArray := make([]string, length)
	for i, v := range from {
		stringArray[i] = v.(string)
	}

	return stringArray
}
