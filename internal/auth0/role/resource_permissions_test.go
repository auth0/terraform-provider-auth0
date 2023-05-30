package role_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccRolePermissionNoneAssigned = givenAResourceServerAndARole + `
data "auth0_role" "role" {
	role_id = auth0_role.role.id
}
`

const testAccRolePermissionOneAssigned = givenAResourceServerAndARole + `
resource "auth0_role_permissions" "role_permissions" {
	role_id = auth0_role.role.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "read:foo"
	}
}

data "auth0_role" "role" {
	depends_on = [ auth0_role_permissions.role_permissions ]

	role_id = auth0_role.role.id
}
`

const testAccRolePermissionTwoAssigned = givenAResourceServerAndARole + `
resource "auth0_role_permissions" "role_permissions" {
	role_id = auth0_role.role.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "create:foo"
	}

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "read:foo"
	}
}

data "auth0_role" "role" {
	depends_on = [ auth0_role_permissions.role_permissions ]

	role_id = auth0_role.role.id
}
`

func TestAccRolePermissions(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRolePermissionNoneAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionOneAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "read:foo"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.resource_server_identifier", fmt.Sprintf("https://uat.%s.terraform-provider-auth0.com/api", testName)),

					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.name", "read:foo"),
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.resource_server_identifier", fmt.Sprintf("https://uat.%s.terraform-provider-auth0.com/api", testName)),
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.resource_server_name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.description", "Can read Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionTwoAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.1.name", "read:foo"),

					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "create:foo"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.1.name", "read:foo"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.resource_server_identifier", fmt.Sprintf("https://uat.%s.terraform-provider-auth0.com/api", testName)),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.1.resource_server_identifier", fmt.Sprintf("https://uat.%s.terraform-provider-auth0.com/api", testName)),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionOneAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.0.name", "read:foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionNoneAssigned, testName),
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

const testAccRolePermissionsImport = `
resource "auth0_resource_server" "resource_server" {
	name       = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.{{.testName}}.terraform-provider-auth0.com/api"

	scopes {
		value       = "read:foo"
		description = "Can read Foo"
	}

	scopes {
		value       = "create:foo"
		description = "Can create Foo"
	}
}

resource "auth0_role" "role" {
	depends_on = [ auth0_resource_server.resource_server ]

	name        = "Acceptance Test - {{.testName}}"
	description = "Acceptance Test Role - {{.testName}}"

	lifecycle {
		ignore_changes = [ permissions ]
	}
}

resource "auth0_role_permissions" "role_permissions" {
	role_id = auth0_role.role.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "create:foo"
	}

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "read:foo"
	}
}
`

func TestAccRolePermissionsImport(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// The test runs only with recordings as it requires an initial setup.
		t.Skip()
	}

	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             acctest.ParseTestName(testAccRolePermissionsImport, testName),
				ResourceName:       "auth0_resource_server.resource_server",
				ImportState:        true,
				ImportStateId:      "646cbf5e8ffa2057b6112389",
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccRolePermissionsImport, testName),
				ResourceName:       "auth0_role.role",
				ImportState:        true,
				ImportStateId:      "rol_y36rul7VuBQGVpNa",
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccRolePermissionsImport, testName),
				ResourceName:       "auth0_role_permissions.role_permissions",
				ImportState:        true,
				ImportStateId:      "rol_y36rul7VuBQGVpNa",
				ImportStatePersist: true,
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Config: acctest.ParseTestName(testAccRolePermissionsImport, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.#", "2"),
				),
			},
		},
	})
}
