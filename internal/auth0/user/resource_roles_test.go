package user_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenTwoRolesAndAUser = `
resource "auth0_role" "owner" {
	name        = "Test Owner {{.testName}}"
	description = "Owner {{.testName}}"
}

resource "auth0_role" "admin" {
	depends_on = [ auth0_role.owner ]

	name        = "Test Admin {{.testName}}"
	description = "Administrator {{.testName}}"
}

resource "auth0_user" "user" {
	depends_on = [ auth0_role.admin ]

	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}@acceptance.test.com"
	password        = "passpass$12$12"
	username        = "{{.testName}}"
}
`

const testAccUserRolesUpdateUserWithOneRoleAssigned = testAccGivenTwoRolesAndAUser + `
resource "auth0_user_roles" "user_roles" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	roles   = [ auth0_role.owner.id ]
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_roles.user_roles ]

	user_id = auth0_user.user.id
}
`

const testAccUserRolesUpdateUserWithTwoRolesAssigned = testAccGivenTwoRolesAndAUser + `
resource "auth0_user_roles" "user_roles" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	roles   = [ auth0_role.owner.id, auth0_role.admin.id ]
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_roles.user_roles ]

	user_id = auth0_user.user.id
}
`

const testAccUserRolesUpdateUserWithNoRolesAssigned = testAccGivenTwoRolesAndAUser + `
resource "auth0_user_roles" "user_roles" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	roles   = []
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_roles.user_roles ]

	user_id = auth0_user.user.id
}
`

const testAccUserRolesDeleteResource = testAccGivenTwoRolesAndAUser + `
data "auth0_user" "user_data" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
}
`

const testAccUserRolesImportSetup = testAccGivenTwoRolesAndAUser + `
resource "auth0_user_role" "user_role-1" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	role_id = auth0_role.owner.id
}

resource "auth0_user_role" "user_role-2" {
	depends_on = [ auth0_user_role.user_role-1 ]

	user_id = auth0_user.user.id
	role_id = auth0_role.admin.id
}
`

const testAccUserRolesImportCheck = testAccUserRolesImportSetup + `
resource "auth0_user_roles" "user_roles" {
	depends_on = [ auth0_user_role.user_role-2 ]

	user_id = auth0_user.user.id
	roles   = [ auth0_role.owner.id, auth0_role.admin.id ]
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_roles.user_roles ]

	user_id = auth0_user.user.id
}
`

func TestAccUserRoles(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserRolesUpdateUserWithOneRoleAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "1"),
					resource.TestCheckResourceAttr("auth0_user_roles.user_roles", "roles.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserRolesUpdateUserWithTwoRolesAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "2"),
					resource.TestCheckResourceAttr("auth0_user_roles.user_roles", "roles.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserRolesUpdateUserWithNoRolesAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "0"),
					resource.TestCheckResourceAttr("auth0_user_roles.user_roles", "roles.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserRolesUpdateUserWithOneRoleAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "1"),
					resource.TestCheckResourceAttr("auth0_user_roles.user_roles", "roles.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserRolesDeleteResource, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserRolesImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccUserRolesImportCheck, testName),
				ResourceName: "auth0_user_roles.user_roles",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return acctest.ExtractResourceAttributeFromState(state, "auth0_user.user", "id")
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccUserRolesImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "2"),
					resource.TestCheckResourceAttr("auth0_user_roles.user_roles", "roles.#", "2"),
				),
			},
		},
	})
}
