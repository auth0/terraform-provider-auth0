package user_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccUserPermissionsNoneAssigned = givenAResourceServerAndUser

const testAccUserPermissionsOneAssigned = givenAResourceServerAndUser + `
resource "auth0_user_permissions" "user_permissions" {
	depends_on = [ auth0_resource_server.resource_server, auth0_user.user ]

	user_id = auth0_user.user.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name = "read:foo"
	}
}
`

const testAccUserPermissionsTwoAssigned = givenAResourceServerAndUser + `
resource "auth0_user_permissions" "user_permissions" {
	depends_on = [ auth0_resource_server.resource_server, auth0_user.user ]

	user_id = auth0_user.user.id

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

func TestAccUserPermissions(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserPermissionsNoneAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionsOneAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "1"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.name", "read:foo"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_identifier", "https://uat.api.terraform-provider-auth0.com/testaccuserpermissions"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_name", "Acceptance Test - testaccuserpermissions"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.description", "Can read Foo"),

					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.0.name", "read:foo"),
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.0.resource_server_identifier", "https://uat.api.terraform-provider-auth0.com/testaccuserpermissions"),
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.0.resource_server_name", "Acceptance Test - testaccuserpermissions"),
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.0.description", "Can read Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionsTwoAssigned, strings.ToLower(t.Name())),
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
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_identifier", "https://uat.api.terraform-provider-auth0.com/testaccuserpermissions"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_name", "Acceptance Test - testaccuserpermissions"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.description", "Can create Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionsOneAssigned, strings.ToLower(t.Name())),
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
				Config: acctest.ParseTestName(testAccUserPermissionsNoneAssigned, strings.ToLower(t.Name())),
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
