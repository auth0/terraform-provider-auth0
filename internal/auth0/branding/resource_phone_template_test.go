package branding_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPhoneNotificationTemplateChangePassword = `
resource "auth0_branding_phone_notification_template" "change_password" {
	type     = "change_password"
	disabled = true
	content {
		from = "+918095250531"
		body {
			text  = "Your password has been changed"
			voice = "Your password has been changed"
		}
	}
}
`

const testAccPhoneNotificationTemplateOTPEnroll = `
resource "auth0_branding_phone_notification_template" "otp_enroll" {
	type     = "otp_enroll"
	disabled = false
	content {
		from = "+1234567890"
		body {
			text  = "Your enrollment code is: @{code}"
			voice = "Your enrollment code is @{code}"
		}
	}
}
`

const testAccPhoneNotificationTemplateOTPEnrollUpdated = `
resource "auth0_branding_phone_notification_template" "otp_enroll" {
	type     = "otp_enroll"
	disabled = true

	content {
		from = "+9876543210"
		body {
			text  = "Updated enrollment code: @{code}"
			voice = "Updated enrollment code: @{code}"
		}
	}
}
`

const testAccPhoneNotificationTemplateOTPVerify = `
resource "auth0_branding_phone_notification_template" "otp_verify" {
	type     = "otp_verify"
	disabled = false

	content {
		from = "+1111111111"
		body {
			text  = "Your verification code is: @{code}"
			voice = "Your verification code is @{code}"
		}
	}
}
`

const testAccPhoneNotificationTemplateBlockedAccount = `
resource "auth0_branding_phone_notification_template" "blocked_account" {
	type     = "blocked_account"
	disabled = false
	content {
		body {
			text  = "Your account has been blocked"
			voice = "Your account has been blocked"
		}
	}
}
`

const testAccDataSourcePhoneNotificationTemplateOTPEnroll = `
resource "auth0_branding_phone_notification_template" "otp_enroll" {
	type     = "otp_enroll"
	disabled = false
	content {
		from = "+1234567890"
		body {
			text  = "Your enrollment code is: @{code}"
			voice = "Your enrollment code is @{code}"
		}
	}
}

data "auth0_branding_phone_notification_template" "otp_enroll" {
	depends_on = [auth0_branding_phone_notification_template.otp_enroll]
	template_id = auth0_branding_phone_notification_template.otp_enroll.id
}
`

const testAccDataSourcePhoneNotificationTemplateChangePassword = `
resource "auth0_branding_phone_notification_template" "change_password" {
	type     = "change_password"
	disabled = true

	content {
		from = "+918095250531"
		body {
			text  = "Your password has been changed"
			voice = "Your password has been changed"
		}
	}
}

data "auth0_branding_phone_notification_template" "change_password" {
	depends_on = [auth0_branding_phone_notification_template.change_password]
	template_id = auth0_branding_phone_notification_template.change_password.id
}
`

// TestAccPhoneNotificationTemplate tests creation, update, and basic operations.
func TestAccPhoneNotificationTemplate(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneNotificationTemplateChangePassword,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.change_password",
						"type",
						"change_password",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.change_password",
						"disabled",
						"true",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.change_password",
						"content.0.from",
						"+918095250531",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.change_password",
						"content.0.body.0.text",
						"Your password has been changed",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.change_password",
						"content.0.body.0.voice",
						"Your password has been changed",
					),
					// Verify computed fields exist and are not empty.
					resource.TestCheckResourceAttrSet(
						"auth0_branding_phone_notification_template.change_password",
						"id",
					),
					resource.TestCheckResourceAttrSet(
						"auth0_branding_phone_notification_template.change_password",
						"template_id",
					),
					resource.TestCheckResourceAttrSet(
						"auth0_branding_phone_notification_template.change_password",
						"channel",
					),
					resource.TestCheckResourceAttrSet(
						"auth0_branding_phone_notification_template.change_password",
						"tenant",
					),
					resource.TestCheckResourceAttrSet(
						"auth0_branding_phone_notification_template.change_password",
						"customizable",
					),
					resource.TestCheckResourceAttrSet(
						"auth0_branding_phone_notification_template.change_password",
						"content.0.syntax",
					),
				),
			},
			{
				Config: testAccPhoneNotificationTemplateOTPEnroll,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"type",
						"otp_enroll",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"disabled",
						"false",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"content.0.from",
						"+1234567890",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"content.0.body.0.text",
						"Your enrollment code is: @{code}",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"content.0.body.0.voice",
						"Your enrollment code is @{code}",
					),
				),
			},
			{
				Config: testAccPhoneNotificationTemplateOTPEnrollUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"type",
						"otp_enroll",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"disabled",
						"true",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"content.0.from",
						"+9876543210",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"content.0.body.0.text",
						"Updated enrollment code: @{code}",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"content.0.body.0.voice",
						"Updated enrollment code: @{code}",
					),
				),
			},
			{
				Config: testAccPhoneNotificationTemplateOTPVerify,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_verify",
						"type",
						"otp_verify",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_verify",
						"disabled",
						"false",
					),
				),
			},
			{
				Config: testAccPhoneNotificationTemplateBlockedAccount,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.blocked_account",
						"type",
						"blocked_account",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.blocked_account",
						"disabled",
						"false",
					),
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.blocked_account",
						"content.0.body.0.text",
						"Your account has been blocked",
					),
				),
			},
		},
	})
}

// TestAccDataSourcePhoneNotificationTemplate tests data source retrieval.
func TestAccDataSourcePhoneNotificationTemplate(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePhoneNotificationTemplateOTPEnroll,
				Check: resource.ComposeTestCheckFunc(
					// Verify resource was created.
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.otp_enroll",
						"type",
						"otp_enroll",
					),
					// Verify data source can read the resource.
					resource.TestCheckResourceAttr(
						"data.auth0_branding_phone_notification_template.otp_enroll",
						"type",
						"otp_enroll",
					),
					resource.TestCheckResourceAttr(
						"data.auth0_branding_phone_notification_template.otp_enroll",
						"disabled",
						"false",
					),
					resource.TestCheckResourceAttr(
						"data.auth0_branding_phone_notification_template.otp_enroll",
						"content.0.from",
						"+1234567890",
					),
					resource.TestCheckResourceAttr(
						"data.auth0_branding_phone_notification_template.otp_enroll",
						"content.0.body.0.text",
						"Your enrollment code is: @{code}",
					),
					// Verify computed fields are present in data source.
					resource.TestCheckResourceAttrSet(
						"data.auth0_branding_phone_notification_template.otp_enroll",
						"template_id",
					),
					resource.TestCheckResourceAttrSet(
						"data.auth0_branding_phone_notification_template.otp_enroll",
						"channel",
					),
					resource.TestCheckResourceAttrSet(
						"data.auth0_branding_phone_notification_template.otp_enroll",
						"tenant",
					),
					resource.TestCheckResourceAttrSet(
						"data.auth0_branding_phone_notification_template.otp_enroll",
						"customizable",
					),
					resource.TestCheckResourceAttrSet(
						"data.auth0_branding_phone_notification_template.otp_enroll",
						"content.0.syntax",
					),
				),
			},
			{
				Config: testAccDataSourcePhoneNotificationTemplateChangePassword,
				Check: resource.ComposeTestCheckFunc(
					// Verify resource was created.
					resource.TestCheckResourceAttr(
						"auth0_branding_phone_notification_template.change_password",
						"type",
						"change_password",
					),
					// Verify data source can read the resource.
					resource.TestCheckResourceAttr(
						"data.auth0_branding_phone_notification_template.change_password",
						"type",
						"change_password",
					),
					resource.TestCheckResourceAttr(
						"data.auth0_branding_phone_notification_template.change_password",
						"disabled",
						"true",
					),
					resource.TestCheckResourceAttr(
						"data.auth0_branding_phone_notification_template.change_password",
						"content.0.from",
						"+918095250531",
					),
					resource.TestCheckResourceAttr(
						"data.auth0_branding_phone_notification_template.change_password",
						"content.0.body.0.text",
						"Your password has been changed",
					),
					// Verify IDs match between resource and data source.
					resource.TestCheckResourceAttrPair(
						"data.auth0_branding_phone_notification_template.change_password",
						"template_id",
						"auth0_branding_phone_notification_template.change_password",
						"template_id",
					),
				),
			},
		},
	})
}

// TestAccPhoneNotificationTemplateInvalidType tests validation of template type.
func TestAccPhoneNotificationTemplateInvalidType(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
resource "auth0_branding_phone_notification_template" "invalid" {
	type     = "invalid_type"
	disabled = false

	content {
		from = "+1234567890"
		body {
			text = "Test"
		}
	}
}
`,
				ExpectError: regexp.MustCompile("expected type to be one of"),
			},
		},
	})
}
