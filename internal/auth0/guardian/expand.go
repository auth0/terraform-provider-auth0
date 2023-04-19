package guardian

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

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
		case "auth0", "phone-message-hook":
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

		if d.HasChange("push.0.direct_apns") {
			if err = updateDirectAPNS(push.GetAttr("direct_apns"), api); err != nil {
				return true
			}
		}

		if d.HasChange("push.0.direct_fcm") {
			if err = updateDirectFCM(push.GetAttr("direct_fcm"), api); err != nil {
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

func updateDirectAPNS(options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.Push.UpdateDirectAPNS(
			&management.MultiFactorPushDirectAPNS{
				Sandbox:  value.Bool(config.GetAttr("sandbox")),
				BundleID: value.String(config.GetAttr("bundle_id")),
				P12:      value.String(config.GetAttr("p12")),
			},
		)

		return stop
	})

	return err
}

func updateDirectFCM(options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.Push.UpdateDirectFCM(
			&management.MultiFactorPushDirectFCM{
				ServerKey: value.String(config.GetAttr("server_key")),
			},
		)

		return stop
	})

	return err
}
