package role_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const givenAResourceServerAndRole = `
resource "auth0_resource_server" "resource_server" {
	name = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	scopes {
		value = "read:foo"
		description = "Can read Foo"
	}
	scopes {
		value = "create:foo"
		description = "Can create Foo"
	}
}

resource "auth0_role" "role" { 

}
`

const testAccRolePermissionsNoneAssigned = givenAResourceServerAndRole

const testAccRolePermissionsOneAssigned = givenAResourceServerAndRole + `
resource "auth0_user_permissions" "user_permissions" {
	depends_on = [ auth0_resource_server.resource_server, auth0_user.user ]

	role_id = auth0_user.role.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name = "read:foo"
	}
}
`

const testAccRolePermissionsTwoAssigned = givenAResourceServerAndRole + `
resource "auth0_user_permissions" "user_permissions" {
	depends_on = [ auth0_resource_server.resource_server, auth0_user.user ]

	role_id = auth0_user.role.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name = "read:foo"
	}
	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name = "create:foo"
	}
}
`

func TestAccRolePermission(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRolePermissionsNoneAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.role", "permissions.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsOneAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "1"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.name", "read:foo"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_identifier", "https://uat.api.terraform-provider-auth0.com/testaccRolepermissions"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_name", "Acceptance Test - testaccRolepermissions"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.description", "Can read Foo"),

					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.0.name", "read:foo"),
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.0.resource_server_identifier", "https://uat.api.terraform-provider-auth0.com/testaccRolepermissions"),
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.0.resource_server_name", "Acceptance Test - testaccRolepermissions"),
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.0.description", "Can read Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsTwoAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.#", "2"),
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.1.name", "read:foo"),

					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "2"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.1.name", "read:foo"),

					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_identifier", "https://uat.api.terraform-provider-auth0.com/testaccRolepermissions"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_name", "Acceptance Test - testaccRolepermissions"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.description", "Can create Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsOneAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.#", "1"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "1"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.name", "read:foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsNoneAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "0"),
				),
			},
		},
	})
}
