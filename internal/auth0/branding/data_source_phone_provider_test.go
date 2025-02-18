package branding_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourcePhoneProviderConfig = `
resource "auth0_phone_provider" "phone_provider" {
    name = "twilio"
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



data "auth0_phone_provider" "phone_provider" {
	depends_on = [auth0_phone_provider.phone_provider]
    id = auth0_phone_provider.phone_provider.id
}
`

const testAccDataSourcePhoneProviderUpdate = `
resource "auth0_phone_provider" "phone_provider" {
    name = "custom"
    disabled= false
    credentials  {}

    configuration {
        delivery_methods = ["text", "voice"]
    }
}



data "auth0_phone_provider" "phone_provider" {
	depends_on = [auth0_phone_provider.phone_provider]
    id = "${auth0_phone_provider.phone_provider.id}"
}
`

func TestAccDataPhoneProvider(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePhoneProviderConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_phone_provider.phone_provider", "name", "twilio"),
					resource.TestCheckResourceAttr("data.auth0_phone_provider.phone_provider", "disabled", "false"),
					resource.TestCheckResourceAttr("data.auth0_phone_provider.phone_provider", "configuration.0.delivery_methods.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_phone_provider.phone_provider", "configuration.0.default_from", "+1234567890"),
					resource.TestCheckResourceAttr("data.auth0_phone_provider.phone_provider", "configuration.0.sid", "sid2"),
				),
			},
			{
				Config: testAccDataSourcePhoneProviderUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_phone_provider.phone_provider", "name", "custom"),
					resource.TestCheckResourceAttr("data.auth0_phone_provider.phone_provider", "disabled", "false"),
					resource.TestCheckResourceAttr("data.auth0_phone_provider.phone_provider", "configuration.0.delivery_methods.#", "2"),
				),
			},
		},
	})
}
