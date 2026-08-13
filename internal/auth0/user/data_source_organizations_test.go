package user_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceUserOrganizationsConfig = `
resource "auth0_user" "my_user" {
	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}@auth0.com"
	password        = "MyPass123$"
}

resource "auth0_organization" "my_organization" {
	depends_on = [ auth0_user.my_user ]

	name                      = "test-org-{{.testName}}"
	display_name              = "Test Org {{.testName}}"
	is_app_entitlement_active = true
}

resource "auth0_organization_member" "my_member" {
	depends_on = [ auth0_organization.my_organization ]

	organization_id = auth0_organization.my_organization.id
	user_id         = auth0_user.my_user.id
}

data "auth0_user_organizations" "my_user_organizations" {
	depends_on = [ auth0_organization_member.my_member ]
	user_id    = auth0_user.my_user.id
}
`

func TestAccDataSourceUserOrganizations(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceUserOrganizationsConfig, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user_organizations.my_user_organizations", "organizations.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_user_organizations.my_user_organizations", "organizations.0.organization_id"),
					resource.TestCheckResourceAttrSet("data.auth0_user_organizations.my_user_organizations", "organizations.0.name"),
					resource.TestCheckResourceAttr("data.auth0_user_organizations.my_user_organizations", "organizations.0.is_app_entitlement_active", "true"),
					resource.TestCheckResourceAttr("data.auth0_user_organizations.my_user_organizations", "organizations.0.third_party_client_access", "block"),
				),
			},
		},
	})
}
