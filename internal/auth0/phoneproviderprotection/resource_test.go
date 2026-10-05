package phoneproviderprotection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPhoneProviderProtectionCreate = `
resource "auth0_phone_provider_protection" "test" {
  type = "default"
}
`

const testAccPhoneProviderProtectionUpdate = `
resource "auth0_phone_provider_protection" "test" {
  type = "exponential"
}
`

const testAccPhoneProviderProtectionReset = `
resource "auth0_phone_provider_protection" "test" {
  type = "default"
}
`

func TestAccPhoneProviderProtection(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPhoneProviderProtectionCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_phone_provider_protection.test", "type", "default"),
				),
			},
			{
				Config: testAccPhoneProviderProtectionUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_phone_provider_protection.test", "type", "exponential"),
				),
			},
			{
				Config: testAccPhoneProviderProtectionReset,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_phone_provider_protection.test", "type", "default"),
				),
			},
			{
				ResourceName:      "auth0_phone_provider_protection.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
