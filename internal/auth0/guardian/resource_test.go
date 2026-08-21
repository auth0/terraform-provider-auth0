package guardian_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
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
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccGuardianEmailCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "true"),
				),
			},
			{
				Config: testAccGuardianEmailDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email", "false"),
				),
			},
			{
				Config: testAccGuardianOTPCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "true"),
				),
			},
			{
				Config: testAccGuardianOTPDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "otp", "false"),
				),
			},
			{
				Config: testAccGuardianRecoveryCodeCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "true"),
				),
			},
			{
				Config: testAccGuardianRecoveryCodeDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "recovery_code", "false"),
				),
			},
		},
	})
}

const testAccGuardianPhoneWithMessageTypeSms = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	phone {
		enabled       = true
		message_types = ["sms"]
	}
}
`

const testAccGuardianPhoneWithMessageTypeVoice = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	phone {
		enabled       = true
		message_types = ["voice"]
	}
}
`

const testAccGuardianPhoneWithMessageTypeSmsAndVoice = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	phone {
		enabled       = true
		message_types = ["sms", "voice"]
	}
}
`

const testAccGuardianPhoneDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	phone {
		enabled = false
	}
}
`

func TestAccGuardianPhone(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccGuardianPhoneWithMessageTypeSms,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.message_types.0", "sms"),
				),
			},
			{
				Config: testAccGuardianPhoneWithMessageTypeVoice,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.message_types.0", "voice"),
				),
			},
			{
				Config: testAccGuardianPhoneWithMessageTypeSmsAndVoice,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.message_types.#", "2"),
				),
			},
			{
				Config: testAccGuardianPhoneDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone.0.message_types.#", "0"),
				),
			},
		},
	})
}

const testAccGuardianSettingsCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"

	settings {
		display_remember_me_checkbox   = true
		remember_me_default_value      = false
		mfa_session_inactivity_timeout = 604800
		mfa_session_overall_timeout    = 2592000
	}

	phone_settings {
		otp_length          = 6
		otp_expiration_time = 300
	}

	email_settings {
		otp_length          = 6
		otp_expiration_time = 300
	}
}
`

const testAccGuardianSettingsUpdate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"

	settings {
		display_remember_me_checkbox   = false
		remember_me_default_value      = true
		mfa_session_inactivity_timeout = 7200
		mfa_session_overall_timeout    = 86400
	}

	phone_settings {
		otp_length          = 8
		otp_expiration_time = 600
	}

	email_settings {
		otp_length          = 10
		otp_expiration_time = 3600
	}
}
`

// testAccGuardianSettingsBoundaries exercises the lower bound of every
// ValidateFunc range: otp_length 4, otp_expiration_time 30, and the shared
// 3600 second floor on both MFA session timeouts.
const testAccGuardianSettingsBoundaries = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"

	settings {
		display_remember_me_checkbox   = true
		remember_me_default_value      = true
		mfa_session_inactivity_timeout = 3600
		mfa_session_overall_timeout    = 3600
	}

	phone_settings {
		otp_length          = 4
		otp_expiration_time = 30
	}

	email_settings {
		otp_length          = 4
		otp_expiration_time = 30
	}
}
`

// testAccGuardianSettingsOmitted drops all three blocks from the config. They are
// Optional+Computed, so the API values must be read back into state rather than
// being blanked, and no write should be attempted for the removed blocks.
const testAccGuardianSettingsOmitted = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
}
`

func TestAccGuardianSettings(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccGuardianSettingsCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.display_remember_me_checkbox", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.remember_me_default_value", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.mfa_session_inactivity_timeout", "604800"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.mfa_session_overall_timeout", "2592000"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone_settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone_settings.0.otp_length", "6"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone_settings.0.otp_expiration_time", "300"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email_settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email_settings.0.otp_length", "6"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email_settings.0.otp_expiration_time", "300"),
				),
			},
			{
				Config: testAccGuardianSettingsUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.display_remember_me_checkbox", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.remember_me_default_value", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.mfa_session_inactivity_timeout", "7200"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.mfa_session_overall_timeout", "86400"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone_settings.0.otp_length", "8"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone_settings.0.otp_expiration_time", "600"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email_settings.0.otp_length", "10"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email_settings.0.otp_expiration_time", "3600"),
				),
			},
			{
				Config: testAccGuardianSettingsBoundaries,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.mfa_session_inactivity_timeout", "3600"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.mfa_session_overall_timeout", "3600"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone_settings.0.otp_length", "4"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone_settings.0.otp_expiration_time", "30"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email_settings.0.otp_length", "4"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email_settings.0.otp_expiration_time", "30"),
				),
			},
			{
				// The blocks are Optional+Computed, so removing them from the config
				// keeps the last applied values in state instead of clearing them.
				Config: testAccGuardianSettingsOmitted,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "settings.0.mfa_session_inactivity_timeout", "3600"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone_settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "phone_settings.0.otp_length", "4"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email_settings.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "email_settings.0.otp_length", "4"),
				),
			},
		},
	})
}

const testAccConfigureWebAuthnRoamingCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	webauthn_roaming {
		enabled = true
	}
}
`

const testAccConfigureWebAuthnRoamingUpdate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	webauthn_roaming {
		enabled = true
		user_verification = "required"
	}
}
`

const testAccConfigureWebAuthnRoamingDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	webauthn_roaming {
		enabled = false
	}
}
`

func TestAccGuardianWebAuthnRoaming(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccConfigureWebAuthnRoamingCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.0.enabled", "true"),
				),
			},
			{
				Config: testAccConfigureWebAuthnRoamingUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.0.user_verification", "required"),
				),
			},
			{
				Config: testAccConfigureWebAuthnRoamingDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_roaming.0.enabled", "false"),
				),
			},
		},
	})
}

const testAccConfigureWebAuthnPlatformCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	webauthn_platform {
		enabled = true
	}
}
`

const testAccConfigureWebAuthnPlatformDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	webauthn_platform {
		enabled = false
	}
}
`

func TestAccGuardianWebAuthnPlatform(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccConfigureWebAuthnPlatformCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.0.enabled", "true"),
				),
			},
			{
				Config: testAccConfigureWebAuthnPlatformDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "webauthn_platform.0.enabled", "false"),
				),
			},
		},
	})
}

const testAccConfigureDUOCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	duo {
		enabled = true
		integration_key = "someKey"
		secret_key = "someSecret"
		hostname = "api-hostname"
	}
}
`

const testAccConfigureDUODelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	duo {
		enabled = false
	}
}
`

func TestAccGuardianDUO(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccConfigureDUOCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.hostname", "api-hostname"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.secret_key", "someSecret"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.integration_key", "someKey"),
				),
			},
			{
				Config: testAccConfigureDUODelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.hostname", ""),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.secret_key", ""),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "duo.0.integration_key", ""),
				),
			},
		},
	})
}

const testAccConfigurePushCreate = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	push {
		enabled  = true
		provider = "guardian"
	}
}
`

const testAccConfigurePushUpdateAmazonSNS = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	push {
		enabled  = true
		provider = "sns"

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
		enabled  = true
		provider = "sns"

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

const testAccConfigurePushUpdateDirectAPNS = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	push {
		enabled  = true
		provider = "direct"

		direct_apns {
			sandbox = false
			bundle_id = "com.my.app"
			p12 = %q
		}
	}
}
`

const testAccConfigurePushUpdateDirectFCM = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	push {
		enabled  = true
		provider = "direct"

		direct_fcm {
			server_key = "abc123"
		}
	}
}
`

const testAccConfigurePushDelete = `
resource "auth0_guardian" "foo" {
	policy = "all-applications"
	push {
		enabled = false
	}
}
`

func TestAccGuardianPush(t *testing.T) {
	apnsCertificate, err := os.ReadFile("./../../../test/data/apns.p12")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurePushCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.provider", "guardian"),
				),
			},
			{
				Config: testAccConfigurePushUpdateAmazonSNS,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.provider", "sns"),
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
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.provider", "sns"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.aws_access_key_id", "test1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.aws_region", "us-west-1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.aws_secret_access_key", "secretKey"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.sns_apns_platform_application_arn", "test_arn"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.amazon_sns.0.sns_gcm_platform_application_arn", "test_arn"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.custom_app.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.custom_app.0.app_name", "CustomApp"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.custom_app.0.apple_app_link", "https://itunes.apple.com/us/app/my-app/id123121"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.custom_app.0.google_app_link", "https://play.google.com/store/apps/details?id=com.my.app"),
				),
			},
			{
				Config: fmt.Sprintf(testAccConfigurePushUpdateDirectAPNS, apnsCertificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.provider", "direct"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.direct_apns.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.direct_apns.0.sandbox", "false"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.direct_apns.0.bundle_id", "com.my.app"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.direct_apns.0.enabled", "true"),
					resource.TestCheckResourceAttrSet("auth0_guardian.foo", "push.0.direct_apns.0.p12"),
				),
			},
			{
				Config: testAccConfigurePushUpdateDirectFCM,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.provider", "direct"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.direct_fcm.#", "1"),
				),
			},
			{
				Config: testAccConfigurePushDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_guardian.foo", "policy", "all-applications"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.#", "1"),
					resource.TestCheckResourceAttr("auth0_guardian.foo", "push.0.enabled", "false"),
				),
			},
		},
	})
}
