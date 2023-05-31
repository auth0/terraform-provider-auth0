package organization_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccOrganizationConnectionsPreventErasingConnectionsOnCreate = `
resource "auth0_organization" "my_organization" {
	name = "test-1-{{.testName}}"
}

resource "auth0_connection" "my-connection-1" {
	depends_on = [ auth0_organization.my_organization ]

	name     = "Acceptance-Test-Connection-1-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_connection" "my-connection-2" {
	depends_on = [ auth0_connection.my-connection-1 ]

	name     = "Acceptance-Test-Connection-2-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_organization_connection" "my_org_connection-1" {
	depends_on = [ auth0_connection.my-connection-2 ]

	organization_id = auth0_organization.my_organization.id
	connection_id   = auth0_connection.my-connection-1.id
}


resource "auth0_organization_connections" "one-many" {
	depends_on = [ auth0_organization_connection.my_org_connection-1 ]

	organization_id = auth0_organization.my_organization.id

	enabled_connections {
		connection_id = auth0_connection.my-connection-2.id
	}
}
`

const testAccOrganizationConnectionsWithOneConnection = `
resource "auth0_organization" "my_org" {
	name         = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}

resource "auth0_connection" "my-conn-1" {
	depends_on = [ auth0_organization.my_org ]

	name     = "Acceptance-Test-Conn-1-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_organization_connections" "one-to-many" {
	depends_on = [ auth0_connection.my-conn-1 ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id = auth0_connection.my-conn-1.id
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one-to-many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsWithTwoConnections = `
resource "auth0_organization" "my_org" {
	name         = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}

resource "auth0_connection" "my-conn-1" {
	depends_on = [ auth0_organization.my_org ]

	name     = "Acceptance-Test-Conn-1-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_connection" "my-conn-2" {
	depends_on = [ auth0_connection.my-conn-1 ]

	name     = "Acceptance-Test-Conn-2-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_organization_connections" "one-to-many" {
	depends_on = [ auth0_connection.my-conn-2 ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id = auth0_connection.my-conn-1.id
	}

	enabled_connections {
		connection_id = auth0_connection.my-conn-2.id
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one-to-many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsUpdateAssignMembership = `
resource "auth0_organization" "my_org" {
	name         = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}

resource "auth0_connection" "my-conn-1" {
	depends_on = [ auth0_organization.my_org ]

	name     = "Acceptance-Test-Conn-1-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_connection" "my-conn-2" {
	depends_on = [ auth0_connection.my-conn-1 ]

	name     = "Acceptance-Test-Conn-2-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_organization_connections" "one-to-many" {
	depends_on = [ auth0_connection.my-conn-2 ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id              = auth0_connection.my-conn-1.id
		assign_membership_on_login = true
	}

	enabled_connections {
		connection_id              = auth0_connection.my-conn-2.id
		assign_membership_on_login = true
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one-to-many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsDelete = `
resource "auth0_organization" "my_org" {
	name         = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
}
`

func TestAccOrganizationConnections(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccOrganizationConnectionsPreventErasingConnectionsOnCreate, testName),
				ExpectError: regexp.MustCompile("Organization with non empty enabled connections"),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsWithOneConnection, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one-to-many", "enabled_connections.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsWithTwoConnections, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "2"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one-to-many", "enabled_connections.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsUpdateAssignMembership, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.0.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.1.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one-to-many", "enabled_connections.#", "2"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one-to-many", "enabled_connections.0.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one-to-many", "enabled_connections.0.assign_membership_on_login", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsDelete, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "0"),
				),
			},
		},
	})
}
