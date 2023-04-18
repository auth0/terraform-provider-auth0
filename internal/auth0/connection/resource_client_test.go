package connection_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccCreateConnectionClient = `
resource "auth0_connection" "my_conn" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_client" "my_client-1" {
	depends_on = [ auth0_connection.my_conn ]

	name = "Acceptance-Test-Client-1-{{.testName}}"
}

resource "auth0_client" "my_client-2" {
	depends_on = [ auth0_client.my_client-1 ]

	name = "Acceptance-Test-Client-2-{{.testName}}"
}

resource "auth0_connection_client" "my_conn_client_assoc-1" {
	connection_id = auth0_connection.my_conn.id
	client_id     = auth0_client.my_client-1.id
}

resource "auth0_connection_client" "my_conn_client_assoc-2" {
	depends_on = [ auth0_connection_client.my_conn_client_assoc-1 ]

	connection_id = auth0_connection.my_conn.id
	client_id     = auth0_client.my_client-2.id
}
`

const testAccDeleteConnectionClient = `
resource "auth0_connection" "my_conn" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_client" "my_client-1" {
	name = "Acceptance-Test-Client-1-{{.testName}}"
}

resource "auth0_client" "my_client-2" {
	name = "Acceptance-Test-Client-2-{{.testName}}"
}

resource "auth0_connection_client" "my_conn_client_assoc-2" {
	connection_id = auth0_connection.my_conn.id
	client_id     = auth0_client.my_client-2.id
}
`

func TestAccConnectionClient(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateConnectionClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_client.my_client-1", "name", fmt.Sprintf("Acceptance-Test-Client-1-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client-2", "name", fmt.Sprintf("Acceptance-Test-Client-2-%s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc-1", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc-1", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc-1", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc-1", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc-2", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc-2", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc-2", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc-2", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDeleteConnectionClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_client.my_client-1", "name", fmt.Sprintf("Acceptance-Test-Client-1-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client-2", "name", fmt.Sprintf("Acceptance-Test-Client-2-%s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc-2", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc-2", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc-2", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc-2", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
				),
			},
		},
	})
}
