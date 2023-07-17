package role_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAResourceServer = `
resource "auth0_resource_server" "matrix" {
    name       = "Role - Acceptance Test - {{.testName}}"
    identifier = "https://{{.testName}}.matrix.com/"
}

resource "auth0_resource_server_scopes" "my_api_scopes" {
	depends_on = [ auth0_resource_server.matrix ]

	resource_server_identifier = auth0_resource_server.matrix.identifier

	scopes {
		name        = "stop:bullets"
		description = "Stop bullets"
	}

	scopes {
		name        = "bring:peace"
		description = "Bring peace"
	}
}

resource "auth0_role" "the_one" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	name        = "The One - Acceptance Test - {{.testName}}"
	description = "The One - Acceptance Test"
}

resource "auth0_role_permissions" "role_permissions" {
	depends_on = [ auth0_role.the_one ]

	role_id = auth0_role.the_one.id

	permissions  {
		resource_server_identifier = auth0_resource_server.matrix.identifier
		name                       = "stop:bullets"
	}

	permissions  {
		resource_server_identifier = auth0_resource_server.matrix.identifier
		name                       = "bring:peace"
	}
}
`

const testAccDataSourceNonExistentRole = `
data "auth0_role" "test" {
	name = "this-role-does-not-exist"
}
`

const testAccDataSourceRoleByName = testAccGivenAResourceServer + `
data "auth0_role" "test" {
	depends_on = [ auth0_role_permissions.role_permissions ]

	name = auth0_role.the_one.name
}
`

const testAccDataSourceRoleByID = testAccGivenAResourceServer + `
data "auth0_role" "test" {
	depends_on = [ auth0_role_permissions.role_permissions ]

	role_id = auth0_role.the_one.id
}
`

func TestAccDataSourceRoleRequiredArguments(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_role" "test" { }`,
				ExpectError: regexp.MustCompile("one of `name,role_id` must be specified"),
			},
		},
	})
}

func TestAccDataSourceRole(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceNonExistentRole, testName),
				ExpectError: regexp.MustCompile(
					`No role found with "name" = "this-role-does-not-exist"`,
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceRoleByName, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.auth0_role.test", "role_id"),
					resource.TestCheckResourceAttr("data.auth0_role.test", "name", fmt.Sprintf("The One - Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("data.auth0_role.test", "description", "The One - Acceptance Test"),
					resource.TestCheckResourceAttr("data.auth0_role.test", "permissions.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceRoleByID, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_role.test", "role_id"),
					resource.TestCheckResourceAttr("data.auth0_role.test", "name", fmt.Sprintf("The One - Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("data.auth0_role.test", "description", "The One - Acceptance Test"),
					resource.TestCheckResourceAttr("data.auth0_role.test", "permissions.#", "2"),
				),
			},
		},
	})
}
