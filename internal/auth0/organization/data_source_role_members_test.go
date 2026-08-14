package organization_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAnOrganizationRole = `
resource "auth0_organization" "org" {
	name         = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}

resource "auth0_role" "operator" {
	depends_on = [ auth0_organization.org ]

	name     = "Test Operator - {{.testName}}"
	type     = "organization"
	owner_id = auth0_organization.org.id
}
`

const testAccDataSourceOrganizationRoleMembers = testAccGivenAnOrganizationRole + `
data "auth0_organization_role_members" "test" {
	depends_on = [ auth0_role.operator ]

	organization_id = auth0_organization.org.id
	role_id         = auth0_role.operator.id
}
`

func TestAccDataSourceOrganizationRoleMembers(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceOrganizationRoleMembers, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_organization_role_members.test", "organization_id"),
					resource.TestCheckResourceAttrPair(
						"data.auth0_organization_role_members.test", "role_id",
						"auth0_role.operator", "id",
					),
					// Nothing has been assigned this role yet.
					resource.TestCheckResourceAttr("data.auth0_organization_role_members.test", "members.#", "0"),
				),
			},
		},
	})
}
