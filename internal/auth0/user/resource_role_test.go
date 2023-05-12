package user_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const updateUserWithOneRoleAssigned = `
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

resource auth0_user_role user_role-1 {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	role_id = auth0_role.owner.id
}

data auth0_user user_data {
	depends_on = [ auth0_user_role.user_role-1 ]

	user_id = auth0_user.user.id
}
`

const updateUserWithTwoRolesAssigned = `
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

resource auth0_user_role user_role-1 {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
	role_id = auth0_role.owner.id
}

resource auth0_user_role user_role-2 {
	depends_on = [ auth0_user_role.user_role-1 ]

	user_id = auth0_user.user.id
	role_id = auth0_role.admin.id
}

data auth0_user user_data {
	depends_on = [ auth0_user_role.user_role-2 ]

	user_id = auth0_user.user.id
}
`

const removeAssignedRolesFromUser = `
resource auth0_user user {
	connection_name = "Username-Password-Authentication"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
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
					resource.TestCheckResourceAttr("auth0_user_role.user_role-1", "role_name", "owner"),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-1", "role_description", "Owner"),
				),
			},
			{
				Config: acctest.ParseTestName(updateUserWithTwoRolesAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.user_data", "roles.#", "2"),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-1", "user_id"),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-1", "role_id"),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-1", "role_name", "owner"),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-1", "role_description", "Owner"),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-2", "user_id"),
					resource.TestCheckResourceAttrSet("auth0_user_role.user_role-2", "role_id"),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-2", "role_name", "admin"),
					resource.TestCheckResourceAttr("auth0_user_role.user_role-2", "role_description", "Administrator"),
				),
			},
			{
				Config: acctest.ParseTestName(removeAssignedRolesFromUser, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "0"),
				),
			},
		},
	})
}
