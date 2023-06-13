package role_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAResourceServerWithTwoScopesAndARole = `
resource "auth0_resource_server" "resource_server" {
	name       = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"

	lifecycle {
		ignore_changes = [ scopes ]
	}
}

resource "auth0_resource_server_scopes" "my_scopes" {
	depends_on = [ auth0_resource_server.resource_server ]

	resource_server_identifier = auth0_resource_server.resource_server.identifier

	scopes {
		name        = "read:foo"
		description = "Can read Foo"
	}

	scopes {
		name        = "create:foo"
		description = "Can create Foo"
	}
}

resource "auth0_role" "role" {
	depends_on = [ auth0_resource_server_scopes.my_scopes ]

	name        = "Acceptance Test - {{.testName}}"
	description = "Acceptance Test Role - {{.testName}}"

	lifecycle {
		ignore_changes = [ permissions ]
	}
}
`

const testAccRolePermissionOneAssigned = testAccGivenAResourceServerWithTwoScopesAndARole + `
resource "auth0_role_permissions" "role_permissions" {
	depends_on = [ auth0_role.role ]

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

const testAccRolePermissionTwoAssigned = testAccGivenAResourceServerWithTwoScopesAndARole + `
resource "auth0_role_permissions" "role_permissions" {
	depends_on = [ auth0_role.role ]

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

const testAccRolePermissionsRemoveOnePermission = testAccGivenAResourceServerWithTwoScopesAndARole + `
resource "auth0_role_permissions" "role_permissions" {
	depends_on = [ auth0_role.role ]

	role_id = auth0_role.role.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "create:foo"
	}
}

data "auth0_role" "role" {
	depends_on = [ auth0_role_permissions.role_permissions ]

	role_id = auth0_role.role.id
}
`

const testAccRolePermissionsDeleteResource = testAccGivenAResourceServerWithTwoScopesAndARole + `
data "auth0_role" "role" {
	depends_on = [ auth0_role.role ]

	role_id = auth0_role.role.id
}
`

const testAccRolePermissionsImportSetup = testAccGivenAResourceServerWithTwoScopesAndARole + `
resource "auth0_role_permission" "role_permission_1" {
	depends_on = [ auth0_role.role ]

	role_id                    = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "read:foo"
}

resource "auth0_role_permission" "role_permission_2" {
	depends_on = [ auth0_role_permission.role_permission_1 ]

	role_id                    = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "create:foo"
}
`

const testAccRolePermissionsImportCheck = testAccRolePermissionsImportSetup + `
resource "auth0_role_permissions" "role_permissions" {
	depends_on = [ auth0_role_permission.role_permission_2 ]

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

const testAccRolePermissionsRemoveOnePermission = testAccGivenAResourceServerWithTwoScopesAndARole + `
resource "auth0_role_permissions" "role_permissions" {
	role_id = auth0_role.role.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "create:foo"
	}
}

data "auth0_role" "role" {
	depends_on = [ auth0_role_permissions.role_permissions ]

	role_id = auth0_role.role.id
}
`

const testAccRolePermissionsDeleteResource = testAccGivenAResourceServerWithTwoScopesAndARole + `
data "auth0_role" "role" {
	depends_on = [ auth0_role.role ]

	role_id = auth0_role.role.id
}
`

const testAccRolePermissionsImportSetup = testAccGivenAResourceServerWithTwoScopesAndARole + `
resource "auth0_role_permission" "role_permission_1" {
	depends_on = [ auth0_role.role ]

	role_id                    = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "read:foo"
}

resource "auth0_role_permission" "role_permission_2" {
	depends_on = [ auth0_role_permission.role_permission_1 ]

	role_id                    = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "create:foo"
}
`

const testAccRolePermissionsImportCheck = testAccRolePermissionsImportSetup + `
resource "auth0_role_permissions" "role_permissions" {
	depends_on = [ auth0_role_permission.role_permission_2 ]

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
				Config: acctest.ParseTestName(testAccRolePermissionOneAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_role_permissions.role_permissions",
						"permissions.*",
						map[string]string{
							"name":                       "read:foo",
							"description":                "Can read Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_role.role",
						"permissions.*",
						map[string]string{
							"name":                       "read:foo",
							"description":                "Can read Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionTwoAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_role_permissions.role_permissions",
						"permissions.*",
						map[string]string{
							"name":                       "read:foo",
							"description":                "Can read Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_role_permissions.role_permissions",
						"permissions.*",
						map[string]string{
							"name":                       "create:foo",
							"description":                "Can create Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_role.role",
						"permissions.*",
						map[string]string{
							"name":                       "read:foo",
							"description":                "Can read Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_role.role",
						"permissions.*",
						map[string]string{
							"name":                       "create:foo",
							"description":                "Can create Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsRemoveOnePermission, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_role_permissions.role_permissions",
						"permissions.*",
						map[string]string{
							"name":                       "create:foo",
							"description":                "Can create Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_role.role",
						"permissions.*",
						map[string]string{
							"name":                       "create:foo",
							"description":                "Can create Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsDeleteResource, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccRolePermissionsImportCheck, testName),
				ResourceName: "auth0_role_permissions.role_permissions",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return acctest.ExtractResourceAttributeFromState(state, "auth0_role.role", "id")
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionsImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "2"),
					resource.TestCheckResourceAttr("auth0_role_permissions.role_permissions", "permissions.#", "2"),
				),
			},
		},
	})
}
