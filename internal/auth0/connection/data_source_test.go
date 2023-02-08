package connection_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/provider"
	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
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
	connection_id = auth0_connection.my_connection.id
	client_id     = auth0_client.my_client.id
}
`

const testAccDataConnectionConfigByName = `
data "auth0_connection" "test" {
	depends_on = [ auth0_connection_client.my_conn_client_assoc ]

	name = "Acceptance-Test-Connection-{{.testName}}"
}
`

const testAccDataConnectionConfigByID = `
data "auth0_connection" "test" {
	depends_on = [ auth0_connection_client.my_conn_client_assoc ]

	connection_id = auth0_connection.my_connection.id
}
`

func TestAccDataSourceConnectionRequiredArguments(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: provider.TestFactories(nil),
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_connection" "test" { }`,
				ExpectError: regexp.MustCompile("one of `connection_id,name` must be specified"),
			},
		},
	})
}

func TestAccDataSourceConnectionByName(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories:         provider.TestFactories(httpRecorder),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccGivenAConnection, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
				),
			},
			{
				Config: template.ParseTestName(testAccGivenAConnection+testAccDataConnectionConfigByName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
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
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories:         provider.TestFactories(httpRecorder),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccGivenAConnection, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
				),
			},
			{
				Config: template.ParseTestName(testAccGivenAConnection+testAccDataConnectionConfigByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_connection.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_connection.test", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_connection.test", "strategy", "auth0"),
					resource.TestCheckResourceAttr("data.auth0_connection.test", "enabled_clients.#", "1"),
				),
			},
		},
	})
}
