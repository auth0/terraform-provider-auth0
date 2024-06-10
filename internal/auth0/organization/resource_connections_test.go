package organization_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenTwoConnectionsAndAnOrganization = `
resource "auth0_connection" "my_db_connection" {
	name     = "Acceptance-Test-DB-Connection-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_connection" "my_enterprise_connection" {
	depends_on = [ auth0_connection.my_db_connection ]

	name     = "Acceptance-Test-Enterprise-Connection-{{.testName}}"
	display_name = "{{.testName}}"
	strategy = "okta"
	options {
		client_id                = "1234567"
		client_secret            = "1234567"
		#domain                   = "example.okta.com"
		domain_aliases           = ["example.com"]
		issuer                   = "https://example.okta.com"
		jwks_uri                 = "https://example.okta.com/oauth2/v1/keys"
		token_endpoint           = "https://example.okta.com/oauth2/v1/token"
		userinfo_endpoint        = "https://example.okta.com/oauth2/v1/userinfo"
		authorization_endpoint   = "https://example.okta.com/oauth2/v1/authorize"
		scopes                   = ["openid", "email"]
		set_user_root_attributes = "on_first_login"
		non_persistent_attrs     = ["ethnicity", "gender"]
		upstream_params = jsonencode({
			"screen_name" : {
				"alias" : "login_hint"
			}
		})
		connection_settings {
			pkce = "auto"
		}
		attribute_map {
			mapping_mode   = "basic_profile"
			userinfo_scope = "openid email profile groups"
			attributes = jsonencode({
			"name" : "$${context.tokenset.name}",
			"email" : "$${context.tokenset.email}",
			"email_verified" : "$${context.tokenset.email_verified}",
			"nickname" : "$${context.tokenset.nickname}",
			"picture" : "$${context.tokenset.picture}",
			"given_name" : "$${context.tokenset.given_name}",
			"family_name" : "$${context.tokenset.family_name}"
			})
		}
	}
}

resource "auth0_organization" "my_org" {
	depends_on = [ auth0_connection.my_enterprise_connection ]

	name         = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}
`

const testAccOrganizationConnectionsPreventErasingConnectionsOnCreate = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connection" "my_org_connection_1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	connection_id   = auth0_connection.my_db_connection.id
}


resource "auth0_organization_connections" "one_many" {
	depends_on = [ auth0_organization_connection.my_org_connection_1 ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id = auth0_connection.my_enterprise_connection.id
	}
}
`

const testAccOrganizationConnectionsWithOneConnection = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connections" "one_to_many" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id = auth0_connection.my_db_connection.id
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsWithTwoConnections = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connections" "one_to_many" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id = auth0_connection.my_db_connection.id
	}

	enabled_connections {
		connection_id = auth0_connection.my_enterprise_connection.id
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsUpdateAssignMembership = testAccGivenTwoConnectionsAndAnOrganization + `
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

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsSetShowAsButton = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connections" "one_to_many" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id              = auth0_connection.my_enterprise_connection.id
		show_as_button             = true
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsClearShowAsButton = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connections" "one_to_many" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id              = auth0_connection.my_enterprise_connection.id
		show_as_button             = false
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsSetIsSignupEnabled = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connections" "one_to_many" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id              = auth0_connection.my_db_connection.id
		assign_membership_on_login = true
		is_signup_enabled          = true
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsClearIsSignupEnabled = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connections" "one_to_many" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id              = auth0_connection.my_db_connection.id
		assign_membership_on_login = true
		is_signup_enabled          = false
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsRemoveOneConnection = testAccGivenTwoConnectionsAndAnOrganization + `
resource "auth0_organization_connections" "one_to_many" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	enabled_connections {
		connection_id              = auth0_connection.my_enterprise_connection.id
		assign_membership_on_login = true
	}
}

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationConnectionsDelete = testAccGivenTwoConnectionsAndAnOrganization + `
data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
}

`

const testAccOrganizationConnectionsImportSetup = testAccGivenTwoConnectionsAndAnOrganization + `
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
`

const testAccOrganizationConnectionsImportCheck = testAccOrganizationConnectionsImportSetup + `
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

data "auth0_organization" "org_data" {
	depends_on = [ auth0_organization_connections.one_to_many ]

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
				Config: acctest.ParseTestName(testAccOrganizationConnectionsDelete, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsWithOneConnection, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsWithTwoConnections, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "2"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.#", "2"),
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
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.#", "2"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.0.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.1.assign_membership_on_login", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsRemoveOneConnection, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.0.assign_membership_on_login", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.0.assign_membership_on_login", "true"),
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
				Config: acctest.ParseTestName(testAccOrganizationConnectionsSetShowAsButton, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.0.show_as_button", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.0.show_as_button", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsClearShowAsButton, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.0.show_as_button", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.0.show_as_button", "false"),
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
				Config: acctest.ParseTestName(testAccOrganizationConnectionsSetIsSignupEnabled, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.0.is_signup_enabled", "true"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.0.is_signup_enabled", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsClearIsSignupEnabled, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.0.is_signup_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.0.is_signup_enabled", "false"),
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
				Config: acctest.ParseTestName(testAccOrganizationConnectionsImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccOrganizationConnectionsImportCheck, testName),
				ResourceName: "auth0_organization_connections.one_to_many",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_org", "id")
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationConnectionsImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organization.org_data", "connections.#", "2"),
					resource.TestCheckResourceAttr("auth0_organization_connections.one_to_many", "enabled_connections.#", "2"),
				),
			},
		},
	})
}
