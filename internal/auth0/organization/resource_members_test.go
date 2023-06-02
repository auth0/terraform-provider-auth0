package organization_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccOrganizationMembersPreventErasingMembersOnCreate = `
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

resource "auth0_organization_member" "org_member_1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	user_id         = auth0_user.user_1.id
}

resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization_member.org_member_1 ]

	organization_id = auth0_organization.my_org.id
	members = [ auth0_user.user_2.id ]
}
`

const testAccOrganizationMembersWithOneMember = `
resource "auth0_user" "user_1" {
	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}1@auth0.com"
	password        = "MyPass123$"
}

resource "auth0_organization" "my_org" {
	depends_on = [ auth0_user.user_1 ]

	name         = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}

resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	members = [ auth0_user.user_1.id ]
}
`

const testAccOrganizationMembersWithTwoMembers = `
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

resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	members = [ auth0_user.user_1.id, auth0_user.user_2.id ]
}
`

const testAccOrganizationMembersRemoveOneMember = `
resource "auth0_user" "user_2" {
	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}2@auth0.com"
	password        = "MyPass123$"
}

resource "auth0_organization" "my_org" {
	depends_on = [ auth0_user.user_2 ]

	name         = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}

resource "auth0_organization_members" "my_members" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	members = [ auth0_user.user_2.id ]
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
				Config: acctest.ParseTestName(testAccOrganizationMembersWithOneMember, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_members.my_members", "members.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersWithTwoMembers, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_members.my_members", "members.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationMembersRemoveOneMember, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_members.my_members", "members.#", "1"),
				),
			},
		},
	})
}
