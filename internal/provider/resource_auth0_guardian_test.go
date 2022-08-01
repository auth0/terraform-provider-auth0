package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const testAccGuardianEmailCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	email  = true
}
`

const testAccGuardianEmailDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	email  = false
}
`

const testAccGuardianOTPCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	otp    = true
}
`

const testAccGuardianOTPDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	otp    = false
}
`

const testAccGuardianRecoveryCodeCreate = `
resource "auth0_guardian" "foo" {
	policy        = "all-applications"
	recovery_code = true
}
`

const testAccGuardianRecoveryCodeDelete = `
resource "auth0_guardian" "foo" {
	policy        = "all-applications"
	recovery_code = false
}
`

func TestAccGuardian(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardianEmailCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
			{
				Config: testAccGuardianEmailDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
			{
				Config: testAccGuardianOTPCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
			{
				Config: testAccGuardianOTPDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
			{
				Config: testAccGuardianRecoveryCodeCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "true"),
				),
			},
			{
				Config: testAccGuardianRecoveryCodeDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
		},
	})
}

const testAccGuardianPhoneWithCustomProviderAndNoOptions = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	phone {
		provider      = "phone-message-hook"
		message_types = ["sms"]
	}
}
`

const testAccGuardianPhoneWithCustomProviderAndEmptyOptions = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	phone {
		provider      = "phone-message-hook"
		message_types = ["sms"]
		options {}
	}
}
`

const testAccGuardianPhoneWithAuth0Provider = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	phone {
		provider      = "auth0"
		message_types = ["voice"]
		options {
			enrollment_message   = "enroll foo"
			verification_message = "verify foo"
		}
	}
}
`

const testAccGuardianPhoneWithTwilioProvider = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	phone {
		provider      = "twilio"
		message_types = ["sms"]
		options {
			enrollment_message    = "enroll foo"
			verification_message  = "verify foo"
			from                  = "from bar"
			messaging_service_sid = "foo"
			auth_token            = "bar"
			sid                   = "foo"
		}
	}
}
`

const testAccGuardianPhoneDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
}
`

func TestAccGuardianPhone(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardianPhoneWithCustomProviderAndNoOptions,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.message_types.0", "sms"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.provider", "phone-message-hook"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.#", "1"),
				),
			},
			{
				Config: testAccGuardianPhoneWithCustomProviderAndEmptyOptions,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.message_types.0", "sms"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.provider", "phone-message-hook"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.#", "1"),
				),
			},
			{
				Config: testAccGuardianPhoneWithAuth0Provider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.message_types.0", "voice"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.provider", "auth0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.0.enrollment_message", "enroll foo"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.0.verification_message", "verify foo"),
				),
			},
			{
				Config: testAccGuardianPhoneWithTwilioProvider,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.message_types.0", "sms"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.provider", "twilio"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.0.enrollment_message", "enroll foo"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.0.verification_message", "verify foo"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.0.from", "from bar"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.0.messaging_service_sid", "foo"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.0.auth_token", "bar"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.options.0.sid", "foo"),
				),
			},
			{
				Config: testAccGuardianPhoneDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
		},
	})
}

const testAccConfigureWebAuthnRoamingCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	webauthn_roaming {}
}
`

const testAccConfigureWebAuthnRoamingUpdate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	webauthn_roaming {
		user_verification = "required"
	}
}
`

const testAccConfigureWebAuthnRoamingDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
}
`

func TestAccGuardianWebAuthnRoaming(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigureWebAuthnRoamingCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "1"),
				),
			},
			{
				Config: testAccConfigureWebAuthnRoamingUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.0.user_verification", "required"),
				),
			},
			{
				Config: testAccConfigureWebAuthnRoamingDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
		},
	})
}

const testAccConfigureWebAuthnPlatformCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	webauthn_platform {}
}
`

const testAccConfigureWebAuthnPlatformDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
}
`

func TestAccGuardianWebAuthnPlatform(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigureWebAuthnPlatformCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "1"),
				),
			},
			{
				Config: testAccConfigureWebAuthnPlatformDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
		},
	})
}

const testAccConfigureDUOCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	duo {
		integration_key = "someKey"
		secret_key = "someSecret"
		hostname = "api-hostname"
	}
}
`

const testAccConfigureDUODelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
}
`

func TestAccGuardianDUO(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigureDUOCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.hostname", "api-hostname"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.secret_key", "someSecret"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.integration_key", "someKey"),
				),
			},
			{
				Config: testAccConfigureDUODelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
		},
	})
}

const testAccConfigurePushCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	push {}
}
`

const testAccConfigurePushUpdateAmazonSNS = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	push {
		amazon_sns {
			aws_access_key_id = "test1"
			aws_region = "us-west-1"
			aws_secret_access_key = "secretKey"
			sns_apns_platform_application_arn = "test_arn"
			sns_gcm_platform_application_arn = "test_arn"
		}
	}
}
`

const testAccConfigurePushUpdateCustomApp = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
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
}
`

const testAccConfigurePushDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
}
`

func TestAccGuardianPush(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurePushCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "1"),
				),
			},
			{
				Config: testAccConfigurePushUpdateAmazonSNS,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.aws_access_key_id", "test1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.aws_region", "us-west-1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.aws_secret_access_key", "secretKey"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.sns_apns_platform_application_arn", "test_arn"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.sns_gcm_platform_application_arn", "test_arn"),
				),
			},
			{
				Config: testAccConfigurePushUpdateCustomApp,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.custom_app.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.custom_app.0.app_name", "CustomApp"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.custom_app.0.apple_app_link", "https://itunes.apple.com/us/app/my-app/id123121"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.custom_app.0.google_app_link", "https://play.google.com/store/apps/details?id=com.my.app"),
				),
			},
			{
				Config: testAccConfigurePushDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "0"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
		},
	})
}
