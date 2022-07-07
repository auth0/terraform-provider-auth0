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
  push {
    amazon_sns {
      aws_access_key_id = "test1"
      aws_region = "us-west-1"
      aws_secret_access_key = "secretKey"
      sns_apns_platform_application_arn = "test_arn"
      sns_gcm_platform_application_arn = "test_arn"
    }
    custom_app {
      app_name = "CustomApp"
      apple_app_link = "https://itunes.apple.com/us/app/my-app/id123121"
      google_app_link = "https://play.google.com/store/apps/details?id=com.my.app"
    }
  }    
  email = true
  otp = true
  recovery_code = true    
}
```

## Argument Reference

Arguments accepted by this resource include:

* `policy` - (Required) String. Policy to use. Available options are `never`, `all-applications` and `confidence-score`.
The option `confidence-score` means the trigger of MFA will be adaptive. See [Auth0 docs](https://auth0.com/docs/mfa/adaptive-mfa).
* `phone` - (Optional) List(Resource). Configuration settings for the phone MFA. For details, see [Phone](#phone).
* `webauthn_roaming` - (Optional) List(Resource). Configuration settings for the WebAuthn with FIDO Security Keys MFA. For details, see [WebAuthn Roaming](#webauthn-roaming).
* `webauthn_platform` - (Optional) List(Resource). Configuration settings for the WebAuthn with FIDO Device Biometrics MFA. For details, see [WebAuthn Platform](#webauthn-platform).
* `duo` - (Optional) List(Resource). Configuration settings for the Duo MFA. For details, see [Duo](#duo).
* `push` - (Optional) List(Resource). Configuration settings for the Push MFA. For details, see [Push](#push).
* `email` - (Optional) Boolean. Indicates whether email MFA is enabled.
* `otp` - (Optional) Boolean. Indicates whether one time password MFA is enabled.
* `recovery_code` - (Optional) Boolean. Indicates whether recovery code MFA is enabled.

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

### Duo

`duo` supports the following arguments:

* `integration_key` - (Optional) String. Duo client ID, see the Duo documentation for more details on Duo setup.
* `secret_key`- (Optional) String. Duo client secret, see the Duo documentation for more details on Duo setup.
* `hostname`- (Optional) String. Duo API Hostname, see the Duo documentation for more details on Duo setup.

### Push

`push` supports the following arguments:

#### AmazonSNS

`amazon_sns` supports the following arguments:

* `aws_access_key_id` - (Required) String. Your AWS Access Key ID.
* `aws_region`- (Required) String. Your AWS application's region.
* `aws_secret_access_key`- (Required) String. Your AWS Secret Access Key.
* `sns_apns_platform_application_arn`- (Required) String. The Amazon Resource Name for your Apple Push Notification Service.
* `sns_gcm_platform_application_arn`- (Required) String. The Amazon Resource Name for your Firebase Cloud Messaging Service.

#### CustomApp

`custom_app` supports the following arguments:

* `app_name` - (Optional) String. Custom Application Name.
* `apple_app_link`- (Optional) String. Apple App Store URL.
* `google_app_link`- (Optional) String. Google Store URL.


## Attributes Reference

No additional attributes are exported by this resource.

## Import

As this is not a resource identifiable by an ID within the Auth0 Management API, guardian can be imported using a random
string. We recommend [Version 4 UUID](https://www.uuidgenerator.net/version4) e.g.

```shell
$ terraform import auth0_guardian.default 24940d4b-4bd4-44e7-894e-f92e4de36a40
```
