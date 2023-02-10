package organization_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/provider"
	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

const testAccGivenAnOrganizationWithConnectionsAndMembers = `
resource "auth0_connection" "my_connection" {
	name     = "Acceptance-Test-Connection-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_organization" "my_organization" {
	depends_on = [auth0_connection.my_connection]

	name         = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}

resource "auth0_organization_connection" "my_org_conn" {
	depends_on = [auth0_organization.my_organization]

	organization_id = auth0_organization.my_organization.id
	connection_id   = auth0_connection.my_connection.id
}
`

const testAccDataSourceOrganizationConfigByName = testAccGivenAnOrganizationWithConnectionsAndMembers + `
data "auth0_organization" "test" {
	name = "test-{{.testName}}"
}
`

const testAccDataSourceOrganizationConfigByID = testAccGivenAnOrganizationWithConnectionsAndMembers + `
data "auth0_organization" "test" {
	organization_id = auth0_organization.my_organization.id
}
`

func TestAccDataSourceOrganizationRequiredArguments(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: provider.TestFactories(nil),
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_organization" "test" { }`,
				ExpectError: regexp.MustCompile("one of `name,organization_id` must be specified"),
			},
		},
	})
}

func TestAccDataSourceOrganizationByName(t *testing.T) {
	httpRecorder := recorder.New(t)
	testName := strings.ToLower(t.Name())

	resource.Test(t, resource.TestCase{
		ProviderFactories:         provider.TestFactories(httpRecorder),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccGivenAnOrganizationWithConnectionsAndMembers, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", testName)),
					resource.TestCheckResourceAttr("auth0_organization.my_organization", "name", fmt.Sprintf("test-%s", testName)),
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "organization_id"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_conn", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", testName)),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_conn", "strategy", "auth0"),
				),
			},
			{
				Config: template.ParseTestName(testAccDataSourceOrganizationConfigByName, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_organization.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "name", fmt.Sprintf("test-%s", testName)),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "connections.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_organization.test", "connections.0.connection_id"),
				),
			},
		},
	})
}

func TestAccDataSourceOrganizationByID(t *testing.T) {
	httpRecorder := recorder.New(t)
	testName := strings.ToLower(t.Name())

	resource.Test(t, resource.TestCase{
		ProviderFactories:         provider.TestFactories(httpRecorder),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccGivenAnOrganizationWithConnectionsAndMembers, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", testName)),
					resource.TestCheckResourceAttr("auth0_organization.my_organization", "name", fmt.Sprintf("test-%s", testName)),
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "organization_id"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_conn", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", testName)),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_conn", "strategy", "auth0"),
				),
			},
			{
				Config: template.ParseTestName(testAccDataSourceOrganizationConfigByID, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_organization.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "name", fmt.Sprintf("test-%s", testName)),
					resource.TestCheckResourceAttr("data.auth0_organization.test", "connections.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_organization.test", "connections.0.connection_id"),
				),
			},
		},
	})
}
