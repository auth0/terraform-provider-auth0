package organization_test

import (
	"fmt"
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
	username        = "tempusername1"
}

resource "auth0_user" "user_2" {
	depends_on = [ auth0_user.user_1 ]

	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}2@auth0.com"
	password        = "MyPass123$"
	username        = "tempusername2"
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
				Config: acctest.ParseTestName(testAccGivenTwoUsersAndAnOrganization, testName),
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

// testAccGivenManyUsersAndAnOrganization declares userCount users, each depending on
// the previous one so that they are always created in the same order, plus an organization.
func testAccGivenManyUsersAndAnOrganization(userCount int) string {
	var config strings.Builder

	for i := 1; i <= userCount; i++ {
		dependsOn := ""
		if i > 1 {
			dependsOn = fmt.Sprintf("\tdepends_on = [ auth0_user.user_%d ]\n\n", i-1)
		}

		fmt.Fprintf(&config, `
resource "auth0_user" "user_%[1]d" {
%[2]s	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}%[1]d@auth0.com"
	password        = "MyPass123$"
	username        = "orgmembers%[1]d"
}
`, i, dependsOn)
	}

	fmt.Fprintf(&config, `
resource "auth0_organization" "my_org" {
	depends_on = [ auth0_user.user_%d ]

	name         = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}
`, userCount)

	return config.String()
}

// testAccOrganizationMembersWithUsers assigns users user_<from> through user_<to> as members.
func testAccOrganizationMembersWithUsers(userCount, from, to int) string {
	members := make([]string, 0, to-from+1)
	for i := from; i <= to; i++ {
		members = append(members, fmt.Sprintf("auth0_user.user_%d.id", i))
	}

	return testAccGivenManyUsersAndAnOrganization(userCount) + fmt.Sprintf(`
resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	members         = [ %s ]
}

data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization_members.my_members ]

	organization_id = auth0_organization.my_org.id
}
`, strings.Join(members, ", "))
}

// testAccCheckOrganizationMembersAre checks that the resource and the organization data
// source hold exactly users user_<from> through user_<to>.
func testAccCheckOrganizationMembersAre(from, to int) resource.TestCheckFunc {
	count := fmt.Sprint(to - from + 1)
	checks := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr("auth0_organization_members.my_members", "members.#", count),
		resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", count),
	}

	for i := from; i <= to; i++ {
		user := fmt.Sprintf("auth0_user.user_%d", i)
		checks = append(checks,
			resource.TestCheckTypeSetElemAttrPair("auth0_organization_members.my_members", "members.*", user, "id"),
			resource.TestCheckTypeSetElemAttrPair("data.auth0_organization.my_org_data", "members.*", user, "id"),
		)
	}

	return resource.ComposeTestCheckFunc(checks...)
}

// TestAccOrganizationMembersBatch covers removing more than 10 members at
// once, which the Management API only accepts in batches of at most 10 user IDs.
func TestAccOrganizationMembersBatch(t *testing.T) {
	testName := strings.ToLower(t.Name())

	const userCount = 15

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersWithUsers(userCount, 1, 14), testName),
				Check:  testAccCheckOrganizationMembersAre(1, 14),
			},
			{
				// Removes 11 members and adds one in the same update.
				Config: acctest.ParseTestName(testAccOrganizationMembersWithUsers(userCount, 12, 15), testName),
				Check:  testAccCheckOrganizationMembersAre(12, 15),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersWithUsers(userCount, 1, 14), testName),
				Check:  testAccCheckOrganizationMembersAre(1, 14),
			},
			{
				// Destroys the resource while it holds 14 members.
				Config: acctest.ParseTestName(testAccGivenManyUsersAndAnOrganization(userCount)+`
data "auth0_organization" "my_org_data" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
}
`, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org_data", "members.#", "0"),
				),
			},
		},
	})
}
