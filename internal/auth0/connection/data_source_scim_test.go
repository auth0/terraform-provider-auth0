package connection_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSCIMConfigurationNonExistentID = `
data "auth0_connection_scim_configuration" "test" {
	connection_id = "con_xxxxxxxxxxxxxxxx"
}
`

const testAccDataSCIMConfigurationGivenAConnection = `
resource "auth0_connection" "my_connection" {
	name     = "Acceptance-Test-SCIM-Connection-{{.testName}}"
	display_name     = "Acceptance-Test-SCIM-{{.testName}}"
	strategy = "okta"
	options {
		client_id                = "1234567"
		client_secret            = "1234567"
		issuer                   = "https://example.okta.com"
		scopes 					 = ["openid", "profile", "email"]
		jwks_uri                 = "https://example.okta.com/oauth2/v1/keys"
		token_endpoint           = "https://example.okta.com/oauth2/v1/token"
		authorization_endpoint   = "https://example.okta.com/oauth2/v1/authorize"
	}
}
`

const testAccDataSCIMConfigurationWithNoSCIMConfiguration = testAccDataSCIMConfigurationGivenAConnection + `
data "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_connection.id
}
`

const testAccDataSCIMConfigurationWithSCIMConfiguration = testAccDataSCIMConfigurationGivenAConnection + `
resource "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.my_connection.id
}

data "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection_scim_configuration.my_scim_config.id
}
`

const testAccDataSCIMConfigurationDelete = testAccDataSCIMConfigurationGivenAConnection

func TestAccDataSCIMConfiguration(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_connection_scim_configuration" "test" { }`,
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      acctest.ParseTestName(testAccDataSCIMConfigurationNonExistentID, t.Name()),
				ExpectError: regexp.MustCompile("404 Not Found: The connection does not exist"),
			},
			{
				Config:      acctest.ParseTestName(testAccDataSCIMConfigurationWithNoSCIMConfiguration, t.Name()),
				ExpectError: regexp.MustCompile("404 Not Found: Not Found"),
			},
			{
				Config: acctest.ParseTestName(testAccDataSCIMConfigurationWithSCIMConfiguration, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "id", "auth0_connection_scim_configuration.my_scim_config", "connection_id"),
					resource.TestCheckResourceAttrPair("auth0_connection.my_connection", "name", "auth0_connection_scim_configuration.my_scim_config", "connection_name"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_scim_configuration.my_scim_config", "tenant_name"),
					resource.TestCheckResourceAttr("data.auth0_connection_scim_configuration.my_scim_config", "strategy", "okta"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_scim_configuration.my_scim_config", "user_id_attribute"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_scim_configuration.my_scim_config", "mapping.#"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_scim_configuration.my_scim_config", "default_mapping.#"),
					resource.TestCheckResourceAttrPair("data.auth0_connection_scim_configuration.my_scim_config", "connection_id", "auth0_connection_scim_configuration.my_scim_config", "connection_id"),
					resource.TestCheckResourceAttrPair("data.auth0_connection_scim_configuration.my_scim_config", "connection_name", "auth0_connection_scim_configuration.my_scim_config", "connection_name"),
					resource.TestCheckResourceAttrPair("data.auth0_connection_scim_configuration.my_scim_config", "tenant_name", "auth0_connection_scim_configuration.my_scim_config", "tenant_name"),
					resource.TestCheckResourceAttrPair("data.auth0_connection_scim_configuration.my_scim_config", "strategy", "auth0_connection_scim_configuration.my_scim_config", "strategy"),
					resource.TestCheckResourceAttrPair("data.auth0_connection_scim_configuration.my_scim_config", "user_id_attribute", "auth0_connection_scim_configuration.my_scim_config", "user_id_attribute"),
					resource.TestCheckResourceAttrPair("data.auth0_connection_scim_configuration.my_scim_config", "mapping.#", "auth0_connection_scim_configuration.my_scim_config", "mapping.#"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSCIMConfigurationDelete, t.Name()),
			},
		},
	})
}
