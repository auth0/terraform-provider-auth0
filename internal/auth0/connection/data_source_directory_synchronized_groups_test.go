package connection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDirectorySyncGroupsDataSource = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
	group_ids     = ["group1abc", "group2def"]
}

data "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory_synchronized_groups.my_sync_groups]
	connection_id = auth0_connection.my_connection.id
}
`

func TestAccDirectorySynchronizedGroupsDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsDataSource, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "connection_id"),
					resource.TestCheckResourceAttr("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "2"),
				),
			},
		},
	})
}
