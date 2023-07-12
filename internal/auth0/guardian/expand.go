package guardian

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func updatePolicy(ctx context.Context, d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("policy") {
		return nil
	}

	multiFactorPolicies := management.MultiFactorPolicies{}

	policy := d.Get("policy").(string)
	if policy != "never" {
		multiFactorPolicies = append(multiFactorPolicies, policy)
	}

	// If the policy is "never" then the slice needs to be empty.
	return api.Guardian.MultiFactor.UpdatePolicy(ctx, &multiFactorPolicies)
}

func updateEmailFactor(ctx context.Context, d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("email") {
		return nil
	}

	enabled := d.Get("email").(bool)
	return api.Guardian.MultiFactor.Email.Enable(ctx, enabled)
}

func updateOTPFactor(ctx context.Context, d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("otp") {
		return nil
	}

	enabled := d.Get("otp").(bool)
	return api.Guardian.MultiFactor.OTP.Enable(ctx, enabled)
}

func updateRecoveryCodeFactor(ctx context.Context, d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("recovery_code") {
		return nil
	}

	enabled := d.Get("recovery_code").(bool)
	return api.Guardian.MultiFactor.RecoveryCode.Enable(ctx, enabled)
}

func updatePhoneFactor(ctx context.Context, d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("phone") {
		return nil
	}

	enabled := d.Get("phone.0.enabled").(bool)

	// Always enable phone factor before configuring it.
	// Otherwise, we encounter an error with message_types.
	if err := api.Guardian.MultiFactor.Phone.Enable(ctx, enabled); err != nil {
		return err
	}
	if !enabled {
		return nil
	}

	return configurePhone(ctx, d.GetRawConfig(), api)
}

func configurePhone(ctx context.Context, config cty.Value, api *management.Management) error {
	var err error

	config.GetAttr("phone").ForEachElement(func(_ cty.Value, phone cty.Value) (stop bool) {
		mfaProvider := &management.MultiFactorProvider{
			Provider: value.String(phone.GetAttr("provider")),
		}
		if err = api.Guardian.MultiFactor.Phone.UpdateProvider(ctx, mfaProvider); err != nil {
			return true
		}

		options := phone.GetAttr("options")
		switch mfaProvider.GetProvider() {
		case "twilio":
			if err = updateTwilioOptions(ctx, options, api); err != nil {
				return true
			}
		case "auth0", "phone-message-hook":
			if err = updateAuth0Options(ctx, options, api); err != nil {
				return true
			}
		}

		messageTypes := &management.PhoneMessageTypes{
			MessageTypes: value.Strings(phone.GetAttr("message_types")),
		}
		if err = api.Guardian.MultiFactor.Phone.UpdateMessageTypes(ctx, messageTypes); err != nil {
			return true
		}

		return stop
	})

	return err
}

func updateAuth0Options(ctx context.Context, options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.SMS.UpdateTemplate(
			ctx,
			&management.MultiFactorSMSTemplate{
				EnrollmentMessage:   value.String(config.GetAttr("enrollment_message")),
				VerificationMessage: value.String(config.GetAttr("verification_message")),
			},
		)

		return stop
	})

	return err
}

func updateTwilioOptions(ctx context.Context, options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		if err = api.Guardian.MultiFactor.SMS.UpdateTwilio(
			ctx,
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
			ctx,
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

func updateWebAuthnRoaming(ctx context.Context, d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("webauthn_roaming") {
		return nil
	}

	enabled := d.Get("webauthn_roaming.0.enabled").(bool)
	if err := api.Guardian.MultiFactor.WebAuthnRoaming.Enable(ctx, enabled); err != nil {
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
			err = api.Guardian.MultiFactor.WebAuthnRoaming.Update(ctx, webAuthnSettings)
		}

		return stop
	})

	return err
}

func updateWebAuthnPlatform(ctx context.Context, d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("webauthn_platform") {
		return nil
	}

	enabled := d.Get("webauthn_platform.0.enabled").(bool)
	if err := api.Guardian.MultiFactor.WebAuthnPlatform.Enable(ctx, enabled); err != nil {
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
			err = api.Guardian.MultiFactor.WebAuthnPlatform.Update(ctx, webAuthnSettings)
		}

		return stop
	})

	return err
}

func updateDUO(ctx context.Context, d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("duo") {
		return nil
	}

	enabled := d.Get("duo.0.enabled").(bool)
	if err := api.Guardian.MultiFactor.DUO.Enable(ctx, enabled); err != nil {
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
				err = api.Guardian.MultiFactor.DUO.Update(ctx, duoSettings)
			}

			return stop
		},
	)

	return err
}

func updatePush(ctx context.Context, d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("push") {
		return nil
	}

	enabled := d.Get("push.0.enabled").(bool)
	if err := api.Guardian.MultiFactor.Push.Enable(ctx, enabled); err != nil {
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
		if err = api.Guardian.MultiFactor.Push.UpdateProvider(ctx, mfaProvider); err != nil {
			return true
		}

		if d.HasChange("push.0.custom_app") {
			if err = updateCustomApp(ctx, push.GetAttr("custom_app"), api); err != nil {
				return true
			}
		}

		if d.HasChange("push.0.direct_apns") {
			if err = updateDirectAPNS(ctx, push.GetAttr("direct_apns"), api); err != nil {
				return true
			}
		}

		if d.HasChange("push.0.direct_fcm") {
			if err = updateDirectFCM(ctx, push.GetAttr("direct_fcm"), api); err != nil {
				return true
			}
		}

		if d.HasChange("push.0.amazon_sns") {
			if err = updateAmazonSNS(ctx, push.GetAttr("amazon_sns"), api); err != nil {
				return true
			}
		}

		return stop
	})
	return err
}

func updateAmazonSNS(ctx context.Context, options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.Push.UpdateAmazonSNS(
			ctx,
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

func updateCustomApp(ctx context.Context, options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.Push.UpdateCustomApp(
			ctx,
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

func updateDirectAPNS(ctx context.Context, options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.Push.UpdateDirectAPNS(
			ctx,
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

func updateDirectFCM(ctx context.Context, options cty.Value, api *management.Management) error {
	var err error

	options.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		err = api.Guardian.MultiFactor.Push.UpdateDirectFCM(
			ctx,
			&management.MultiFactorPushDirectFCM{
				ServerKey: value.String(config.GetAttr("server_key")),
			},
		)

		return stop
	})

	return err
}
