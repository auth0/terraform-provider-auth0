package organization_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenTwoClientsAndAnOrganization = `
resource "auth0_organization" "my_organization" {
	name         = "test-org-{{.testName}}"
	display_name = "Test Org {{.testName}}"
}

resource "auth0_client" "my_client" {
	name = "test-client-1-{{.testName}}"
}

# depends_on serializes the two client creations. Recorded HTTP interactions are matched
# on method and URL only, so two concurrent POSTs to /api/v2/clients could otherwise be
# replayed against the wrong client instance.
resource "auth0_client" "my_other_client" {
	depends_on = [ auth0_client.my_client ]

	name = "test-client-2-{{.testName}}"
}
`

const testAccOrganizationClientsCreate = testAccGivenTwoClientsAndAnOrganization + `
resource "auth0_organization_clients" "my_organization_clients" {
	organization_id = auth0_organization.my_organization.id

	clients {
		client_id             = auth0_client.my_client.id
		use_for_member_access = true
	}
}
`

const testAccOrganizationClientsAddOne = testAccGivenTwoClientsAndAnOrganization + `
resource "auth0_organization_clients" "my_organization_clients" {
	organization_id = auth0_organization.my_organization.id

	clients {
		client_id             = auth0_client.my_client.id
		use_for_member_access = true
	}

	clients {
		client_id = auth0_client.my_other_client.id
	}
}
`

const testAccOrganizationClientsFlipMemberAccess = testAccGivenTwoClientsAndAnOrganization + `
resource "auth0_organization_clients" "my_organization_clients" {
	organization_id = auth0_organization.my_organization.id

	clients {
		client_id             = auth0_client.my_client.id
		use_for_member_access = false
	}

	clients {
		client_id             = auth0_client.my_other_client.id
		use_for_member_access = true
	}
}
`

const testAccOrganizationClientsRemoveOne = testAccGivenTwoClientsAndAnOrganization + `
resource "auth0_organization_clients" "my_organization_clients" {
	organization_id = auth0_organization.my_organization.id

	clients {
		client_id             = auth0_client.my_other_client.id
		use_for_member_access = true
	}
}

data "auth0_organization_clients" "my_organization_clients" {
	depends_on = [ auth0_organization_clients.my_organization_clients ]

	organization_id = auth0_organization.my_organization.id
}
`

func TestAccOrganizationClients(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccOrganizationClientsCreate, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_organization_clients.my_organization_clients", "organization_id"),
					resource.TestCheckResourceAttr("auth0_organization_clients.my_organization_clients", "clients.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_organization_clients.my_organization_clients",
						"clients.*",
						map[string]string{"use_for_member_access": "true"},
					),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationClientsAddOne, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_clients.my_organization_clients", "clients.#", "2"),
					// The use_for_member_access default of false is what the API requires to be
					// sent explicitly for every client in the batch.
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_organization_clients.my_organization_clients",
						"clients.*",
						map[string]string{"use_for_member_access": "false"},
					),
				),
			},
			{
				// Flipping use_for_member_access on clients that are already associated must
				// be a PATCH per client, not a disassociate followed by a re-associate.
				Config: acctest.ParseTestName(testAccOrganizationClientsFlipMemberAccess, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_clients.my_organization_clients", "clients.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationClientsRemoveOne, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_clients.my_organization_clients", "clients.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_organization_clients.my_organization_clients", "clients.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_organization_clients.my_organization_clients", "clients.0.use_for_member_access", "true"),
				),
			},
			{
				Config:            acctest.ParseTestName(testAccOrganizationClientsRemoveOne, testName),
				ResourceName:      "auth0_organization_clients.my_organization_clients",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					organizationID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_organization", "id")
					assert.NoError(t, err)

					return organizationID, nil
				},
			},
		},
	})
}

const testAccOrganizationClientsBatchingBase = `
resource "auth0_organization" "my_organization" {
	name         = "test-org-{{.testName}}"
	display_name = "Test Org {{.testName}}"
}

# The one client this test has to tell apart from the others is created on its own and
# before them, so that the recorded HTTP interactions (which are matched on method and URL
# only) are always replayed against the same client instance.
resource "auth0_client" "my_kept_client" {
	name = "test-client-kept-{{.testName}}"
}

# The remaining 11 clients deliberately share one name: they are interchangeable for this
# test, and identical names keep the recorded interactions from being replayed against the
# wrong client instance.
resource "auth0_client" "my_clients" {
	depends_on = [ auth0_client.my_kept_client ]

	count = 11

	name = "test-client-{{.testName}}"
}
`

// No client ID is known at plan time here, and the kept client deliberately differs on
// use_for_member_access: unknown IDs all look alike to the SDK, so entries that disagree on
// the flag are where a plan time uniqueness check can mistake distinct clients for a duplicate.
const testAccOrganizationClientsBatchingAll = testAccOrganizationClientsBatchingBase + `
resource "auth0_organization_clients" "my_organization_clients" {
	organization_id = auth0_organization.my_organization.id

	clients {
		client_id             = auth0_client.my_kept_client.id
		use_for_member_access = true
	}

	dynamic "clients" {
		for_each = auth0_client.my_clients

		content {
			client_id = clients.value.id
		}
	}
}
`

const testAccOrganizationClientsBatchingKeepOne = testAccOrganizationClientsBatchingBase + `
resource "auth0_organization_clients" "my_organization_clients" {
	organization_id = auth0_organization.my_organization.id

	clients {
		client_id             = auth0_client.my_kept_client.id
		use_for_member_access = true
	}
}
`

// TestAccOrganizationClientsBatching covers associating and disassociating more clients than
// the batch endpoints accept in a single request (10), which has to be split across requests.
func TestAccOrganizationClientsBatching(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccOrganizationClientsBatchingAll, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_clients.my_organization_clients", "clients.#", "12"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_organization_clients.my_organization_clients",
						"clients.*",
						map[string]string{"use_for_member_access": "true"},
					),
				),
			},
			{
				// Eleven disassociations, which is one more than a single request accepts.
				Config: acctest.ParseTestName(testAccOrganizationClientsBatchingKeepOne, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_clients.my_organization_clients", "clients.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization_clients.my_organization_clients", "clients.0.use_for_member_access", "true"),
				),
			},
		},
	})
}
