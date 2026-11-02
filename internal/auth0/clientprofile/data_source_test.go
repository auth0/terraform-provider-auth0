package clientprofile_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAClientProfile = `
resource "auth0_client_profile" "my_profile" {
  name        = "Acceptance Test Profile"
  description = "Created by the acceptance tests."

  client_configuration {
    refresh_token {
      token_lifetime {
        default_value = 2592000
      }
      idle_token_lifetime {
        default_value = 1296000
      }
      leeway {
        default_value = 0
      }
    }
    organization_usage {
      default_value = "allow"
    }
  }
}
`

const testAccDataSourceClientProfileWithoutID = testAccGivenAClientProfile + `
data "auth0_client_profile" "my_profile" {
  depends_on = [auth0_client_profile.my_profile]
}
`

const testAccDataSourceClientProfileWithInvalidID = testAccGivenAClientProfile + `
data "auth0_client_profile" "my_profile" {
  depends_on = [auth0_client_profile.my_profile]
  id         = "cpr_aaaaaaaaaaaaaaaaaaaaaa"
}
`

const testAccDataSourceClientProfile = testAccGivenAClientProfile + `
data "auth0_client_profile" "my_profile" {
  id = auth0_client_profile.my_profile.id
}
`

func TestAccDataSourceClientProfile(t *testing.T) {
	const name = "data.auth0_client_profile.my_profile"

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceClientProfileWithoutID,
				ExpectError: regexp.MustCompile(`The argument "id" is required`),
			},
			{
				Config:      testAccDataSourceClientProfileWithInvalidID,
				ExpectError: regexp.MustCompile(`404|[Nn]ot [Ff]ound`),
			},
			{
				Config: testAccDataSourceClientProfile,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(name, "id", "auth0_client_profile.my_profile", "id"),
					resource.TestCheckResourceAttr(name, "name", "Acceptance Test Profile"),
					resource.TestCheckResourceAttr(name, "description", "Created by the acceptance tests."),
					resource.TestCheckResourceAttr(name, "apply_policy.0.external_metadata_type", "cimd"),
					resource.TestCheckResourceAttr(name, "apply_policy.0.external_metadata_created_by", "client"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.token_lifetime.0.default_value", "2592000"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.idle_token_lifetime.0.default_value", "1296000"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.leeway.0.default_value", "0"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.organization_usage.0.default_value", "allow"),
					resource.TestCheckResourceAttrSet(name, "created_at"),
					resource.TestCheckResourceAttrSet(name, "updated_at"),
				),
			},
		},
	})
}
