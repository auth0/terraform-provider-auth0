package organization_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccUpdateOrganizationMemberWithOneRoleAssigned = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_role" "role1" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id
	role_id         = auth0_role.reader.id
}
`

const testAccUpdateOrganizationMemberWithTwoRolesAssigned = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_role" "role1" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id
	role_id         = auth0_role.reader.id
}

resource "auth0_organization_member_role" "role2" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id
	role_id         = auth0_role.writer.id
}
`

const testAccRemoveAssignedRolesFromOrganizationMember = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_roles" "roles" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id

	roles = []
}
`

const testAccOrganizationMemberRolesImportSetup = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_roles" "roles" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id

	roles = [ auth0_role.reader.id, auth0_role.writer.id ]
}
`

const testAccOrganizationMemberRolesImportCheck = testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember + `
resource "auth0_organization_member_roles" "roles" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id

	roles = [ auth0_role.reader.id, auth0_role.writer.id ]
}

resource "auth0_organization_member_role" "role1" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id
	role_id         = auth0_role.reader.id
}

resource "auth0_organization_member_role" "role2" {
	depends_on = [ auth0_organization_member.member ]

	organization_id = auth0_organization.org.id
	user_id         = auth0_user.user.id
	role_id         = auth0_role.writer.id
}
`

func TestAccOrganizationMemberRole(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUpdateOrganizationMemberWithOneRoleAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_organization_member_role.role1", "user_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_member_role.role1", "organization_id"),
					resource.TestCheckResourceAttr("auth0_organization_member_role.role1", "role_name", fmt.Sprintf("Test Reader - %s", testName)),
					resource.TestCheckResourceAttr("auth0_organization_member_role.role1", "role_description", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateOrganizationMemberWithTwoRolesAssigned, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_organization_member_role.role1", "user_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_member_role.role1", "organization_id"),
					resource.TestCheckResourceAttr("auth0_organization_member_role.role1", "role_name", fmt.Sprintf("Test Reader - %s", testName)),
					resource.TestCheckResourceAttr("auth0_organization_member_role.role1", "role_description", ""),
					resource.TestCheckResourceAttrSet("auth0_organization_member_role.role2", "user_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_member_role.role2", "organization_id"),
					resource.TestCheckResourceAttr("auth0_organization_member_role.role2", "role_name", fmt.Sprintf("Test Writer - %s", testName)),
					resource.TestCheckResourceAttr("auth0_organization_member_role.role2", "role_description", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccGivenTwoRolesAUserAnOrganizationAndAnOrganizationMember, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccRemoveAssignedRolesFromOrganizationMember, testName),
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
				Config: acctest.ParseTestName(testAccRemoveAssignedRolesFromOrganizationMember, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberRolesImportSetup, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_member_roles.roles", "roles.#", "2"),
				),
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationMemberRolesImportCheck, testName),
				ResourceName: "auth0_organization_member_role.role1",
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

					roleID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_role.reader", "id")
					if err != nil {
						return "", err
					}

					return organizationID + "::" + userID + "::" + roleID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationMemberRolesImportCheck, testName),
				ResourceName: "auth0_organization_member_role.role2",
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

					roleID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_role.writer", "id")
					if err != nil {
						return "", err
					}

					return organizationID + "::" + userID + "::" + roleID, nil
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
			},
		},
	})
}
