package connection_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccSCIMConfigurationNonExistentID = `
resource "auth0_connection_scim_configuration" "test" {
	connection_id = "con_xxxxxxxxxxxxxxxx"
}
`

const testAccSCIMConfigurationGivenAnUnsupportedConnection = `
resource "auth0_connection" "my_unsupported_connection" {
	name     = "Acceptance-Test-Non-SCIM-Connection-{{.testName}}"
	display_name     = "Acceptance-Test-Non-SCIM-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_unsupported_connection.id
}
`

const testAccSCIMConfigurationGivenAConnection = `
resource "auth0_connection" "my_connection" {
	name     = "Acceptance-Test-SCIM-Connection-{{.testName}}"
	display_name     = "Acceptance-Test-SCIM-{{.testName}}"
	strategy = "okta"
	options {
		client_id                = "1234567"
		client_secret            = "1234567"
		scopes 					 = ["openid", "profile", "email"]
		issuer                   = "https://example.okta.com"
		jwks_uri                 = "https://example.okta.com/oauth2/v1/keys"
		token_endpoint           = "https://example.okta.com/oauth2/v1/token"
		authorization_endpoint   = "https://example.okta.com/oauth2/v1/authorize"
	}
}
`

const testAccSCIMConfigurationDelete = testAccSCIMConfigurationGivenAConnection

const testAccSCIMConfigurationWithDefaults = testAccSCIMConfigurationGivenAConnection + `
resource "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_connection.id
}
`

const testAccSCIMConfigurationDeleted = testAccSCIMConfigurationDelete + `
data "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_connection.id
}
`

const testAccSCIMConfigurationWithUserIDAttributeOnly = testAccSCIMConfigurationGivenAConnection + `
resource "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_connection.id
	user_id_attribute = "test_attribute"
}
`

const testAccSCIMConfigurationWithMappingOnly = testAccSCIMConfigurationGivenAConnection + `
resource "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_connection.id
	mapping {
		auth0 = "attr_auth0"
		scim = "scim"
	}
}
`

const testAccSCIMConfigurationWithMapping = testAccSCIMConfigurationGivenAConnection + `
resource "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_connection.id
	user_id_attribute = "test_attribute"
	mapping {
		auth0 = "attr_auth0"
		scim = "scim"
	}
}
`

const testAccSCIMConfigurationWithChangedUserIDAttribute = testAccSCIMConfigurationGivenAConnection + `
resource "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_connection.id
	user_id_attribute = "modified_test_attribute"
	mapping {
		auth0 = "attr_auth0"
		scim = "scim"
	}
}
`

const testAccSCIMConfigurationWithMultipleMappings = testAccSCIMConfigurationGivenAConnection + `
resource "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_connection.id
	user_id_attribute = "test_attribute"
	mapping {
		auth0 = "attr_auth0_1"
		scim = "scim_1"
	}
	mapping {
		auth0 = "attr_auth0_2"
		scim = "scim_2"
	}
}
`

func TestAccSCIMConfiguration(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_connection_scim_configuration" "test" { }`,
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      acctest.ParseTestName(testAccSCIMConfigurationNonExistentID, t.Name()),
				ExpectError: regexp.MustCompile("404 Not Found: The connection does not exist"),
			},
			{
				Config:      acctest.ParseTestName(testAccSCIMConfigurationGivenAnUnsupportedConnection, t.Name()),
				ExpectError: regexp.MustCompile("404 Not Found: This connection type does not support SCIM."),
			},
			{
				Config: acctest.ParseTestName(testAccSCIMConfigurationWithDefaults, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "id", "auth0_connection_scim_configuration.my_scim_config", "connection_id"),
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "name", "auth0_connection_scim_configuration.my_scim_config", "connection_name"),
					resource.TestCheckResourceAttrSet("auth0_connection_scim_configuration.my_scim_config", "tenant_name"),
					resource.TestCheckResourceAttr("auth0_connection_scim_configuration.my_scim_config", "strategy", "okta"),
					resource.TestCheckResourceAttrSet("auth0_connection_scim_configuration.my_scim_config", "user_id_attribute"),
					resource.TestCheckResourceAttrSet("auth0_connection_scim_configuration.my_scim_config", "mapping.#"),
				),
			},
			{
				Config:      acctest.ParseTestName(testAccSCIMConfigurationDeleted, t.Name()),
				ExpectError: regexp.MustCompile("404 Not Found: Not Found"),
			},
			{
				Config:      acctest.ParseTestName(testAccSCIMConfigurationWithUserIDAttributeOnly, t.Name()),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      acctest.ParseTestName(testAccSCIMConfigurationWithMappingOnly, t.Name()),
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config: acctest.ParseTestName(testAccSCIMConfigurationWithMultipleMappings, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "id", "auth0_connection_scim_configuration.my_scim_config", "connection_id"),
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "name", "auth0_connection_scim_configuration.my_scim_config", "connection_name"),
					resource.TestCheckResourceAttrSet("auth0_connection_scim_configuration.my_scim_config", "tenant_name"),
					resource.TestCheckResourceAttr("auth0_connection_scim_configuration.my_scim_config", "strategy", "okta"),
					resource.TestCheckResourceAttrSet("auth0_connection_scim_configuration.my_scim_config", "user_id_attribute"),
					resource.TestCheckResourceAttr("auth0_connection_scim_configuration.my_scim_config", "mapping.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccSCIMConfigurationWithMapping, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "id", "auth0_connection_scim_configuration.my_scim_config", "connection_id"),
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "name", "auth0_connection_scim_configuration.my_scim_config", "connection_name"),
					resource.TestCheckResourceAttrSet("auth0_connection_scim_configuration.my_scim_config", "tenant_name"),
					resource.TestCheckResourceAttr("auth0_connection_scim_configuration.my_scim_config", "strategy", "okta"),
					resource.TestCheckResourceAttr("auth0_connection_scim_configuration.my_scim_config", "user_id_attribute", "test_attribute"),
					resource.TestCheckResourceAttr("auth0_connection_scim_configuration.my_scim_config", "mapping.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccSCIMConfigurationWithChangedUserIDAttribute, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "id", "auth0_connection_scim_configuration.my_scim_config", "connection_id"),
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "name", "auth0_connection_scim_configuration.my_scim_config", "connection_name"),
					resource.TestCheckResourceAttrSet("auth0_connection_scim_configuration.my_scim_config", "tenant_name"),
					resource.TestCheckResourceAttr("auth0_connection_scim_configuration.my_scim_config", "strategy", "okta"),
					resource.TestCheckResourceAttr("auth0_connection_scim_configuration.my_scim_config", "user_id_attribute", "modified_test_attribute"),
					resource.TestCheckResourceAttr("auth0_connection_scim_configuration.my_scim_config", "mapping.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccSCIMConfigurationDelete, t.Name()),
			},
		},
	})
}
