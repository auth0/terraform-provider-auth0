package organization_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAnOrganizationWithConnectionsAndMembers = `
resource "auth0_user" "user" {
	connection_name = "Username-Password-Authentication"
	email           = "{{.testName}}@auth0.com"
	password        = "MyPass123$"
	username        = "{{.testName}}"
}

resource "auth0_connection" "my_connection" {
	depends_on = [ auth0_user.user ]

	name     = "Acceptance-Test-Connection-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_organization" "my_organization" {
	depends_on = [ auth0_connection.my_connection ]

	name         = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}

resource "auth0_organization_connection" "my_org_conn" {
	depends_on = [ auth0_organization.my_organization ]

	organization_id = auth0_organization.my_organization.id
	connection_id   = auth0_connection.my_connection.id
}

resource "auth0_organization_member" "org_member" {
	depends_on = [ auth0_organization_connection.my_org_conn ]

	organization_id = auth0_organization.my_organization.id
	user_id         = auth0_user.user.id
}
`

const testAccDataSourceOrganizationConfigByName = testAccGivenAnOrganizationWithConnectionsAndMembers + `
data "auth0_organization" "test" {
	depends_on = [ auth0_organization_member.org_member ]

	name = "test-{{.testName}}"
}
`

const testAccDataSourceOrganizationConfigByID = testAccGivenAnOrganizationWithConnectionsAndMembers + `
data "auth0_organization" "test" {
	depends_on = [ auth0_organization_member.org_member ]

	organization_id = auth0_organization.my_organization.id
}
`

const testAccDataSourceOrganizationNonExistentID = `
data "auth0_organization" "test" {
	organization_id = "org_XXXXXXXXXXXXXXXX"
}
`

func TestAccDataSourceOrganizationRequiredArguments(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_organization" "test" { }`,
				ExpectError: regexp.MustCompile("one of `name,organization_id` must be specified"),
			},
		},
	})
}

func TestAccDataSourceOrganization(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceOrganizationNonExistentID, t.Name()),
				ExpectError: regexp.MustCompile(
					"404 Not Found: No organization found by that id or name",
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceOrganizationConfigByName, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_organization.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "name", fmt.Sprintf("test-%s", testName)),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "connections.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_organization.test", "connections.0.connection_id"),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "members.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceOrganizationConfigByID, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_organization.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "name", fmt.Sprintf("test-%s", testName)),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "connections.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_organization.test", "connections.0.connection_id"),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "members.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceOrganizationInsufficientScope(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				// Uncomment the below code to test functionality against a non enabled tenants and record it.
				//PreConfig: func() {
				//	// Add your env variables here.
				//	_ = os.Setenv("AUTH0_DOMAIN", "some-domain-name")
				//	_ = os.Setenv("AUTH0_CLIENT_ID", "some-client-id")
				//	_ = os.Setenv("AUTH0_CLIENT_SECRET", "some-client-secret")
				// },.
				Config: `data "auth0_organization" "test" {
					organization_id = "org_P0nITxTkwnKQvD22"
				}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.test", "grant_ids.#", "0")),
			},
		},
	})
}
