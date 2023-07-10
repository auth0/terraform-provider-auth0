package connection_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

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
data "auth0_connection" "test" {
	depends_on = [ auth0_connection_client.my_conn_client_assoc ]

	name = "Acceptance-Test-Connection-{{.testName}}"
}
`

const testAccDataConnectionConfigByID = testAccGivenAConnection + `
data "auth0_connection" "test" {
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

func TestAccDataSourceConnectionByName(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataConnectionConfigByName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttrSet("data.auth0_connection.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_connection.test", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_connection.test", "strategy", "auth0"),
					resource.TestCheckResourceAttr("data.auth0_connection.test", "enabled_clients.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceConnectionByID(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataConnectionConfigByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttrSet("data.auth0_connection.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_connection.test", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_connection.test", "strategy", "auth0"),
					resource.TestCheckResourceAttr("data.auth0_connection.test", "enabled_clients.#", "1"),
				),
			},
		},
	})
}

const testAccDataConnectionNonexistentID = `
data "auth0_connection" "test" {
	connection_id = "con_xxxxxxxxxxxxxxxx"
}
`

func TestAccDataSourceConnectionNonexistentID(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataConnectionNonexistentID, t.Name()),
				ExpectError: regexp.MustCompile(
					`data source with that identifier not found \((404\))`,
				),
			},
		},
	})
}
