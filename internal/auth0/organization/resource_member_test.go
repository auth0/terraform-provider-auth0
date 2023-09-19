package organization_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccOrganizationMemberAddOneMember = testAccGivenTwoUsersAndAnOrganization + `
resource "auth0_organization_member" "member_1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	user_id         = auth0_user.user_1.id
}

data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization_member.member_1 ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationMemberAddTwoMembers = testAccGivenTwoUsersAndAnOrganization + `
resource "auth0_organization_member" "member_1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	user_id         = auth0_user.user_1.id
}

resource "auth0_organization_member" "member_2" {
	depends_on = [ auth0_organization_member.member_1 ]

	organization_id = auth0_organization.my_org.id
	user_id         = auth0_user.user_2.id
}

data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization_member.member_2 ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationMemberImportSetup = testAccGivenTwoUsersAndAnOrganization + `
resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	members         = [ auth0_user.user_1.id, auth0_user.user_2.id ]
}
`

const testAccOrganizationMemberImportCheck = testAccOrganizationMemberImportSetup + `
resource "auth0_organization_member" "member_1" {
	depends_on = [ auth0_organization_members.my_members ]

	organization_id = auth0_organization.my_org.id
	user_id         = auth0_user.user_1.id
}

resource "auth0_organization_member" "member_2" {
	depends_on = [ auth0_organization_member.member_1 ]

	organization_id = auth0_organization.my_org.id
	user_id         = auth0_user.user_2.id
}

data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization_member.member_1 ]

	organization_id = auth0_organization.my_org.id
}
`

func TestAccOrganizationMember(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberAddOneMember, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_1", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_1", "user_id", "auth0_user.user_1", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberAddTwoMembers, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_1", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_1", "user_id", "auth0_user.user_1", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_2", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_2", "user_id", "auth0_user.user_2", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersDeleteResource, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationMemberImportCheck, testName),
				ResourceName: "auth0_organization_member.member_1",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					orgID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_org", "id")
					assert.NoError(t, err)

					userID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_user.user_1", "id")
					assert.NoError(t, err)

					return orgID + "::" + userID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationMemberImportCheck, testName),
				ResourceName: "auth0_organization_member.member_2",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					orgID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_org", "id")
					assert.NoError(t, err)

					userID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_user.user_2", "id")
					assert.NoError(t, err)

					return orgID + "::" + userID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMemberImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_1", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_1", "user_id", "auth0_user.user_1", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_2", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.member_2", "user_id", "auth0_user.user_2", "id"),
				),
			},
		},
	})
}
