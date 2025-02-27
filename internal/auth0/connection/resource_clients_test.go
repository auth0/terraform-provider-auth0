package connection_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccConnectionClientsPreventErasingEnabledClientsOnCreate(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
resource "auth0_connection" "my_conn" {
	name = "Acceptance-Test-Connection-PreventErasing"
	strategy = "auth0"
}

resource "auth0_client" "my_client" {
	depends_on = [ auth0_connection.my_conn ]
	name = "Acceptance-Test-Client-1-PreventErasing"
}

# Pre-existing enabled client
resource "auth0_connection_client" "one_to_one" {
	depends_on = [ auth0_client.my_client ]
	connection_id = auth0_connection.my_conn.id
	client_id = auth0_client.my_client.id
}

resource "auth0_connection_clients" "one_to_many" {
	depends_on = [ auth0_connection_client.one_to_one ]
	connection_id = auth0_connection.my_conn.id
	enabled_clients = []
}
`,
				ExpectError: regexp.MustCompile("Connection with non empty enabled clients"),
			},
		},
	})
}

const givenASingleConnection = `
resource "auth0_connection" "my_conn" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	strategy = "auth0"
}
`

const testAccConnectionClientsWithMinimalConfig = givenASingleConnection + `
resource "auth0_connection_clients" "my_conn_client_assoc" {
	depends_on = [ auth0_connection.my_conn ]

	connection_id = auth0_connection.my_conn.id
	enabled_clients = []
}
`

const testAccConnectionClientsWithOneEnabledClient = givenASingleConnection + `
resource "auth0_client" "my_client-1" {
	depends_on = [ auth0_connection.my_conn ]

	name = "Acceptance-Test-Client-1-{{.testName}}"
}

resource "auth0_connection_clients" "my_conn_client_assoc" {
	depends_on = [ auth0_client.my_client-1 ]

	connection_id = auth0_connection.my_conn.id
	enabled_clients = [auth0_client.my_client-1.id]
}
`

const testAccConnectionClientsWithTwoEnabledClients = givenASingleConnection + `
resource "auth0_client" "my_client-1" {
	depends_on = [ auth0_connection.my_conn ]

	name = "Acceptance-Test-Client-1-{{.testName}}"
}

resource "auth0_client" "my_client-2" {
	depends_on = [ auth0_client.my_client-1 ]

	name = "Acceptance-Test-Client-1-{{.testName}}"
}

resource "auth0_connection_clients" "my_conn_client_assoc" {
	depends_on = [ auth0_client.my_client-2 ]

	connection_id = auth0_connection.my_conn.id
	enabled_clients = [auth0_client.my_client-1.id, auth0_client.my_client-2.id]
}
`

func TestAccConnectionClients(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccConnectionClientsWithMinimalConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "enabled_clients.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccConnectionClientsWithOneEnabledClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "enabled_clients.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccConnectionClientsWithTwoEnabledClients, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "enabled_clients.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(givenASingleConnection, t.Name()),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "enabled_clients.#", "0"),
				),
			},
		},
	})
}

const testAccConnectionClientsImport = `
resource "auth0_connection" "my_conn" {
	name = "Acceptance-Test-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_client" "my_client-1" {
	depends_on = [ auth0_connection.my_conn ]

	name = "Acceptance-Test-Client-1-{{.testName}}"
}

resource "auth0_connection_clients" "my_conn_client_assoc" {
	depends_on = [ auth0_client.my_client-1 ]

	connection_id = auth0_connection.my_conn.id
	enabled_clients = [ auth0_client.my_client-1.id ]
}
`

func TestAccConnectionClientsImport(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// The test runs only with recordings as it requires an initial setup.
		// If you need to re-record it, comment out the t.Skip, create the
		// resources defined above and fix the IDs in the test before running
		// make record.
		t.Skip()
	}

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             acctest.ParseTestName(testAccConnectionClientsImport, t.Name()),
				ResourceName:       "auth0_connection.my_conn",
				ImportState:        true,
				ImportStateId:      "con_QfclIwGFilDZfig3",
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccConnectionClientsImport, t.Name()),
				ResourceName:       "auth0_client.my_client-1",
				ImportState:        true,
				ImportStateId:      "iUytqzFf4KueZRmFL7DRFNqn8RrQGu5Q",
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccConnectionClientsImport, t.Name()),
				ResourceName:       "auth0_connection_clients.my_conn_client_assoc",
				ImportState:        true,
				ImportStateId:      "con_QfclIwGFilDZfig3",
				ImportStatePersist: true,
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Config: acctest.ParseTestName(testAccConnectionClientsImport, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection_clients.my_conn_client_assoc", "enabled_clients.#", "1"),
				),
			},
		},
	})
}
