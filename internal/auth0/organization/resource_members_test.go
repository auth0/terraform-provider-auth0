package organization_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenTwoUsersAndAnOrganization = `
resource "auth0_user" "user_1" {
	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}1@auth0.com"
	password        = "MyPass123$"
}

resource "auth0_user" "user_2" {
	depends_on = [ auth0_user.user_1 ]

	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}2@auth0.com"
	password        = "MyPass123$"
}

resource "auth0_organization" "my_org" {
	depends_on = [ auth0_user.user_2 ]

	name         = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}
`

const testAccOrganizationMembersPreventErasingMembersOnCreate = testAccGivenTwoUsersAndAnOrganization + `
resource "auth0_organization_member" "org_member_1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	user_id         = auth0_user.user_1.id
}

resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization_member.org_member_1 ]

	organization_id = auth0_organization.my_org.id
	members         = [ auth0_user.user_2.id ]
}
`

const testAccOrganizationMembersWithOneMember = testAccGivenTwoUsersAndAnOrganization + `
resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	members         = [ auth0_user.user_1.id ]
}

data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization_members.my_members ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationMembersWithTwoMembers = testAccGivenTwoUsersAndAnOrganization + `
resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	members         = [ auth0_user.user_1.id, auth0_user.user_2.id ]
}

data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization_members.my_members ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationMembersRemoveOneMember = testAccGivenTwoUsersAndAnOrganization + `
resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	members         = [ auth0_user.user_2.id ]
}

data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization_members.my_members ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationMembersRemoveAllMembers = testAccGivenTwoUsersAndAnOrganization + `
resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	members         = []
}

data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization_members.my_members ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationMembersDeleteResource = testAccGivenTwoUsersAndAnOrganization + `
data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationMembersImportSetup = testAccGivenTwoUsersAndAnOrganization + `
resource "auth0_organization_member" "member_1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	user_id         = auth0_user.user_1.id
}

resource "auth0_organization_member" "member_2" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	user_id         = auth0_user.user_2.id
}
`

const testAccOrganizationMembersImportCheck = testAccOrganizationMembersImportSetup + `
resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization_member.member_2 ]

	organization_id = auth0_organization.my_org.id
	members         = [ auth0_user.user_1.id, auth0_user.user_2.id ]
}

data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization_members.my_members ]

	organization_id = auth0_organization.my_org.id
}
`

func TestAccOrganizationMembers(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccOrganizationMembersPreventErasingMembersOnCreate, testName),
				ExpectError: regexp.MustCompile("Organization with non empty members"),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersRemoveAllMembers, testName),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersWithOneMember, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_members.my_members", "members.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersWithTwoMembers, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_members.my_members", "members.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersRemoveOneMember, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_members.my_members", "members.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersRemoveAllMembers, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_members.my_members", "members.#", "0"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "0"),
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
				Config: acctest.ParseTestName(testAccOrganizationMembersImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationMembersImportCheck, testName),
				ResourceName: "auth0_organization_members.my_members",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_org", "id")
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "2"),
					resource.TestCheckResourceAttr("auth0_organization_members.my_members", "members.#", "2"),
				),
			},
		},
	})
}
