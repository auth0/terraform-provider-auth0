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
resource "auth0_organization_connection" "my_org_db_conn" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	connection_id   = auth0_connection.my_db_connection.id
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_db_conn ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionSetShowAsButton = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_enterprise_conn" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	connection_id   = auth0_connection.my_enterprise_connection.id
	show_as_button  = true
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_enterprise_conn ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionClearShowAsButton = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_enterprise_conn" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	connection_id   = auth0_connection.my_enterprise_connection.id
	show_as_button  = false
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_enterprise_conn ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionSetIsSignupEnabled = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_db_conn" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	connection_id   = auth0_connection.my_db_connection.id
	assign_membership_on_login  = true
	is_signup_enabled  = true
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_db_conn ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionClearIsSignupEnabled = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_db_conn" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	connection_id   = auth0_connection.my_db_connection.id
	assign_membership_on_login  = true
	is_signup_enabled  = false
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_db_conn ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionWithOneConnectionEnabledAndUpdated = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_db_conn" {
	depends_on = [ auth0_organization.my_org ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_db_connection.id
	assign_membership_on_login = true
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_db_conn ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionWithTwoConnectionsEnabled = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_db_conn" {
	depends_on = [ auth0_organization.my_org ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_db_connection.id
	assign_membership_on_login = true
}

resource "auth0_organization_connection" "my_org_enterprise_conn" {
	depends_on = [ auth0_organization_connection.my_org_db_conn ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_enterprise_connection.id
	assign_membership_on_login = true
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_enterprise_conn ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionImportSetup = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connections" "one_to_many" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id              = auth0_connection.my_db_connection.id
		assign_membership_on_login = true
	}

	enabled_connections {
		connection_id              = auth0_connection.my_enterprise_connection.id
		assign_membership_on_login = true
	}
}
`

const testAccOrganizationConnectionImportCheck = testAccOrganizationConnectionImportSetup + `
resource "auth0_organization_connection" "my_org_db_conn" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_db_connection.id
	assign_membership_on_login = true
}

resource "auth0_organization_connection" "my_org_enterprise_conn" {
	depends_on = [ auth0_organization_connection.my_org_db_conn ]

	organization_id            = auth0_organization.my_org.id
	connection_id              = auth0_connection.my_enterprise_connection.id
	assign_membership_on_login = true
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_connection.my_org_enterprise_conn ]

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
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "assign_membership_on_login", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "show_as_button", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.assign_membership_on_login", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.show_as_button", "true"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_db_conn", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_db_conn", "connection_id", "auth0_connection.my_db_connection", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionWithOneConnectionEnabledAndUpdated, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "show_as_button", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.show_as_button", "true"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_db_conn", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_db_conn", "connection_id", "auth0_connection.my_db_connection", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionWithTwoConnectionsEnabled, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "2"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "show_as_button", "true"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_db_conn", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_db_conn", "connection_id", "auth0_connection.my_db_connection", "id"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_enterprise_conn", "assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_enterprise_conn", "is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_enterprise_conn", "show_as_button", "true"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_enterprise_conn", "organization_id", "auth0_organization.my_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_connection.my_org_enterprise_conn", "connection_id", "auth0_connection.my_enterprise_connection", "id"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.show_as_button", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.1.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.1.is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.1.show_as_button", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsDelete, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionSetShowAsButton, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_enterprise_conn", "assign_membership_on_login", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_enterprise_conn", "is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_enterprise_conn", "show_as_button", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.assign_membership_on_login", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.show_as_button", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionClearShowAsButton, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_enterprise_conn", "assign_membership_on_login", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_enterprise_conn", "is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_enterprise_conn", "show_as_button", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.assign_membership_on_login", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.show_as_button", "false"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsDelete, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionSetIsSignupEnabled, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "is_signup_enabled", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "show_as_button", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.is_signup_enabled", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.show_as_button", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionClearIsSignupEnabled, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connection.my_org_db_conn", "show_as_button", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.0.show_as_button", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsDelete, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.my_org", "connections.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationConnectionImportCheck, testName),
				ResourceName: "auth0_organization_connection.my_org_db_conn",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					orgID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_org", "id")
					assert.NoError(t, err)

					connID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_connection.my_db_connection", "id")
					assert.NoError(t, err)

					return orgID + "::" + connID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationConnectionImportCheck, testName),
				ResourceName: "auth0_organization_connection.my_org_enterprise_conn",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					orgID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_org", "id")
					assert.NoError(t, err)

					connID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_connection.my_enterprise_connection", "id")
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
