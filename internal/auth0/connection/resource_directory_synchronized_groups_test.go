package connection_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

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

// TestAccDirectorySynchronizedGroups covers the deprecated `group_ids` on its own, the configuration
// every existing practitioner already has, so it has to keep passing unchanged.
func TestAccDirectorySynchronizedGroups(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_connection_directory_synchronized_groups.my_sync_groups", "connection_id"),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "0"),
					// The read must leave the unused attribute alone, or every plan would propose
					// emptying a `groups` this practitioner never wrote.
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsWithOne, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.*", "group1abc"),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsWithMultiple, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "3"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.*", "group1abc"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.*", "group2def"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.*", "group3ghi"),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsWithOne, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.*", "group1abc"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsDelete, t.Name()),
			},
		},
	})
}

const testAccDirectorySyncGroupsAsGroupsWithMultiple = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id

	groups {
		id = "group1abc"
	}
	groups {
		id = "group2def"
	}
	groups {
		id = "group3ghi"
	}
}
`

const testAccDirectorySyncGroupsAsGroupsWithOne = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id

	groups {
		id = "group1abc"
	}
}
`

// TestAccDirectorySynchronizedGroupsAsGroups covers `groups` on its own: the lifecycle a new
// practitioner gets, and the one a migrated practitioner lands on.
func TestAccDirectorySynchronizedGroupsAsGroups(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsAsGroupsWithOne, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_connection_directory_synchronized_groups.my_sync_groups", "connection_id"),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group1abc",
					}),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsAsGroupsWithMultiple, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group1abc",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group2def",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group3ghi",
					}),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsAsGroupsWithOne, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group1abc",
					}),
				),
			},
			{
				// Importing lands the IDs in `groups`, the attribute that is not deprecated.
				ResourceName:      "auth0_connection_directory_synchronized_groups.my_sync_groups",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return acctest.ExtractResourceAttributeFromState(state, "auth0_connection.my_connection", "id")
				},
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsDelete, t.Name()),
			},
		},
	})
}

// Each migration configuration below declares exactly one of `group_ids` and `groups`, never both:
// `ConflictsWith` rejects an attribute being present at all rather than holding anything, so a
// practitioner migrating deletes the `group_ids` line rather than emptying it.

const testAccDirectorySyncGroupsMigrationFromGroupIDs = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
	group_ids     = ["group1abc", "group2def", "group3ghi"]
}
`

const testAccDirectorySyncGroupsMigrationToSameGroups = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id

	groups {
		id = "group1abc"
	}
	groups {
		id = "group2def"
	}
	groups {
		id = "group3ghi"
	}
}
`

const testAccDirectorySyncGroupsMigrationToMoreGroups = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id

	groups {
		id = "group1abc"
	}
	groups {
		id = "group2def"
	}
	groups {
		id = "group3ghi"
	}
	groups {
		id = "group4jkl"
	}
}
`

const testAccDirectorySyncGroupsMigrationToFewerGroups = testAccDirectorySyncGroupsGivenAConnection + `
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
`

const testAccDirectorySyncGroupsMigrationToDisjointGroups = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id

	groups {
		id = "group4jkl"
	}
	groups {
		id = "group5mno"
	}
}
`

const testAccDirectorySyncGroupsMigrationToNoGroups = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
}
`

const testAccDirectorySyncGroupsMigrationBackToGroupIDs = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
	group_ids     = ["group4jkl", "group5mno"]
}
`

const testAccDirectorySyncGroupsBothAttributes = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
	group_ids     = ["group1abc"]

	groups {
		id = "group1abc"
	}
}
`

const testAccDirectorySyncGroupsEmptiedGroupIDsBesideGroups = testAccDirectorySyncGroupsGivenAConnection + `
resource "auth0_connection_directory_synchronized_groups" "my_sync_groups" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
	group_ids     = []

	groups {
		id = "group1abc"
	}
}
`

// TestAccDirectorySynchronizedGroupsMigration walks a practitioner from the deprecated `group_ids`
// to `groups` and back, changing the IDs along the way: the same IDs, more, fewer, a disjoint set,
// none at all, and growing from none.
//
// The assertion that matters most is the framework's own empty-plan check at each step. It proves
// the migration converges, and that the Computed metadata settles, since a configuration declaring
// only `id` has to match a state carrying the name, email and member count too.
func TestAccDirectorySynchronizedGroupsMigration(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsMigrationFromGroupIDs, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "3"),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "0"),
				),
			},
			{
				// Migrate group_ids (1,2,3) => groups (1,2,3). Same IDs, so the API is left alone.
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsMigrationToSameGroups, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "0"),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group1abc",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group2def",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group3ghi",
					}),
				),
			},
			{
				// Then groups (1,2,3) => groups (1,2,3,4). Only the new group is added.
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsMigrationToMoreGroups, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "4"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group4jkl",
					}),
				),
			},
			{
				// Then groups (1,2,3,4) => groups (1,2). Only the dropped groups are removed.
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsMigrationToFewerGroups, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group1abc",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group2def",
					}),
				),
			},
			{
				// Then groups (1,2) => groups (4,5). Nothing in common, so the whole set turns over.
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsMigrationToDisjointGroups, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group4jkl",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.*", map[string]string{
						"id": "group5mno",
					}),
				),
			},
			{
				// Then groups (4,5) => neither attribute. This is how every group is unsynchronized
				// while the resource stays managed, and why the schema uses `ConflictsWith`.
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsMigrationToNoGroups, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "0"),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "0"),
				),
			},
			{
				// Then neither => groups (4,5). Growing from nothing.
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsMigrationToDisjointGroups, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "2"),
				),
			},
			{
				// Finally groups (4,5) => group_ids (4,5). Migrating back has to be just as free,
				// since nothing forces the move.
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsMigrationBackToGroupIDs, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.*", "group4jkl"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "group_ids.*", "group5mno"),
					resource.TestCheckResourceAttr("auth0_connection_directory_synchronized_groups.my_sync_groups", "groups.#", "0"),
				),
			},
			{
				// Setting both is rejected during config validation, so no API call is made.
				Config:      acctest.ParseTestName(testAccDirectorySyncGroupsBothAttributes, t.Name()),
				ExpectError: regexp.MustCompile(`conflicts with groups`),
			},
			{
				// Emptying the deprecated attribute is not a way around the conflict: it has to be
				// deleted. Pinned here because it is the first mistake a practitioner makes.
				Config:      acctest.ParseTestName(testAccDirectorySyncGroupsEmptiedGroupIDsBesideGroups, t.Name()),
				ExpectError: regexp.MustCompile(`conflicts with groups`),
			},
			{
				Config: acctest.ParseTestName(testAccDirectorySyncGroupsDelete, t.Name()),
			},
		},
	})
}
