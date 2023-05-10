package user_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const givenAResourceServerWithPermissions = `
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
`

const givenAUserPermission = `
resource "auth0_user_permission" "user_permission_read" {
	depends_on = [ auth0_resource_server.resource_server, auth0_user.user ]

	user_id = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission = "read:foo"
}
`

const givenAnotherUserPermission = `
resource "auth0_user_permission" "user_permission_create" {
	depends_on = [ auth0_resource_server.resource_server, auth0_user.user ]

	user_id = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission = "create:foo"
}
`

const testAccUserPermissionNoneAssigned = givenAResourceServerWithPermissions + testAccUserEmpty
const testAccUserPermissionOneAssigned = givenAResourceServerWithPermissions + testAccUserEmpty + givenAUserPermission
const testAccUserPermissionTwoAssigned = givenAResourceServerWithPermissions + testAccUserEmpty + givenAUserPermission + givenAnotherUserPermission

func TestAccUserPermission(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserPermissionNoneAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionOneAssigned, strings.ToLower(t.Name())),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionOneAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "1"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.name", "read:foo"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_identifier", "https://uat.api.terraform-provider-auth0.com/testaccuserpermission"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.resource_server_name", "Acceptance Test - testaccuserpermission"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.description", "Can read Foo"),

					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "permission", "read:foo"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "user_id", "auth0|testaccuserpermission"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "resource_server_identifier", "https://uat.api.terraform-provider-auth0.com/testaccuserpermission"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "resource_server_name", "Acceptance Test - testaccuserpermission"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "description", "Can read Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionTwoAssigned, strings.ToLower(t.Name())),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionTwoAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "2"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.1.name", "read:foo"),

					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "permission", "create:foo"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "user_id", "auth0|testaccuserpermission"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "resource_server_identifier", "https://uat.api.terraform-provider-auth0.com/testaccuserpermission"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "resource_server_name", "Acceptance Test - testaccuserpermission"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "description", "Can create Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionOneAssigned, strings.ToLower(t.Name())),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionOneAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "1"),
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.0.name", "read:foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionNoneAssigned, strings.ToLower(t.Name())),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionNoneAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "permissions.#", "0"),
				),
			},
		},
	})
}
