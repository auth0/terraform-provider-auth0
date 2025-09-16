package client_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenOneClient = `
resource "auth0_client" "my_client_1" {
    name = "Acceptance Test 1 - {{.testName}}"
    app_type = "non_interactive"
		is_first_party = true
    description = "Description for client 1 {{.testName}}"
}
`

const testAccGivenSomeClients = testAccGivenOneClient + `
resource "auth0_client" "my_client_2" {
    name = "Acceptance Test 2 - {{.testName}}"
    app_type = "spa"
		is_first_party = false
    description = "Description for client 2 {{.testName}}"
}
`

const testAccDataClientsWithNameFilter = `
data "auth0_clients" "test" {
    depends_on = [
        auth0_client.my_client_1,
        auth0_client.my_client_2
    ]

		name_filter = "{{.testName}}"
}
`

const testAccDataClientsWithAppTypeFilter = `
data "auth0_clients" "test" {
    depends_on = [
        auth0_client.my_client_1,
        auth0_client.my_client_2
    ]

		name_filter = "{{.testName}}"
    app_types = ["non_interactive"]
}
`

const testAccDataClientsWithIsFirstPartyFilter = `
data "auth0_clients" "test" {
    depends_on = [
        auth0_client.my_client_1,
        auth0_client.my_client_2
    ]

		name_filter = "{{.testName}}"
    is_first_party = true
}
`

const testAccDataClientsWithInvalidAppTypeFilter = `
data "auth0_clients" "test" {
    app_types = ["invalid"]
}
`

func TestAccDataClients(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataClientsWithInvalidAppTypeFilter, t.Name()),
				ExpectError: regexp.MustCompile(
					`expected app_types\.0 to be one of \["native" "spa" "regular_web" "non_interactive" "resource_server" "rms" "box" "cloudbees" "concur" "dropbox" "mscrm" "echosign" "egnyte" "newrelic" "office365" "salesforce" "sentry" "sharepoint" "slack" "springcm" "sso_integration" "zendesk" "zoom"\], got invalid`,
				),
			},
			{
				// We had to split this into two separate posts to work around an issue
				// in the test recording library. We need to add X-Request-Id header to the POST requests
				// to fix this, and make sure that go-vcr uses that to match requests.
				Config: acctest.ParseTestName(testAccGivenOneClient, t.Name()),
			},
			{
				Config: acctest.ParseTestName(testAccGivenSomeClients, t.Name()),
			},
			{
				Config: acctest.ParseTestName(testAccGivenSomeClients+testAccDataClientsWithNameFilter, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_clients.test", "clients.*", map[string]string{
						"name":           fmt.Sprintf("Acceptance Test 1 - %s", t.Name()),
						"app_type":       "non_interactive",
						"is_first_party": "true",
						"description":    fmt.Sprintf("Description for client 1 %s", t.Name()),
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_clients.test", "clients.*", map[string]string{
						"name":           fmt.Sprintf("Acceptance Test 2 - %s", t.Name()),
						"app_type":       "spa",
						"is_first_party": "false",
						"description":    fmt.Sprintf("Description for client 2 %s", t.Name()),
					}),
				),
			},
			{
				Config: acctest.ParseTestName(testAccGivenSomeClients+testAccDataClientsWithAppTypeFilter, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_clients.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_clients.test", "clients.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_clients.test", "clients.0.app_type", "non_interactive"),
					resource.TestCheckResourceAttr("data.auth0_clients.test", "clients.0.name", fmt.Sprintf("Acceptance Test 1 - %v", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccGivenSomeClients+testAccDataClientsWithIsFirstPartyFilter, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_clients.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_clients.test", "clients.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_clients.test", "clients.0.is_first_party", "true"),
					resource.TestCheckResourceAttr("data.auth0_clients.test", "clients.0.name", fmt.Sprintf("Acceptance Test 1 - %v", t.Name())),
				),
			},
		},
	})
}
