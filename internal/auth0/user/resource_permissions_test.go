package user_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

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

const testAccUserPermissionsImport = `
resource "auth0_resource_server" "resource_server" {
	name       = "Acceptance Test - {{.testName}}"
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

resource "auth0_user" "user" {
	depends_on = [ auth0_resource_server.resource_server ]

	connection_name = "Username-Password-Authentication"
	password        = "passpass$12$12"
	email           = "{{.testName}}@acceptance.test.com"

	lifecycle {
		ignore_changes = [ connection_name, password ]
	}
}

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

func TestAccUserPermissionsImport(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// The test runs only with recordings as it requires an initial setup.
		t.Skip()
	}

	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             acctest.ParseTestName(testAccUserPermissionsImport, testName),
				ResourceName:       "auth0_resource_server.resource_server",
				ImportState:        true,
				ImportStateId:      "646cad063390a55e156ee4cd",
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccUserPermissionsImport, testName),
				ResourceName:       "auth0_user.user",
				ImportState:        true,
				ImportStateId:      "auth0|646cad06363aac78e30fc478",
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccUserPermissionsImport, testName),
				ResourceName:       "auth0_user_permissions.user_permissions",
				ImportState:        true,
				ImportStateId:      "auth0|646cad06363aac78e30fc478",
				ImportStatePersist: true,
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Config: acctest.ParseTestName(testAccUserPermissionsImport, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.#", "2"),
				),
			},
		},
	})
}
