package user_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAResourceServerWithTwoScopesAndAUser = `
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

resource "auth0_user" "user" {
	depends_on = [ auth0_resource_server_scopes.my_scopes ]

	connection_name = "Username-Password-Authentication"
	user_id         = "{{.testName}}"
	password        = "passpass$12$12"
	email           = "{{.testName}}@acceptance.test.com"
}
`

const testAccUserPermissionsOneAssigned = testAccGivenAResourceServerWithTwoScopesAndAUser + `
resource "auth0_user_permissions" "user_permissions" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "read:foo"
	}
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_permissions.user_permissions ]

	user_id = auth0_user.user.id
}
`

const testAccUserPermissionsTwoAssigned = testAccGivenAResourceServerWithTwoScopesAndAUser + `
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

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_permissions.user_permissions ]

	user_id = auth0_user.user.id
}
`

const testAccUserPermissionsRemoveOnePermission = testAccGivenAResourceServerWithTwoScopesAndAUser + `
resource "auth0_user_permissions" "user_permissions" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id

	permissions  {
		resource_server_identifier = auth0_resource_server.resource_server.identifier
		name                       = "create:foo"
	}
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_permissions.user_permissions ]

	user_id = auth0_user.user.id
}
`

const testAccUserPermissionsDeleteResource = testAccGivenAResourceServerWithTwoScopesAndAUser + `
data "auth0_user" "user_data" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
}
`

const testAccUserPermissionsImportSetup = testAccGivenAResourceServerWithTwoScopesAndAUser + `
resource "auth0_user_permission" "user_permission_1" {
	depends_on = [ auth0_user.user ]

	user_id                    = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "read:foo"
}

resource "auth0_user_permission" "user_permission_2" {
	depends_on = [ auth0_user_permission.user_permission_1 ]

	user_id                    = auth0_user.user.id
	resource_server_identifier = auth0_resource_server.resource_server.identifier
	permission                 = "create:foo"
}
`

const testAccUserPermissionsImportCheck = testAccUserPermissionsImportSetup + `
resource "auth0_user_permissions" "user_permissions" {
	depends_on = [ auth0_user_permission.user_permission_2 ]

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

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_permissions.user_permissions ]

	user_id = auth0_user.user.id
}
`

func TestAccUserPermissions(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserPermissionsOneAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_user_permissions.user_permissions",
						"permissions.*",
						map[string]string{
							"name":                       "read:foo",
							"description":                "Can read Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
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
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionsTwoAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_user_permissions.user_permissions",
						"permissions.*",
						map[string]string{
							"name":                       "read:foo",
							"description":                "Can read Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_user_permissions.user_permissions",
						"permissions.*",
						map[string]string{
							"name":                       "create:foo",
							"description":                "Can create Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
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
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionsRemoveOnePermission, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_user_permissions.user_permissions",
						"permissions.*",
						map[string]string{
							"name":                       "create:foo",
							"description":                "Can create Foo",
							"resource_server_identifier": fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
							"resource_server_name":       fmt.Sprintf("Acceptance Test - %s", testName),
						},
					),
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "permissions.#", "1"),
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
				Config: acctest.ParseTestName(testAccUserPermissionsImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccUserPermissionsImportCheck, testName),
				ResourceName: "auth0_user_permissions.user_permissions",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return acctest.ExtractResourceAttributeFromState(state, "auth0_user.user", "id")
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccUserPermissionsImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "permissions.#", "2"),
					resource.TestCheckResourceAttr("auth0_user_permissions.user_permissions", "permissions.#", "2"),
				),
			},
		},
	})
}
