package provider

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
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

		return configurePhone(d.GetRawConfig(), api)
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

func configurePhone(config cty.Value, api *management.Management) error {
	var err error

	config.GetAttr("phone").ForEachElement(func(_ cty.Value, phone cty.Value) (stop bool) {
		mfaProvider := &management.MultiFactorProvider{
			Provider: value.String(phone.GetAttr("provider")),
		}
		if err = api.Guardian.MultiFactor.Phone.UpdateProvider(mfaProvider); err != nil {
			return stop
		}

		options := phone.GetAttr("options")
		switch mfaProvider.GetProvider() {
		case "twilio":
			if err = updateTwilioOptions(options, api); err != nil {
				return true
			}
		case "auth0":
			if err = updateAuth0Options(options, api); err != nil {
				return true
			}
		}

		messageTypes := &management.PhoneMessageTypes{
			MessageTypes: value.Strings(phone.GetAttr("message_types")),
		}
		if err = api.Guardian.MultiFactor.Phone.UpdateMessageTypes(messageTypes); err != nil {
			return stop
		}

		return stop
	})

	return err
}

func updateAuth0Options(options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.SMS.UpdateTemplate(
			&management.MultiFactorSMSTemplate{
				EnrollmentMessage:   value.String(config.GetAttr("enrollment_message")),
				VerificationMessage: value.String(config.GetAttr("verification_message")),
			},
		)

		return stop
	})

	return err
}

func updateTwilioOptions(options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		if err = api.Guardian.MultiFactor.SMS.UpdateTwilio(
			&management.MultiFactorProviderTwilio{
				From:                value.String(config.GetAttr("from")),
				MessagingServiceSid: value.String(config.GetAttr("messaging_service_sid")),
				AuthToken:           value.String(config.GetAttr("auth_token")),
				SID:                 value.String(config.GetAttr("sid")),
			},
		); err != nil {
			return stop
		}

		if err = api.Guardian.MultiFactor.SMS.UpdateTemplate(
			&management.MultiFactorSMSTemplate{
				EnrollmentMessage:   value.String(config.GetAttr("enrollment_message")),
				VerificationMessage: value.String(config.GetAttr("verification_message")),
			},
		); err != nil {
			return stop
		}

		return stop
	})

	return err
}

func updateWebAuthnRoaming(d *schema.ResourceData, api *management.Management) error {
	if factorShouldBeUpdated(d, "webauthn_roaming") {
		if err := api.Guardian.MultiFactor.WebAuthnRoaming.Enable(true); err != nil {
			return err
		}

		var webAuthnSettings management.MultiFactorWebAuthnSettings

		d.GetRawConfig().GetAttr("webauthn_roaming").ForEachElement(
			func(_ cty.Value, config cty.Value) (stop bool) {
				webAuthnSettings.OverrideRelyingParty = value.Bool(config.GetAttr("override_relying_party"))
				webAuthnSettings.RelyingPartyIdentifier = value.String(config.GetAttr("relying_party_identifier"))
				webAuthnSettings.UserVerification = value.String(config.GetAttr("user_verification"))
				return stop
			},
		)

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

		d.GetRawConfig().GetAttr("webauthn_platform").ForEachElement(
			func(_ cty.Value, config cty.Value) (stop bool) {
				webAuthnSettings.OverrideRelyingParty = value.Bool(config.GetAttr("override_relying_party"))
				webAuthnSettings.RelyingPartyIdentifier = value.String(config.GetAttr("relying_party_identifier"))
				return stop
			},
		)

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

		d.GetRawConfig().GetAttr("duo").ForEachElement(
			func(_ cty.Value, config cty.Value) (stop bool) {
				duoSettings.SecretKey = value.String(config.GetAttr("secret_key"))
				duoSettings.Hostname = value.String(config.GetAttr("hostname"))
				duoSettings.IntegrationKey = value.String(config.GetAttr("integration_key"))
				return stop
			},
		)

		if duoSettings == (management.MultiFactorDUOSettings{}) {
			return nil
		}

		return api.Guardian.MultiFactor.DUO.Update(&duoSettings)
	}

	return api.Guardian.MultiFactor.DUO.Enable(false)
}

func updatePush(d *schema.ResourceData, api *management.Management) error {
	if factorShouldBeUpdated(d, "push") {
		if err := api.Guardian.MultiFactor.Push.Enable(true); err != nil {
			return err
		}

		var err error
		d.GetRawConfig().GetAttr("push").ForEachElement(func(_ cty.Value, push cty.Value) (stop bool) {
			if d.HasChange("push.0.amazon_sns") {
				var amazonSNS *management.MultiFactorProviderAmazonSNS
				push.GetAttr("amazon_sns").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
					amazonSNS = &management.MultiFactorProviderAmazonSNS{
						AccessKeyID:                value.String(config.GetAttr("aws_access_key_id")),
						SecretAccessKeyID:          value.String(config.GetAttr("aws_secret_access_key")),
						Region:                     value.String(config.GetAttr("aws_region")),
						APNSPlatformApplicationARN: value.String(config.GetAttr("sns_apns_platform_application_arn")),
						GCMPlatformApplicationARN:  value.String(config.GetAttr("sns_gcm_platform_application_arn")),
					}
					return stop
				})
				if amazonSNS != nil {
					if err = api.Guardian.MultiFactor.Push.UpdateAmazonSNS(amazonSNS); err != nil {
						return stop
					}
				}
			}

			if d.HasChange("push.0.custom_app") {
				var customApp *management.MultiFactorPushCustomApp
				push.GetAttr("custom_app").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
					customApp = &management.MultiFactorPushCustomApp{
						AppName:       value.String(config.GetAttr("app_name")),
						AppleAppLink:  value.String(config.GetAttr("apple_app_link")),
						GoogleAppLink: value.String(config.GetAttr("google_app_link")),
					}
					return stop
				})
				if customApp != nil {
					if err = api.Guardian.MultiFactor.Push.UpdateCustomApp(customApp); err != nil {
						return stop
					}
				}
			}

			return stop
		})

		return err
	}

	return api.Guardian.MultiFactor.Push.Enable(false)
}
