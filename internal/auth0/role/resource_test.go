package role_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAResourceServerWithScopes = `
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
`

const testAccRoleEmpty = `
resource "auth0_role" "the_one" {
	name = "The One - Acceptance Test - {{.testName}}"
}
`

const testAccRoleCreate = testAccGivenAResourceServerWithScopes + `
resource "auth0_role" "the_one" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	name        = "The One - Acceptance Test - {{.testName}}"
	description = "The One - Acceptance Test"

	permissions {
		name                       = "stop:bullets"
		resource_server_identifier = auth0_resource_server.matrix.identifier
	}
}
`

const testAccRoleUpdate = testAccGivenAResourceServerWithScopes + `
resource "auth0_role" "the_one" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	name        = "The One - Acceptance Test - {{.testName}}"
	description = "The One who will bring peace - Acceptance Test"

	permissions {
		name                       = "stop:bullets"
		resource_server_identifier = auth0_resource_server.matrix.identifier
	}

	permissions {
		name                       = "bring:peace"
		resource_server_identifier = auth0_resource_server.matrix.identifier
	}
}
`

const testAccRoleEmptyAgain = `
resource "auth0_role" "the_one" {
	name        = "The One - Acceptance Test - {{.testName}}"
	description = " "
}
`

func TestAccRole(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRoleEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "name", fmt.Sprintf("The One - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", ""),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRoleCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "name", fmt.Sprintf("The One - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", "The One - Acceptance Test"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.#", "1"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.name", "stop:bullets"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.description", "Stop bullets"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.resource_server_identifier", fmt.Sprintf("https://%s.matrix.com/", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.resource_server_name", fmt.Sprintf("Role - Acceptance Test - %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRoleUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", "The One who will bring peace - Acceptance Test"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.#", "2"),

					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.name", "bring:peace"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.description", "Bring peace"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.resource_server_identifier", fmt.Sprintf("https://%s.matrix.com/", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.resource_server_name", fmt.Sprintf("Role - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.1.name", "stop:bullets"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.1.description", "Stop bullets"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.1.resource_server_identifier", fmt.Sprintf("https://%s.matrix.com/", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.1.resource_server_name", fmt.Sprintf("Role - Acceptance Test - %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRoleCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "name", fmt.Sprintf("The One - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", "The One - Acceptance Test"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.#", "1"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.name", "stop:bullets"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.description", "Stop bullets"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.resource_server_identifier", fmt.Sprintf("https://%s.matrix.com/", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.0.resource_server_name", fmt.Sprintf("Role - Acceptance Test - %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRoleEmptyAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "name", fmt.Sprintf("The One - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", " "), // Management API ignores empty strings for role descriptions.
					resource.TestCheckResourceAttr("auth0_role.the_one", "permissions.#", "0"),
				),
			},
		},
	})
}
