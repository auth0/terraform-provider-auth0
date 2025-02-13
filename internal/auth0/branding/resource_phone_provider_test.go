package branding_test

import (
	"testing"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const testAccPhoneProviderConfigure = `
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

const testAccPhoneProviderUpdate = `
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
				Config: testAccPhoneProviderConfigure,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "name", "twilio"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "disabled", "false"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.delivery_methods.#", "2"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.default_from", "+1234567890"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.sid", "sid1"),
				),
			},
			{
				Config: testAccPhoneProviderUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "name", "custom"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "disabled", "false"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.delivery_methods.#", "1"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.default_from", "+1234567890"),
					resource.TestCheckResourceAttr("auth0_phone_provider.phone_provider_config", "configuration.0.sid", "sid2"),
				),
			},
		},
	})
}
