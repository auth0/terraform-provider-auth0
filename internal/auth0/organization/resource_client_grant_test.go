package organization_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAssociateOrganizationClientGrant = `
resource "auth0_organization" "my_organization" {
    name         = "test-org-acceptance-testing"
    display_name = "Test Org Acceptance Testing"
}

resource "auth0_resource_server" "new_resource_server" {
    name       = "Example API"
    identifier = "https://api.travel00123.com/"
}


resource "auth0_client" "my_test_client"{
    depends_on = [ auth0_organization.my_organization, auth0_client.my_test_client ]
    name = "test_client"
    organization_usage = "allow"
    default_organization {
        organization_id = auth0_organization.my_organization.id
        flows = ["client_credentials"]
    }
}

resource "auth0_client_grant" "my_client_grant" {
    depends_on = [ auth0_resource_server.new_resource_server, auth0_client.my_test_client ]
    client_id = auth0_client.my_test_client.id
    audience = auth0_resource_server.new_resource_server.identifier
    scopes = ["create:organization_client_grants","create:resource"]
    allow_any_organization = true
    organization_usage = "allow"
}


resource "auth0_organization_client_grant" "associate_org_client_grant"{
    depends_on = [ auth0_client_grant.my_client_grant ]
    organization_id = auth0_organization.my_organization.id
    grant_id = auth0_client_grant.my_client_grant.id
}

data "auth0_organization" "retrieve_org_data" {
	depends_on = [ auth0_organization_client_grant.associate_org_client_grant ]
	organization_id = auth0_organization.my_organization.id
}

`

func TestOrganizationClientGrant(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAssociateOrganizationClientGrant, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_organization_client_grant.associate_org_client_grant", "organization_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_client_grant.associate_org_client_grant", "grant_id"),
					resource.TestCheckResourceAttrSet("data.auth0_organization.retrieve_org_data", "client_grants.0"),
				),
			},
		},
	})
}
