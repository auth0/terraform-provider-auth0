package role_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const givenAResourceServerAndARole = `
resource "auth0_resource_server" "resource_server" {
	name = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.{{.testName}}.terraform-provider-auth0.com/api"
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
	depends_on = [ auth0_resource_server.resource_server ]

	name = "Acceptance Test - {{.testName}}"
	description = "Acceptance Test Role - {{.testName}}"

	lifecycle {
		ignore_changes = [ permissions ]
	}
}
`

const testAccRolePermissionsNoneAssigned = givenAResourceServerAndARole + `data "auth0_role" "role" {
	role_id = auth0_role.role.id
}`

const givenAResourceServerAndARoleAndAPermission = givenAResourceServerAndARole + `
resource "auth0_role_permission" "role_permission" {
	role_id = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission = "create:foo"
}
`

const testAccRolePermissionsOneAssigned = givenAResourceServerAndARoleAndAPermission + `
data "auth0_role" "role" {
	depends_on = [ auth0_role_permission.role_permission ]
	role_id = auth0_role.role.id
}
`

const testAccRolePermissionsTwoAssigned = givenAResourceServerAndARoleAndAPermission + `
resource "auth0_role_permission" "another_role_permission" {
	depends_on = [ auth0_role_permission.another_role_permission ]

	role_id = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission = "read:foo"
}

data "auth0_role" "role" {
	depends_on = [ auth0_role_permission.role_permission, auth0_role_permission.another_role_permission ]
	role_id = auth0_role.role.id
}
`

func TestAccRolePermission(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRolePermissionsNoneAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsOneAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.resource_server_identifier", "https://uat.testaccrolepermission.terraform-provider-auth0.com/api"),

					resource.TestCheckResourceAttr("auth0_role_permission.role_permission", "permission", "create:foo"),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission", "resource_server_identifier", "https://uat.testaccrolepermission.terraform-provider-auth0.com/api"),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission", "resource_server_name", "Acceptance Test - testaccrolepermission"),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission", "description", "Can create Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsTwoAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission", "permission", "create:foo"),
					resource.TestCheckResourceAttr("auth0_role_permission.another_role_permission", "permission", "read:foo"),

					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.1.name", "read:foo"),

					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.resource_server_identifier", "https://uat.testaccrolepermission.terraform-provider-auth0.com/api"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsOneAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "create:foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsNoneAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "0"),
				),
			},
		},
	})
}
