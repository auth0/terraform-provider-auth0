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

func flattenPhone(enabled bool, api *management.Management) ([]interface{}, error) {
	phoneData := make(map[string]interface{})
	phoneData["enabled"] = enabled

	if !enabled {
		return []interface{}{phoneData}, nil
	}

	phoneMessageTypes, err := api.Guardian.MultiFactor.Phone.MessageTypes()
	if err != nil {
		return nil, err
	}
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

func flattenWebAuthnRoaming(enabled bool, api *management.Management) ([]interface{}, error) {
	webAuthnRoamingData := make(map[string]interface{})
	webAuthnRoamingData["enabled"] = enabled

	if !enabled {
		return []interface{}{webAuthnRoamingData}, nil
	}

	webAuthnSettings, err := api.Guardian.MultiFactor.WebAuthnRoaming.Read()
	if err != nil {
		return nil, err
	}

	webAuthnRoamingData["user_verification"] = webAuthnSettings.GetUserVerification()
	webAuthnRoamingData["override_relying_party"] = webAuthnSettings.GetOverrideRelyingParty()
	webAuthnRoamingData["relying_party_identifier"] = webAuthnSettings.GetRelyingPartyIdentifier()

	return []interface{}{webAuthnRoamingData}, nil
}

func flattenWebAuthnPlatform(enabled bool, api *management.Management) ([]interface{}, error) {
	webAuthnPlatformData := make(map[string]interface{})
	webAuthnPlatformData["enabled"] = enabled

	if !enabled {
		return []interface{}{webAuthnPlatformData}, nil
	}

	webAuthnSettings, err := api.Guardian.MultiFactor.WebAuthnPlatform.Read()
	if err != nil {
		return nil, err
	}

	webAuthnPlatformData["override_relying_party"] = webAuthnSettings.GetOverrideRelyingParty()
	webAuthnPlatformData["relying_party_identifier"] = webAuthnSettings.GetRelyingPartyIdentifier()

	return []interface{}{webAuthnPlatformData}, nil
}

func flattenDUO(enabled bool, api *management.Management) ([]interface{}, error) {
	duoData := make(map[string]interface{})
	duoData["enabled"] = enabled

	if !enabled {
		return []interface{}{duoData}, nil
	}

	duoSettings, err := api.Guardian.MultiFactor.DUO.Read()
	if err != nil {
		return nil, err
	}

	duoData["integration_key"] = duoSettings.GetIntegrationKey()
	duoData["secret_key"] = duoSettings.GetSecretKey()
	duoData["hostname"] = duoSettings.GetHostname()

	return []interface{}{duoData}, nil
}

func flattenPush(d *schema.ResourceData, enabled bool, api *management.Management) ([]interface{}, error) {
	pushData := make(map[string]interface{})
	pushData["enabled"] = enabled

	if !enabled {
		return []interface{}{pushData}, nil
	}

	pushProvider, err := api.Guardian.MultiFactor.Push.Provider()
	if err != nil {
		return nil, err
	}
	pushData["provider"] = pushProvider.GetProvider()

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

	if pushProvider.GetProvider() == "sns" {
		amazonSNS, err := api.Guardian.MultiFactor.Push.AmazonSNS()
		if err != nil {
			return nil, err
		}

		pushData["amazon_sns"] = []interface{}{
			map[string]interface{}{
				"aws_access_key_id":                 amazonSNS.GetAccessKeyID(),
				"aws_region":                        amazonSNS.GetRegion(),
				"aws_secret_access_key":             d.Get("push.0.amazon_sns.0.aws_secret_access_key"), // Does not get read back.
				"sns_apns_platform_application_arn": amazonSNS.GetAPNSPlatformApplicationARN(),
				"sns_gcm_platform_application_arn":  amazonSNS.GetGCMPlatformApplicationARN(),
			},
		}
	}

	return []interface{}{pushData}, nil
}

func updatePolicy(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("policy") {
		return nil
	}

	multiFactorPolicies := management.MultiFactorPolicies{}

	policy := d.Get("policy").(string)
	if policy != "never" {
		multiFactorPolicies = append(multiFactorPolicies, policy)
	}

	// If the policy is "never" then the slice needs to be empty.
	return api.Guardian.MultiFactor.UpdatePolicy(&multiFactorPolicies)
}

func updateEmailFactor(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("email") {
		return nil
	}

	enabled := d.Get("email").(bool)
	return api.Guardian.MultiFactor.Email.Enable(enabled)
}

func updateOTPFactor(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("otp") {
		return nil
	}

	enabled := d.Get("otp").(bool)
	return api.Guardian.MultiFactor.OTP.Enable(enabled)
}

func updateRecoveryCodeFactor(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("recovery_code") {
		return nil
	}

	enabled := d.Get("recovery_code").(bool)
	return api.Guardian.MultiFactor.RecoveryCode.Enable(enabled)
}

func updatePhoneFactor(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("phone") {
		return nil
	}

	enabled := d.Get("phone.0.enabled").(bool)

	// Always enable phone factor before configuring it.
	// Otherwise, we encounter an error with message_types.
	if err := api.Guardian.MultiFactor.Phone.Enable(enabled); err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	return configurePhone(d.GetRawConfig(), api)
}

func configurePhone(config cty.Value, api *management.Management) error {
	var err error

	config.GetAttr("phone").ForEachElement(func(_ cty.Value, phone cty.Value) (stop bool) {
		mfaProvider := &management.MultiFactorProvider{
			Provider: value.String(phone.GetAttr("provider")),
		}
		if err = api.Guardian.MultiFactor.Phone.UpdateProvider(mfaProvider); err != nil {
			return true
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
			return true
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
			return true
		}

		if err = api.Guardian.MultiFactor.SMS.UpdateTemplate(
			&management.MultiFactorSMSTemplate{
				EnrollmentMessage:   value.String(config.GetAttr("enrollment_message")),
				VerificationMessage: value.String(config.GetAttr("verification_message")),
			},
		); err != nil {
			return true
		}

		return stop
	})

	return err
}

func updateWebAuthnRoaming(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("webauthn_roaming") {
		return nil
	}

	enabled := d.Get("webauthn_roaming.0.enabled").(bool)
	if err := api.Guardian.MultiFactor.WebAuthnRoaming.Enable(enabled); err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	var err error
	d.GetRawConfig().GetAttr("webauthn_roaming").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		webAuthnSettings := &management.MultiFactorWebAuthnSettings{
			UserVerification:       value.String(config.GetAttr("user_verification")),
			OverrideRelyingParty:   value.Bool(config.GetAttr("override_relying_party")),
			RelyingPartyIdentifier: value.String(config.GetAttr("relying_party_identifier")),
		}

		if webAuthnSettings.String() != "{}" {
			err = api.Guardian.MultiFactor.WebAuthnRoaming.Update(webAuthnSettings)
		}

		return stop
	})

	return err
}

func updateWebAuthnPlatform(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("webauthn_platform") {
		return nil
	}

	enabled := d.Get("webauthn_platform.0.enabled").(bool)
	if err := api.Guardian.MultiFactor.WebAuthnPlatform.Enable(enabled); err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	var err error
	d.GetRawConfig().GetAttr("webauthn_platform").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		webAuthnSettings := &management.MultiFactorWebAuthnSettings{
			OverrideRelyingParty:   value.Bool(config.GetAttr("override_relying_party")),
			RelyingPartyIdentifier: value.String(config.GetAttr("relying_party_identifier")),
		}

		if webAuthnSettings.String() != "{}" {
			err = api.Guardian.MultiFactor.WebAuthnPlatform.Update(webAuthnSettings)
		}

		return stop
	})

	return err
}

func updateDUO(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("duo") {
		return nil
	}

	enabled := d.Get("duo.0.enabled").(bool)
	if err := api.Guardian.MultiFactor.DUO.Enable(enabled); err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	var err error
	d.GetRawConfig().GetAttr("duo").ForEachElement(
		func(_ cty.Value, config cty.Value) (stop bool) {
			duoSettings := &management.MultiFactorDUOSettings{
				SecretKey:      value.String(config.GetAttr("secret_key")),
				Hostname:       value.String(config.GetAttr("hostname")),
				IntegrationKey: value.String(config.GetAttr("integration_key")),
			}

			if duoSettings.String() != "{}" {
				err = api.Guardian.MultiFactor.DUO.Update(duoSettings)
			}

			return stop
		},
	)

	return err
}

func updatePush(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("push") {
		return nil
	}

	enabled := d.Get("push.0.enabled").(bool)
	if err := api.Guardian.MultiFactor.Push.Enable(enabled); err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	var err error
	d.GetRawConfig().GetAttr("push").ForEachElement(func(_ cty.Value, push cty.Value) (stop bool) {
		mfaProvider := &management.MultiFactorProvider{
			Provider: value.String(push.GetAttr("provider")),
		}
		if err = api.Guardian.MultiFactor.Push.UpdateProvider(mfaProvider); err != nil {
			return true
		}

		if d.HasChange("push.0.custom_app") {
			if err = updateCustomApp(push.GetAttr("custom_app"), api); err != nil {
				return true
			}
		}

		if d.HasChange("push.0.amazon_sns") {
			if err = updateAmazonSNS(push.GetAttr("amazon_sns"), api); err != nil {
				return true
			}
		}

		return stop
	})
	return err
}

func updateAmazonSNS(options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.Push.UpdateAmazonSNS(
			&management.MultiFactorProviderAmazonSNS{
				AccessKeyID:                value.String(config.GetAttr("aws_access_key_id")),
				SecretAccessKeyID:          value.String(config.GetAttr("aws_secret_access_key")),
				Region:                     value.String(config.GetAttr("aws_region")),
				APNSPlatformApplicationARN: value.String(config.GetAttr("sns_apns_platform_application_arn")),
				GCMPlatformApplicationARN:  value.String(config.GetAttr("sns_gcm_platform_application_arn")),
			},
		)

		return stop
	})

	return err
}

func updateCustomApp(options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.Push.UpdateCustomApp(
			&management.MultiFactorPushCustomApp{
				AppName:       value.String(config.GetAttr("app_name")),
				AppleAppLink:  value.String(config.GetAttr("apple_app_link")),
				GoogleAppLink: value.String(config.GetAttr("google_app_link")),
			},
		)

		return stop
	})

	return err
}
