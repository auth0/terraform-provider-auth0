package organization_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceOrganizationRoleGroups = testAccGivenAnOrganizationRole + `
data "auth0_organization_role_groups" "test" {
	depends_on = [ auth0_role.operator ]

	organization_id = auth0_organization.org.id
	role_id         = auth0_role.operator.id
}
`

func TestAccDataSourceOrganizationRoleGroups(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceOrganizationRoleGroups, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_organization_role_groups.test", "organization_id"),
					resource.TestCheckResourceAttrPair(
						"data.auth0_organization_role_groups.test", "role_id",
						"auth0_role.operator", "id",
					),
					// Groups come from SCIM, so there is no way to create one here.
					resource.TestCheckResourceAttr("data.auth0_organization_role_groups.test", "groups.#", "0"),
				),
			},
		},
	})
}
