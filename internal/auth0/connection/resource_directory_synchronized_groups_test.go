package connection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDirectorySyncGroupsGivenAConnection = `
resource "auth0_connection" "my_connection" {
	name = "Acceptance-Test-Directory-SyncGroups-{{.testName}}"
	display_name = "Acceptance-Test-Directory-SyncGroups-{{.testName}}"
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
		api_enable_groups = true
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

resource "auth0_connection_directory" "my_directory" {
	connection_id = auth0_connection.my_connection.id
	synchronize_groups = "selected"
}
`

const testAccDirectorySyncGroupsEmpty = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
	group_ids     = []
}
`

const testAccDirectorySyncGroupsWithOne = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
	group_ids     = ["group1abc"]
}
`

const testAccDirectorySyncGroupsWithMultiple = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
	group_ids     = ["group1abc", "group2def", "group3ghi"]
}
`

const testAccDirectorySyncGroupsDelete = testAccDirectorySyncGroupsGivenAConnection

func TestAccDirectorySynchronizedGroups(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_connection_directory_synchronized_groups.my_sync_groups", "connection_id"),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsWithOne, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsWithMultiple, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "3"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsWithOne, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsDelete, t.Name()),
			},
		},
	})
}
