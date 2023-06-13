package user_test

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

const testAccUserPermissionWithOnePermissionAssigned = testAccGivenAResourceServerWithTwoScopesAndAUser + `
resource "auth0_user_permission" "user_permission_read" {
	depends_on = [ auth0_user.user ]

	user_id                    = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "read:foo"
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_permission.user_permission_read ]

	user_id = auth0_user.user.id
}
`

const testAccUserPermissionWithTwoPermissionsAssigned = testAccGivenAResourceServerWithTwoScopesAndAUser + `
resource "auth0_user_permission" "user_permission_read" {
	depends_on = [ auth0_user.user ]

	user_id                    = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "read:foo"
}

resource "auth0_user_permission" "user_permission_create" {
	depends_on = [ auth0_user_permission.user_permission_read ]

	user_id                    = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "create:foo"
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_permission.user_permission_create ]

	user_id = auth0_user.user.id
}
`

const testAccUserPermissionImportSetup = testAccGivenAResourceServerWithTwoScopesAndAUser + `
resource "auth0_user_permissions" "user_permissions" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "read:foo"
	}

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "create:foo"
	}
}
`

const testAccUserPermissionImportCheck = testAccUserPermissionImportSetup + `
resource "auth0_user_permission" "user_permission_read" {
	depends_on = [ auth0_user.user ]

	user_id                    = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "read:foo"
}

resource "auth0_user_permission" "user_permission_create" {
	depends_on = [ auth0_user_permission.user_permission_read ]

	user_id                    = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "create:foo"
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_permission.user_permission_create ]

	user_id = auth0_user.user.id
}
`

func TestAccUserPermission(t *testing.T) {
	testName := strings.ToLower(t.Name())
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserPermissionWithOnePermissionAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "permissions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_user.user_data",
						"permissions.*",
						map[string]string{
							"name":                       "read:foo",
							"description":                "Can read Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "permission", "read:foo"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "user_id", fmt.Sprintf("auth0|%s", testName)),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "resource_server_name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "description", "Can read Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionWithTwoPermissionsAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "permissions.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_user.user_data",
						"permissions.*",
						map[string]string{
							"name":                       "read:foo",
							"description":                "Can read Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_user.user_data",
						"permissions.*",
						map[string]string{
							"name":                       "create:foo",
							"description":                "Can create Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "permission", "read:foo"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "user_id", fmt.Sprintf("auth0|%s", testName)),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "resource_server_name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_read", "description", "Can read Foo"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "permission", "create:foo"),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "user_id", fmt.Sprintf("auth0|%s", testName)),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "resource_server_name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_user_permission.user_permission_create", "description", "Can create Foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionsDeleteResource, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "permissions.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccUserPermissionImportCheck, testName),
				ResourceName: "auth0_user_permission.user_permission_read",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					userID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_user.user", "id")
					assert.NoError(t, err)

					apiID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_resource_server.resource_server", "identifier")
					assert.NoError(t, err)

					return userID + "::" + apiID + "::" + "read:foo", nil
				},
				ImportStatePersist: true,
			},
			{
				Config:       acctest.ParseTestName(testAccUserPermissionImportCheck, testName),
				ResourceName: "auth0_user_permission.user_permission_create",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					userID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_user.user", "id")
					assert.NoError(t, err)

					apiID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_resource_server.resource_server", "identifier")
					assert.NoError(t, err)

					return userID + "::" + apiID + "::" + "create:foo", nil
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "permissions.#", "2"),
				),
			},
		},
	})
}
