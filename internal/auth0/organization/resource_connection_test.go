package organization_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccOrganizationConnectionWithOneConnectionEnabled = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_conn_1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	connection_id   = auth0_connection.my_connection_1.id
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_conn_1 ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionWithOneConnectionEnabledAndUpdated = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_conn_1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_connection_1.id
	assign_membership_on_login = true
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_conn_1 ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionWithTwoConnectionsEnabled = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_conn_1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_connection_1.id
	assign_membership_on_login = true
}

resource "auth0_organization_connection" "my_org_conn_2" {
	depends_on = [ auth0_organization_connection.my_org_conn_1 ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_connection_2.id
	assign_membership_on_login = true
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_conn_2 ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionImportSetup = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connections" "one_to_many" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id              = auth0_connection.my_connection_1.id
		assign_membership_on_login = true
	}

	enabled_connections {
		connection_id              = auth0_connection.my_connection_2.id
		assign_membership_on_login = true
	}
}
`

const testAccOrganizationConnectionImportCheck = testAccOrganizationConnectionImportSetup + `
resource "auth0_organization_connection" "my_org_conn_1" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_connection_1.id
	assign_membership_on_login = true
}

resource "auth0_organization_connection" "my_org_conn_2" {
	depends_on = [ auth0_organization_connection.my_org_conn_1 ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_connection_2.id
	assign_membership_on_login = true
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_conn_2 ]

	organization_id = auth0_organization.my_org.id
}
`

func TestAccOrganizationConnection(t *testing.T) {
	testName := strings.ToLower(t.Name())
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionWithOneConnectionEnabled, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_conn_1", "assign_membership_on_login", "false"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_conn_1", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_conn_1", "connection_id", "auth0_connection.my_connection_1", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionWithOneConnectionEnabledAndUpdated, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_conn_1", "assign_membership_on_login", "true"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_conn_1", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_conn_1", "connection_id", "auth0_connection.my_connection_1", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionWithTwoConnectionsEnabled, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "2"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_conn_1", "assign_membership_on_login", "true"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_conn_1", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_conn_1", "connection_id", "auth0_connection.my_connection_1", "id"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_conn_2", "assign_membership_on_login", "true"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_conn_2", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_conn_2", "connection_id", "auth0_connection.my_connection_2", "id"),
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
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationConnectionImportCheck, testName),
				ResourceName: "auth0_organization_connection.my_org_conn_1",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					orgID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_org", "id")
					assert.NoError(t, err)

					connID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_connection.my_connection_1", "id")
					assert.NoError(t, err)

					return orgID + "::" + connID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationConnectionImportCheck, testName),
				ResourceName: "auth0_organization_connection.my_org_conn_2",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					orgID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_org", "id")
					assert.NoError(t, err)

					connID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_connection.my_connection_2", "id")
					assert.NoError(t, err)

					return orgID + "::" + connID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "2"),
				),
			},
		},
	})
}
