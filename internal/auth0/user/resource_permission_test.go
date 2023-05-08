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

const testAccUserPermission = `
resource "auth0_user_permission" "user_permission" {
	depends_on = [auth0_resource_server.resource_server, auth0_user.user ]
	user_id = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission = "read:foo"
}
`

func TestAccUserPermission(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserEmpty+givenAResourceServerWithPermissions+testAccUserPermission, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission", "permission", "read:foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserEmpty+givenAResourceServerWithPermissions, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission", "permission", "read:foo"),
				),
			},
		},
	})
}
