package auth0

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/template"
)

const testAccOrganizationConnectionGivenAnOrgAndAConnection = `
resource auth0_connection my_connection {
	name = "Acceptance-Test-Connection-First-{{.testName}}"
	strategy = "auth0"
}

resource auth0_organization my_organization {
	depends_on = [auth0_connection.my_connection]
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}
`

const TestAccOrganizationConnectionCreate = testAccOrganizationConnectionGivenAnOrgAndAConnection + `
resource auth0_organization_connection my_org_conn {
	organization_id = auth0_organization.my_organization.id
	connection_id = auth0_connection.my_connection.id
}
`

const TestAccOrganizationConnectionUpdate = testAccOrganizationConnectionGivenAnOrgAndAConnection + `
resource auth0_organization_connection my_org_conn {
	organization_id = auth0_organization.my_organization.id
	connection_id = auth0_connection.my_connection.id
	assign_membership_on_login = true
}
`

func TestAccOrganizationConnection(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(TestAccOrganizationConnectionCreate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "organization_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "connection_id"),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"name",
						"Acceptance-Test-Connection-First-"+strings.ToLower(t.Name()),
					),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"strategy",
						"auth0",
					),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"assign_membership_on_login",
						"false",
					),
				),
			},
			{
				Config: template.ParseTestName(TestAccOrganizationConnectionUpdate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "organization_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "connection_id"),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"name",
						"Acceptance-Test-Connection-First-"+strings.ToLower(t.Name()),
					),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"strategy",
						"auth0",
					),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"assign_membership_on_login",
						"true",
					),
				),
			},
		},
	})
}
