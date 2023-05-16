package role_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccRolePermissionNoneAssigned = givenARole

const testAccRolePermissionOneAssigned = givenARole + `
resource "auth0_role_permissions" "role_permissions" {
	role_id = auth0_role.role.id
	
	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name = "read:foo"
	}
}
`

const testAccRolePermissionTwoAssigned = givenARole + `
resource "auth0_role_permissions" "role_permissions" {
	role_id = auth0_role.role.id
	
	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name = "create:foo"
	}
	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name = "read:foo"
	}
}
`

func TestAccRolePermissions(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRolePermissionNoneAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionOneAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "read:foo"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.resource_server_identifier", "https://uat.testaccrolepermissions.terraform-provider-auth0.com/api"),

					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.name", "read:foo"),
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.resource_server_identifier", "https://uat.testaccrolepermissions.terraform-provider-auth0.com/api"),
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.resource_server_name", "Acceptance Test - testaccrolepermissions"),
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.description", "Can read Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionTwoAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.1.name", "read:foo"),

					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.1.name", "read:foo"),

					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.resource_server_identifier", "https://uat.testaccrolepermissions.terraform-provider-auth0.com/api"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionOneAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "read:foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionNoneAssigned, strings.ToLower(t.Name())),
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
