package organization_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember = `
resource "auth0_role" "reader" {
	name = "Test Reader - {{.testName}}"
}

resource "auth0_role" "writer" {
	depends_on = [ auth0_role.reader ]

	name = "Test Writer - {{.testName}}"
}

resource "auth0_user" "user" {
	depends_on = [ auth0_role.writer ]

	connection_name = "Username-Password-Authentication"

	email    = "{{.testName}}@auth0.com"
	password = "MyPass123$"
}

resource "auth0_organization" "org" {
	depends_on = [ auth0_user.user ]

	name         = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}

resource "auth0_organization_member" "member" {
	depends_on = [ auth0_organization.org ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id

	lifecycle {
		ignore_changes = [ roles ]
	}
}
`

const testAccOrganizationMemberRolesCreateWithOneRole = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_roles" "roles" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id

	roles = [ auth0_role.reader.id ]
}
`

const testAccOrganizationMemberRolesCreateWithTwoRoles = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_roles" "roles" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id

	roles = [ auth0_role.reader.id, auth0_role.writer.id ]
}
`

const testAccOrganizationMemberRolesCreateWithOneRoleRemoved = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_roles" "roles" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id

	roles = [ auth0_role.writer.id ]
}
`

const testAccOrganizationMemberRolesCreateWithNoRoles = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_roles" "roles" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id

	roles = []
}
`

func TestAccOrganizationMemberRoles(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberRolesCreateWithOneRole, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_member_roles.roles", "roles.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberRolesCreateWithTwoRoles, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_member_roles.roles", "roles.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberRolesCreateWithOneRoleRemoved, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_member_roles.roles", "roles.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberRolesCreateWithNoRoles, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_member_roles.roles", "roles.#", "0"),
				),
			},
		},
	})
}
