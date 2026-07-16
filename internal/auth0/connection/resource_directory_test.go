package connection_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDirectoryNonExistentID = `
resource "auth0_connection_directory" "test" {
	connection_id = "con_xxxxxxxxxxxxxxxx"
}
`

const testAccDirectoryGivenAConnection = `
resource "auth0_connection" "my_connection" {
	name = "Acceptance-Test-Directory-Connection-{{.testName}}"
	display_name = "Acceptance-Test-Directory-{{.testName}}"
	is_domain_connection = false
	strategy = "google-apps"
	show_as_button = false
	options {
		client_id = ""
		client_secret = ""
		domain = "example.com"
		tenant_domain = "example.com"
		domain_aliases = [ "example.com", "api.example.com" ]
		api_enable_users = true
		set_user_root_attributes = "on_first_login"
		map_user_id_to_id = true
		scopes = [ "ext_profile", "ext_groups" ]
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

const testAccDirectoryDelete = testAccDirectoryGivenAConnection

const testAccDirectoryWithDefaults = testAccDirectoryGivenAConnection + `
resource "auth0_connection_directory" "my_directory" {
	connection_id = auth0_connection.my_connection.id
}
`

const testAccDirectoryWithMapping = testAccDirectoryGivenAConnection + `
resource "auth0_connection_directory" "my_directory" {
	connection_id = auth0_connection.my_connection.id
	mapping {
		auth0 = "email"
		idp   = "primaryEmail"
	}
	mapping {
		auth0 = "external_id"
		idp   = "id"
	}
}
`

const testAccDirectoryWithUpdatedMapping = testAccDirectoryGivenAConnection + `
resource "auth0_connection_directory" "my_directory" {
	connection_id = auth0_connection.my_connection.id
	synchronize_automatically = false
	mapping {
		auth0 = "email"
		idp   = "primaryEmail"
	}
	mapping {
		auth0 = "blocked"
		idp   = "suspended"
	}
	mapping {
		auth0 = "app_metadata.org_unit"
		idp   = "orgUnitPath"
	}
}
`

func TestAccDirectoryProvisioning(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccDirectoryNonExistentID, t.Name()),
				ExpectError: regexp.MustCompile(``),
			},
			{
				Config: acctest.ParseTestName(testAccDirectoryWithDefaults, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory.my_directory", "synchronize_automatically", "false"),
					resource.TestCheckResourceAttrSet("auth0_connection_directory.my_directory", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_directory.my_directory", "connection_name"),
					resource.TestCheckResourceAttrSet("auth0_connection_directory.my_directory", "strategy"),
					resource.TestCheckResourceAttrSet("auth0_connection_directory.my_directory", "created_at"),
					resource.TestCheckResourceAttrSet("auth0_connection_directory.my_directory", "updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectoryWithMapping, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory.my_directory", "mapping.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectoryWithUpdatedMapping, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory.my_directory", "synchronize_automatically", "false"),
					resource.TestCheckResourceAttr("auth0_connection_directory.my_directory", "mapping.#", "3"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectoryDelete, t.Name()),
			},
		},
	})
}
