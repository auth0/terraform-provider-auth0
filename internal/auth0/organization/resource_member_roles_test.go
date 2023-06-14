package organization_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

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

const testAccOrganizationMemberRolesImportSetup = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_role" "role1" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id
	role_id         = auth0_role.reader.id
}

resource "auth0_organization_member_role" "role2" {
	depends_on = [ auth0_organization_member_role.role1 ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id
	role_id         = auth0_role.writer.id
}
`

const testAccOrganizationMemberRolesImportCheck = testAccOrganizationMemberRolesImportSetup + `
resource "auth0_organization_member_roles" "roles" {
	depends_on = [ auth0_organization_member_role.role2 ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id

	roles = [ auth0_role.reader.id, auth0_role.writer.id ]
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
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberRolesImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationMemberRolesImportCheck, testName),
				ResourceName: "auth0_organization_member_roles.roles",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					organizationID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_organization.org", "id")
					if err != nil {
						return "", err
					}

					userID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_user.user", "id")
					if err != nil {
						return "", err
					}

					return organizationID + ":" + userID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberRolesImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_member_roles.roles", "roles.#", "2"),
				),
			},
		},
	})
}
