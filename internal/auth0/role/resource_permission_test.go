package role_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccRolePermissionWithOnePermissionAssigned = testAccGivenAResourceServerWithTwoScopesAndARole + `
resource "auth0_role_permission" "role_permission_create" {
	depends_on = [ auth0_role.role ]

	role_id                    = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "create:foo"
}

data "auth0_role" "role" {
	depends_on = [ auth0_role_permission.role_permission_create ]

	role_id = auth0_role.role.id
}
`

const testAccRolePermissionWithTwoPermissionAssigned = testAccGivenAResourceServerWithTwoScopesAndARole + `
resource "auth0_role_permission" "role_permission_create" {
	depends_on = [ auth0_role.role ]

	role_id                    = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "create:foo"
}

resource "auth0_role_permission" "role_permission_read" {
	depends_on = [ auth0_role_permission.role_permission_create ]

	role_id                    = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "read:foo"
}

data "auth0_role" "role" {
	depends_on = [ auth0_role_permission.role_permission_read ]

	role_id = auth0_role.role.id
}
`

const testAccRolePermissionImportSetup = testAccGivenAResourceServerWithTwoScopesAndARole + `
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
`

const testAccRolePermissionImportCheck = testAccRolePermissionImportSetup + `
resource "auth0_role_permission" "role_permission_create" {
	depends_on = [ auth0_role_permissions.role_permissions ]

	role_id                    = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "create:foo"
}

resource "auth0_role_permission" "role_permission_read" {
	depends_on = [ auth0_role_permission.role_permission_create ]

	role_id                    = auth0_role.role.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "read:foo"
}

data "auth0_role" "role" {
	depends_on = [ auth0_role_permission.role_permission_read ]

	role_id = auth0_role.role.id
}
`

func TestAccRolePermission(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRolePermissionWithOnePermissionAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
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
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_create", "permission", "create:foo"),
					resource.TestCheckResourceAttrSet("auth0_role_permission.role_permission_create", "role_id"),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_create", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_create", "resource_server_name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_create", "description", "Can create Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionWithTwoPermissionAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
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
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_read", "permission", "read:foo"),
					resource.TestCheckResourceAttrSet("auth0_role_permission.role_permission_read", "role_id"),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_read", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_read", "resource_server_name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_read", "description", "Can read Foo"),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_create", "permission", "create:foo"),
					resource.TestCheckResourceAttrSet("auth0_role_permission.role_permission_create", "role_id"),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_create", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_create", "resource_server_name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_role_permission.role_permission_create", "description", "Can create Foo"),
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
				Config: acctest.ParseTestName(testAccRolePermissionImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccRolePermissionImportCheck, testName),
				ResourceName: "auth0_role_permission.role_permission_create",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					roleID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_role.role", "id")
					assert.NoError(t, err)

					apiID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_resource_server.resource_server", "identifier")
					assert.NoError(t, err)

					return roleID + "::" + apiID + "::" + "create:foo", nil
				},
				ImportStatePersist: true,
			},
			{
				Config:       acctest.ParseTestName(testAccRolePermissionImportCheck, testName),
				ResourceName: "auth0_role_permission.role_permission_read",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					roleID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_role.role", "id")
					assert.NoError(t, err)

					apiID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_resource_server.resource_server", "identifier")
					assert.NoError(t, err)

					return roleID + "::" + apiID + "::" + "read:foo", nil
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccRolePermissionImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_role.role", "permissions.#", "2"),
				),
			},
		},
	})
}
