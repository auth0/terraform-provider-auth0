package connection_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataConnectionNonExistentID = `
data "auth0_connection" "test" {
	connection_id = "con_xxxxxxxxxxxxxxxx"
}
`

const testAccGivenAConnection = `
resource "auth0_connection" "my_connection" {
	name     = "Acceptance-Test-Connection-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_client" "my_client" {
	depends_on = [ auth0_connection.my_connection ]

	name     = "Acceptance Test - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_connection_client" "my_conn_client_assoc" {
	depends_on = [ auth0_client.my_client ]

	connection_id = auth0_connection.my_connection.id
	client_id     = auth0_client.my_client.id
}
`

const testAccDataConnectionConfigByName = testAccGivenAConnection + `
data "auth0_connection" "test-with-name" {
	depends_on = [ auth0_connection_client.my_conn_client_assoc ]

	name = "Acceptance-Test-Connection-{{.testName}}"
}
`

const testAccDataConnectionConfigByID = testAccGivenAConnection + `
data "auth0_connection" "test-with-id" {
	depends_on = [ auth0_connection_client.my_conn_client_assoc ]

	connection_id = auth0_connection.my_connection.id
}
`

func TestAccDataSourceConnectionRequiredArguments(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_connection" "test" { }`,
				ExpectError: regexp.MustCompile("one of `connection_id,name` must be specified"),
			},
		},
	})
}

func TestAccDataSourceConnection(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccDataConnectionNonExistentID, t.Name()),
				ExpectError: regexp.MustCompile("404 Not Found: The connection does not exist"),
			},
			{
				Config: acctest.ParseTestName(testAccDataConnectionConfigByName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttrSet("data.auth0_connection.test-with-name", "id"),
					resource.TestCheckResourceAttr("data.auth0_connection.test-with-name", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_connection.test-with-name", "strategy", "auth0"),
					resource.TestCheckResourceAttr("data.auth0_connection.test-with-name", "enabled_clients.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataConnectionConfigByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttrSet("data.auth0_connection.test-with-id", "id"),
					resource.TestCheckResourceAttr("data.auth0_connection.test-with-id", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_connection.test-with-id", "strategy", "auth0"),
					resource.TestCheckResourceAttr("data.auth0_connection.test-with-id", "enabled_clients.#", "1"),
				),
			},
		},
	})
}
