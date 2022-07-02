---
layout: "auth0"
page_title: "Auth0: auth0_guardian"
description: |-
  With this resource, you can configure MFA options.
---

# auth0_guardian

Multi-Factor Authentication works by requiring additional factors during the login process to prevent unauthorized
access. With this resource you can configure some options available for MFA.

## Example Usage

```hcl
resource "auth0_guardian" "default" {
  policy = "all-applications"
  webauthn_roaming {
      user_verification = "required"
  } 
  webauthn_roaming {}
  phone {
    provider      = "auth0"
    message_types = ["sms"]
    options {
      enrollment_message   = "{{code}} is your verification code for {{tenant.friendly_name}}. Please enter this code to verify your enrollment."
      verification_message = "{{code}} is your verification code for {{tenant.friendly_name}}."
    }
  }
  email = true
  otp = true
}
```

## Argument Reference

Arguments accepted by this resource include:

* `policy` - (Required) String. Policy to use. Available options are `never`, `all-applications` and `confidence-score`.
The option `confidence-score` means the trigger of MFA will be adaptive. See [Auth0 docs](https://auth0.com/docs/mfa/adaptive-mfa).
* `phone` - (Optional) List(Resource). Configuration settings for the phone MFA. For details, see [Phone](#phone).
* `webauthn_roaming` - (Optional) List(Resource). Configuration settings for the WebAuthn with FIDO Security Keys MFA. For details, see [WebAuthn Roaming](#webauthn-roaming).
* `webauthn_platform` - (Optional) List(Resource). Configuration settings for the WebAuthn with FIDO Device Biometrics MFA. For details, see [WebAuthn Platform](#webauthn-platform).
* `email` - (Optional) Boolean. Indicates whether email MFA is enabled.
* `OTP` - (Optional) Boolean. Indicates whether one time password MFA is enabled.

### Phone

`phone` supports the following arguments:

* `provider` - (Required) String, Case-sensitive. Provider to use, one of `auth0`, `twilio` or `phone-message-hook`.
* `message_types` - (Required) List(String). Message types to use, array of `sms` and or `voice`. Adding both to array should enable the user to choose.
* `options`- (Required) List(Resource). Options for the various providers. See [Options](#options).

#### Options
`options` supports different arguments depending on the provider specified in [Phone](#phone).

##### Auth0
* `enrollment_message` (Optional) String. This message will be sent whenever a user enrolls a new device for the first time using MFA. Supports liquid syntax, see [Auth0 docs](https://auth0.com/docs/mfa/customize-sms-or-voice-messages).
* `verification_message` (Optional) String. This message will be sent whenever a user logs in after the enrollment. Supports liquid syntax, see [Auth0 docs](https://auth0.com/docs/mfa/customize-sms-or-voice-messages).

##### Twilio
* `enrollment_message` (Optional) String. This message will be sent whenever a user enrolls a new device for the first time using MFA. Supports liquid syntax, see [Auth0 docs](https://auth0.com/docs/mfa/customize-sms-or-voice-messages).
* `verification_message` (Optional) String. This message will be sent whenever a user logs in after the enrollment. Supports liquid syntax, see [Auth0 docs](https://auth0.com/docs/mfa/customize-sms-or-voice-messages).
* `sid`(Optional) String.
* `auth_token`(Optional) String.
* `from` (Optional) String.
* `messaging_service_sid`(Optional) String.

##### Phone message hook

Options have to be empty. Custom code has to be written in a phone message hook.
See [phone message hook docs](https://auth0.com/docs/hooks/extensibility-points/send-phone-message).

### WebAuthn Roaming

`webauth_roaming` supports the following arguments:

* `user_verification` - (Optional) String. User verification, one of `discouraged`, `preferred` or `required`.
* `override_relying_party` - (Optional) Bool. The Relying Party is the domain for which the WebAuthn keys will be issued, set to true if you are customizing the identifier. 
* `relying_party_identifier`- (Optional) String. The Relying Party should be a suffix of the custom domain.

### WebAuthn Platform

`webauth_roaming` supports the following arguments:

* `override_relying_party` - (Optional) Bool. The Relying Party is the domain for which the WebAuthn keys will be issued, set to true if you are customizing the identifier.
* `relying_party_identifier`- (Optional) String. The Relying Party should be a suffix of the custom domain.

## Attributes Reference

No additional attributes are exported by this resource.

## Import

As this is not a resource identifiable by an ID within the Auth0 Management API, guardian can be imported using a random
string. We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4) e.g.

```shell
$ terraform import auth0_guardian.default 24940d4b-4bd4-44e7-894e-f92e4de36a40
```
