package branding_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPhoneProviderConfigureWithoutDeliveryMethods = `
resource "auth0_phone_provider" "phone_provider_config" {
    name = "twilio"
    disabled= false
    credentials  {
        auth_token = "auth123"
    }

    configuration {
		default_from = "+1234567890"
        sid = "sid1"
    }
}`

const testAccPhoneProviderConfigureWithInvalidDeliveryMethods = `
resource "auth0_phone_provider" "phone_provider_config" {
    name = "twilio"
    disabled= false
    credentials  {
        auth_token = "auth123"
    }

    configuration {
	delivery_methods = ["xxx","yyy"]
		default_from = "+1234567890"
        sid = "sid1"
    }
}`

const testAccPhoneProviderConfigureWithSIDWithoutDefaultFrom = `
resource "auth0_phone_provider" "phone_provider_config" {
    name = "twilio"
    disabled= false
    credentials  {
        auth_token = "auth123"
    }

    configuration {
        delivery_methods = ["text","voice"]
        sid = "sid1"
    }
}`

const testAccPhoneProviderConfigureWithDefaultFromWithoutSid = `
resource "auth0_phone_provider" "phone_provider_config" {
    name = "twilio"
    disabled= false
    credentials  {
        auth_token = "auth123"
    }

    configuration {
        delivery_methods = ["text","voice"]
        default_from = "+1234567890"
    }
}`

const testAccPhoneProviderTwilio = `
resource "auth0_phone_provider" "phone_provider_config" {
    name = "twilio"
    disabled= false
    credentials  {
        auth_token = "auth123"
    }

    configuration {
        delivery_methods = ["text","voice"]
        default_from = "+1234567890"
        sid = "sid1"
    }
}
`

const testAccPhoneProviderCustom = `
resource "auth0_phone_provider" "phone_provider_config" {
    name = "custom"
    disabled= false
    credentials  {
        auth_token = "auth123"
    }

    configuration {
        delivery_methods = ["text"]
        default_from = "+1234567890"
        sid = "sid2"

    }
}
`

func TestAccCheckPhoneProvider(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccPhoneProviderConfigureWithoutDeliveryMethods,
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      testAccPhoneProviderConfigureWithInvalidDeliveryMethods,
				ExpectError: regexp.MustCompile("expected configuration.0.delivery_methods.0 to be one of"),
			},
			{
				Config:      testAccPhoneProviderConfigureWithSIDWithoutDefaultFrom,
				ExpectError: regexp.MustCompile("Bad Operation on Notification Resource"),
			},
			{
				Config:      testAccPhoneProviderConfigureWithDefaultFromWithoutSid,
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config: testAccPhoneProviderTwilio,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "name", "twilio"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "disabled", "false"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.delivery_methods.#", "2"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.default_from", "+1234567890"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.sid", "sid1"),
				),
			},
			{
				Config: testAccPhoneProviderCustom,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "name", "custom"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "disabled", "false"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.delivery_methods.#", "1"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.default_from", "+1234567890"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.sid", "sid2"),
				),
			},
			{
				Config: testAccPhoneProviderTwilio,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "name", "twilio"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "disabled", "false"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.delivery_methods.#", "2"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.default_from", "+1234567890"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.sid", "sid1"),
				),
			},
		},
	})
}
