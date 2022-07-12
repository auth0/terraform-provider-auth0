package auth0

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenMultiFactorPolicy(api *management.Management) (string, error) {
	multiFactorPolicies, err := api.Guardian.MultiFactor.Policy()
	if err != nil {
		return "", err
	}

	flattenedPolicy := "never"
	if len(*multiFactorPolicies) > 0 {
		flattenedPolicy = (*multiFactorPolicies)[0]
	}

	return flattenedPolicy, nil
}

func flattenPhone(api *management.Management) ([]interface{}, error) {
	phoneMessageTypes, err := api.Guardian.MultiFactor.Phone.MessageTypes()
	if err != nil {
		return nil, err
	}

	phoneData := make(map[string]interface{})
	phoneData["message_types"] = phoneMessageTypes.GetMessageTypes()

	phoneProvider, err := api.Guardian.MultiFactor.Phone.Provider()
	if err != nil {
		return nil, err
	}
	phoneData["provider"] = phoneProvider.GetProvider()

	var phoneProviderOptions []interface{}
	switch phoneProvider.GetProvider() {
	case "twilio":
		phoneProviderOptions, err = flattenTwilioOptions(api)
		if err != nil {
			return nil, err
		}
	case "auth0":
		phoneProviderOptions, err = flattenAuth0Options(api)
		if err != nil {
			return nil, err
		}
	case "phone-message-hook":
		phoneProviderOptions = []interface{}{nil}
	}

	phoneData["options"] = phoneProviderOptions

	return []interface{}{phoneData}, nil
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

func flattenWebAuthnRoaming(api *management.Management) ([]interface{}, error) {
	webAuthnSettings, err := api.Guardian.MultiFactor.WebAuthnRoaming.Read()
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{
		"user_verification":        webAuthnSettings.GetUserVerification(),
		"override_relying_party":   webAuthnSettings.GetOverrideRelyingParty(),
		"relying_party_identifier": webAuthnSettings.GetRelyingPartyIdentifier(),
	}

	return []interface{}{m}, nil
}

func flattenWebAuthnPlatform(api *management.Management) ([]interface{}, error) {
	webAuthnSettings, err := api.Guardian.MultiFactor.WebAuthnPlatform.Read()
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{
		"override_relying_party":   webAuthnSettings.GetOverrideRelyingParty(),
		"relying_party_identifier": webAuthnSettings.GetRelyingPartyIdentifier(),
	}

	return []interface{}{m}, nil
}

func flattenDUO(api *management.Management) ([]interface{}, error) {
	duoSettings, err := api.Guardian.MultiFactor.DUO.Read()
	if err != nil {
		return nil, err
	}

	m := map[string]interface{}{
		"integration_key": duoSettings.GetIntegrationKey(),
		"secret_key":      duoSettings.GetSecretKey(),
		"hostname":        duoSettings.GetHostname(),
	}

	return []interface{}{m}, nil
}

func flattenPush(api *management.Management) ([]interface{}, error) {
	amazonSNS, err := api.Guardian.MultiFactor.Push.AmazonSNS()
	if err != nil {
		return nil, err
	}

	pushData := make(map[string]interface{})
	pushData["amazon_sns"] = []interface{}{
		map[string]interface{}{
			"aws_access_key_id":                 amazonSNS.GetAccessKeyID(),
			"aws_region":                        amazonSNS.GetRegion(),
			"aws_secret_access_key":             amazonSNS.GetSecretAccessKeyID(),
			"sns_apns_platform_application_arn": amazonSNS.GetAPNSPlatformApplicationARN(),
			"sns_gcm_platform_application_arn":  amazonSNS.GetGCMPlatformApplicationARN(),
		},
	}

	customApp, err := api.Guardian.MultiFactor.Push.CustomApp()
	if err != nil {
		return nil, err
	}

	pushData["custom_app"] = []interface{}{
		map[string]interface{}{
			"app_name":        customApp.GetAppName(),
			"apple_app_link":  customApp.GetAppleAppLink(),
			"google_app_link": customApp.GetGoogleAppLink(),
		},
	}

	return []interface{}{pushData}, nil
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

func updateRecoveryCodeFactor(d *schema.ResourceData, api *management.Management) error {
	if d.HasChange("recovery_code") {
		enabled := d.Get("recovery_code").(bool)
		return api.Guardian.MultiFactor.RecoveryCode.Enable(enabled)
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

func updateWebAuthnRoaming(d *schema.ResourceData, api *management.Management) error {
	if factorShouldBeUpdated(d, "webauthn_roaming") {
		if err := api.Guardian.MultiFactor.WebAuthnRoaming.Enable(true); err != nil {
			return err
		}

		var webAuthnSettings management.MultiFactorWebAuthnSettings

		List(d, "webauthn_roaming").Elem(func(d ResourceData) {
			webAuthnSettings.OverrideRelyingParty = Bool(d, "override_relying_party")
			webAuthnSettings.RelyingPartyIdentifier = String(d, "relying_party_identifier")
			webAuthnSettings.UserVerification = String(d, "user_verification")
		})

		if webAuthnSettings == (management.MultiFactorWebAuthnSettings{}) {
			return nil
		}

		return api.Guardian.MultiFactor.WebAuthnRoaming.Update(&webAuthnSettings)
	}

	return api.Guardian.MultiFactor.WebAuthnRoaming.Enable(false)
}

func updateWebAuthnPlatform(d *schema.ResourceData, api *management.Management) error {
	if factorShouldBeUpdated(d, "webauthn_platform") {
		if err := api.Guardian.MultiFactor.WebAuthnPlatform.Enable(true); err != nil {
			return err
		}

		var webAuthnSettings management.MultiFactorWebAuthnSettings

		List(d, "webauthn_platform").Elem(func(d ResourceData) {
			webAuthnSettings.OverrideRelyingParty = Bool(d, "override_relying_party")
			webAuthnSettings.RelyingPartyIdentifier = String(d, "relying_party_identifier")
		})

		if webAuthnSettings == (management.MultiFactorWebAuthnSettings{}) {
			return nil
		}

		return api.Guardian.MultiFactor.WebAuthnPlatform.Update(&webAuthnSettings)
	}

	return api.Guardian.MultiFactor.WebAuthnPlatform.Enable(false)
}

func updateDUO(d *schema.ResourceData, api *management.Management) error {
	if factorShouldBeUpdated(d, "duo") {
		if err := api.Guardian.MultiFactor.DUO.Enable(true); err != nil {
			return err
		}

		var duoSettings management.MultiFactorDUOSettings

		List(d, "duo").Elem(func(d ResourceData) {
			duoSettings.SecretKey = String(d, "secret_key")
			duoSettings.Hostname = String(d, "hostname")
			duoSettings.IntegrationKey = String(d, "integration_key")
		})

		return api.Guardian.MultiFactor.DUO.Update(&duoSettings)
	}

	return api.Guardian.MultiFactor.DUO.Enable(false)
}

func updatePush(d *schema.ResourceData, api *management.Management) error {
	if factorShouldBeUpdated(d, "push") {
		if err := api.Guardian.MultiFactor.Push.Enable(true); err != nil {
			return err
		}

		var amazonSNS *management.MultiFactorProviderAmazonSNS
		List(d, "amazon_sns", HasChange()).Elem(func(d ResourceData) {
			amazonSNS = &management.MultiFactorProviderAmazonSNS{
				AccessKeyID:                String(d, "aws_access_key_id"),
				SecretAccessKeyID:          String(d, "aws_secret_access_key"),
				Region:                     String(d, "aws_region"),
				APNSPlatformApplicationARN: String(d, "sns_apns_platform_application_arn"),
				GCMPlatformApplicationARN:  String(d, "sns_gcm_platform_application_arn"),
			}
		})
		if amazonSNS != nil {
			if err := api.Guardian.MultiFactor.Push.UpdateAmazonSNS(amazonSNS); err != nil {
				return err
			}
		}

		var customApp *management.MultiFactorPushCustomApp
		List(d, "custom_app", HasChange()).Elem(func(d ResourceData) {
			customApp = &management.MultiFactorPushCustomApp{
				AppName:       String(d, "app_name"),
				AppleAppLink:  String(d, "apple_app_link"),
				GoogleAppLink: String(d, "google_app_link"),
			}
		})
		if customApp != nil {
			if err := api.Guardian.MultiFactor.Push.UpdateCustomApp(customApp); err != nil {
				return err
			}
		}

		return nil
	}

	return api.Guardian.MultiFactor.Push.Enable(false)
}
