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

	groups {
		id = "group1abc"
	}
	groups {
		id = "group2def"
	}
}

data "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory_synchronized_groups.my_sync_groups]
	connection_id = auth0_connection.my_connection.id
}
`

const testAccDirectorySyncGroupsDataSourceWithQuery = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id

	groups {
		id = "group1abc"
	}
	groups {
		id = "group2def"
	}
}

data "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory_synchronized_groups.my_sync_groups]
	connection_id = auth0_connection.my_connection.id
	query         = "name:zzzznomatch*"
}
`

// TestAccDirectorySynchronizedGroupsDataSource covers the data source populating both attributes,
// unlike the resource: they are outputs, so writing both cannot produce a diff.
func TestAccDirectorySynchronizedGroupsDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsDataSource, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "connection_id"),
					resource.TestCheckResourceAttr("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.*", "group1abc"),
					resource.TestCheckTypeSetElemAttr("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.*", "group2def"),
					resource.TestCheckResourceAttr("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group1abc",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group2def",
					}),
				),
			},
		},
	})
}

// TestAccDirectorySynchronizedGroupsDataSourceWithQuery asserts a term that matches nothing, which
// is what distinguishes a query being honored from one being silently dropped: the unfiltered read
// returns both groups, so an empty result can only mean the API applied the term.
func TestAccDirectorySynchronizedGroupsDataSourceWithQuery(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsDataSourceWithQuery, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "connection_id"),
					resource.TestCheckResourceAttr("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "query", "name:zzzznomatch*"),
					resource.TestCheckResourceAttr("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "0"),
					resource.TestCheckResourceAttr("data.auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "0"),
				),
			},
		},
	})
}
