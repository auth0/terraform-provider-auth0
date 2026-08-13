package client_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceClientGrantOrganizationsConfig = `
resource "auth0_organization" "my_organization" {
	name                      = "test-org-{{.testName}}"
	display_name              = "Test Org {{.testName}}"
	is_app_entitlement_active = true
}

resource "auth0_resource_server" "my_resource_server" {
	name       = "Acceptance Test - {{.testName}}"
	identifier = "https://api.{{.testName}}.com/"
}

resource "auth0_client" "my_test_client" {
	name = "test-client-{{.testName}}"
}

resource "auth0_client_grant" "my_client_grant" {
	depends_on              = [ auth0_resource_server.my_resource_server, auth0_client.my_test_client ]
	client_id               = auth0_client.my_test_client.id
	audience                = auth0_resource_server.my_resource_server.identifier
	scopes                  = []
	allow_any_organization  = true
	organization_usage      = "allow"
}

resource "auth0_organization_client_grant" "my_organization_client_grant" {
	depends_on      = [ auth0_client_grant.my_client_grant ]
	organization_id = auth0_organization.my_organization.id
	grant_id        = auth0_client_grant.my_client_grant.id
}

data "auth0_client_grant_organizations" "my_client_grant_organizations" {
	depends_on      = [ auth0_organization_client_grant.my_organization_client_grant ]
	client_grant_id = auth0_client_grant.my_client_grant.id
}
`

func TestAccDataSourceClientGrantOrganizations(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceClientGrantOrganizationsConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_client_grant_organizations.my_client_grant_organizations", "organizations.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_client_grant_organizations.my_client_grant_organizations", "organizations.0.organization_id"),
					resource.TestCheckResourceAttrSet("data.auth0_client_grant_organizations.my_client_grant_organizations", "organizations.0.name"),
					resource.TestCheckResourceAttr("data.auth0_client_grant_organizations.my_client_grant_organizations", "organizations.0.is_app_entitlement_active", "true"),
					resource.TestCheckResourceAttr("data.auth0_client_grant_organizations.my_client_grant_organizations", "organizations.0.third_party_client_access", "block"),
				),
			},
		},
	})
}
