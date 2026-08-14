package organization_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccOrganizationClientCreate = `
resource "auth0_organization" "my_organization" {
	name         = "test-org-{{.testName}}"
	display_name = "Test Org {{.testName}}"
}

resource "auth0_client" "my_test_client" {
	name               = "test-client-{{.testName}}"
	organization_usage = "allow"
}

resource "auth0_organization_client" "my_organization_client" {
	depends_on = [ auth0_organization.my_organization, auth0_client.my_test_client ]

	organization_id       = auth0_organization.my_organization.id
	client_id             = auth0_client.my_test_client.id
	use_for_member_access = true
}
`

const testAccOrganizationClientUpdate = `
resource "auth0_organization" "my_organization" {
	name         = "test-org-{{.testName}}"
	display_name = "Test Org {{.testName}}"
}

resource "auth0_client" "my_test_client" {
	name               = "test-client-{{.testName}}"
	organization_usage = "allow"
}

resource "auth0_organization_client" "my_organization_client" {
	depends_on = [ auth0_organization.my_organization, auth0_client.my_test_client ]

	organization_id       = auth0_organization.my_organization.id
	client_id             = auth0_client.my_test_client.id
	use_for_member_access = false
}

data "auth0_organization_client" "my_organization_client" {
	depends_on = [ auth0_organization_client.my_organization_client ]

	organization_id = auth0_organization.my_organization.id
	client_id       = auth0_client.my_test_client.id
}

data "auth0_organization_clients" "my_organization_clients" {
	depends_on = [ auth0_organization_client.my_organization_client ]

	organization_id = auth0_organization.my_organization.id
}
`

func TestAccOrganizationClient(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccOrganizationClientCreate, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_organization_client.my_organization_client", "organization_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_client.my_organization_client", "client_id"),
					resource.TestCheckResourceAttr("auth0_organization_client.my_organization_client", "use_for_member_access", "true"),
					resource.TestCheckResourceAttrSet("auth0_organization_client.my_organization_client", "name"),
					resource.TestCheckResourceAttr("auth0_organization_client.my_organization_client", "organization_usage", "allow"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationClientUpdate, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_client.my_organization_client", "use_for_member_access", "false"),
					resource.TestCheckResourceAttrSet("data.auth0_organization_client.my_organization_client", "name"),
					resource.TestCheckResourceAttr("data.auth0_organization_client.my_organization_client", "use_for_member_access", "false"),
					resource.TestCheckResourceAttr("auth0_organization_client.my_organization_client", "organization_usage", "allow"),
					resource.TestCheckResourceAttr("data.auth0_organization_client.my_organization_client", "organization_usage", "allow"),
					resource.TestCheckResourceAttr("data.auth0_organization_clients.my_organization_clients", "clients.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_organization_clients.my_organization_clients", "clients.0.client_id"),
					resource.TestCheckResourceAttr("data.auth0_organization_clients.my_organization_clients", "clients.0.organization_usage", "allow"),
				),
			},
			{
				Config:            acctest.ParseTestName(testAccOrganizationClientUpdate, testName),
				ResourceName:      "auth0_organization_client.my_organization_client",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					organizationID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_organization", "id")
					assert.NoError(t, err)

					clientID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_client.my_test_client", "id")
					assert.NoError(t, err)

					return organizationID + "::" + clientID, nil
				},
			},
		},
	})
}
