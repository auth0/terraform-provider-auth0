package user_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccUserRolesUpdateUserWithOneRoleAssigned = `
resource auth0_role owner {
	name = "owner"
	description = "Owner"
}

resource auth0_user user {
	depends_on = [auth0_role.owner]

	connection_name = "Username-Password-Authentication"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"

	lifecycle {
		ignore_changes = [roles]
	}
}

resource auth0_user_roles user_roles {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	roles = [auth0_role.owner.id]
}

data auth0_user user_data {
	depends_on = [auth0_user_roles.user_roles]

	user_id = auth0_user.user.id
}
`

const testAccUserRolesUpdateUserWithTwoRolesAssigned = `
resource auth0_role owner {
	name = "owner"
	description = "Owner"
}

resource auth0_role admin {
	name = "admin"
	description = "Administrator"
}

resource auth0_user user {
	depends_on = [auth0_role.owner, auth0_role.admin]

	connection_name = "Username-Password-Authentication"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"

	lifecycle {
		ignore_changes = [roles]
	}
}

resource auth0_user_roles user_roles {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	roles = [auth0_role.owner.id, auth0_role.admin.id]
}

data auth0_user user_data {
	depends_on = [auth0_user_roles.user_roles]

	user_id = auth0_user.user.id
}
`

const testAccUserRolesUpdateUserWithNoRolesAssigned = `
resource auth0_user user {
	connection_name = "Username-Password-Authentication"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"

	lifecycle {
		ignore_changes = [roles]
	}
}

resource auth0_user_roles user_roles {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	roles = []
}

data auth0_user user_data {
	depends_on = [auth0_user_roles.user_roles]

	user_id = auth0_user.user.id
}
`

const testAccUserRolesDeleteResource = `
resource auth0_user user {
	connection_name = "Username-Password-Authentication"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "0"),
				),
			},
		},
	})
}

const testAccUserRolesImport = `
resource auth0_role owner {
	name        = "owner"
	description = "Owner"
}

resource auth0_role admin {
	name        = "admin"
	description = "Administrator"
}

resource auth0_user user {
	depends_on = [auth0_role.owner, auth0_role.admin]

	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}@acceptance.test.com"
	password        = "passpass$12$12"

	lifecycle {
		ignore_changes = [ roles, connection_name, password ]
	}
}

resource auth0_user_roles user_roles {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	roles   = [ auth0_role.owner.id, auth0_role.admin.id ]
}
`

func TestAccUserRolesImport(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// The test runs only with recordings as it requires an initial setup.
		t.Skip()
	}

	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             acctest.ParseTestName(testAccUserRolesImport, testName),
				ResourceName:       "auth0_role.owner",
				ImportState:        true,
				ImportStateId:      "rol_XLLMqPwfx8kdG63e",
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccUserRolesImport, testName),
				ResourceName:       "auth0_role.admin",
				ImportState:        true,
				ImportStateId:      "rol_LjLyGzVZE5K34IaY",
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccUserRolesImport, testName),
				ResourceName:       "auth0_user.user",
				ImportState:        true,
				ImportStateId:      "auth0|646cbae1e37ebde3a5ad6fd0",
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccUserRolesImport, testName),
				ResourceName:       "auth0_user_roles.user_roles",
				ImportState:        true,
				ImportStateId:      "auth0|646cbae1e37ebde3a5ad6fd0",
				ImportStatePersist: true,
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Config: acctest.ParseTestName(testAccUserRolesImport, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_roles.user_roles", "roles.#", "2"),
				),
			},
		},
	})
}
