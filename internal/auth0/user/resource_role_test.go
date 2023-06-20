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

const updateUserWithOneRoleAssigned = testAccGivenTwoRolesAndAUser + `
resource "auth0_user_role" "user_role-1" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	role_id = auth0_role.owner.id
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_role.user_role-1 ]

	user_id = auth0_user.user.id
}
`

const updateUserWithTwoRolesAssigned = testAccGivenTwoRolesAndAUser + `
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

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_role.user_role-2 ]

	user_id = auth0_user.user.id
}
`

const testAccUserRoleImportSetup = testAccGivenTwoRolesAndAUser + `
resource "auth0_user_roles" "user_roles" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	roles   = [ auth0_role.owner.id, auth0_role.admin.id ]
}
`

const testAccUserRoleImportCheck = testAccUserRoleImportSetup + `
resource "auth0_user_role" "user_role-1" {
	depends_on = [ auth0_user_roles.user_roles ]

	user_id = auth0_user.user.id
	role_id = auth0_role.owner.id
}

resource "auth0_user_role" "user_role-2" {
	depends_on = [ auth0_user_role.user_role-1 ]

	user_id = auth0_user.user.id
	role_id = auth0_role.admin.id
}

data "auth0_user" "user_data" {
	depends_on = [ auth0_user_role.user_role-2 ]

	user_id = auth0_user.user.id
}
`

func TestAccUserRole(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(updateUserWithOneRoleAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "1"),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-1", "user_id"),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-1", "role_id"),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-1", "role_name", fmt.Sprintf("Test Owner %s", testName)),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-1", "role_description", fmt.Sprintf("Owner %s", testName)),
				),
			},
			{
				Config: acctest.ParseTestName(updateUserWithTwoRolesAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "2"),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-1", "user_id"),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-1", "role_id"),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-1", "role_name", fmt.Sprintf("Test Owner %s", testName)),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-1", "role_description", fmt.Sprintf("Owner %s", testName)),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-2", "user_id"),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-2", "role_id"),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-2", "role_name", fmt.Sprintf("Test Admin %s", testName)),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-2", "role_description", fmt.Sprintf("Administrator %s", testName)),
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
				Config: acctest.ParseTestName(testAccUserRoleImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccUserRoleImportCheck, testName),
				ResourceName: "auth0_user_role.user_role-1",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					userID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_user.user", "id")
					assert.NoError(t, err)

					roleID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_role.owner", "id")
					assert.NoError(t, err)

					return userID + "::" + roleID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config:       acctest.ParseTestName(testAccUserRoleImportCheck, testName),
				ResourceName: "auth0_user_role.user_role-2",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					userID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_user.user", "id")
					assert.NoError(t, err)

					roleID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_role.admin", "id")
					assert.NoError(t, err)

					return userID + "::" + roleID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccUserRoleImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "2"),
				),
			},
		},
	})
}
