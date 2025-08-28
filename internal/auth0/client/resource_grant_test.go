package client_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAClientAndAResourceServerWithScopes = `
resource "auth0_client" "my_client" {
	name                 = "Acceptance Test - Client Grant - {{.testName}}"
	custom_login_page_on = true
	is_first_party       = true
}

resource "auth0_resource_server" "my_resource_server" {
	name       = "Acceptance Test - Client Grant - {{.testName}}"
	identifier = "https://uat.tf.terraform-provider-auth0.com/client-grant/{{.testName}}"
}

resource "auth0_resource_server_scopes" "my_api_scopes" {
	depends_on = [ auth0_resource_server.my_resource_server ]

	resource_server_identifier = auth0_resource_server.my_resource_server.identifier

	scopes {
		name        = "create:foo"
		description = "Create foos"
	}

	scopes {
		name        = "create:bar"
		description = "Create bars"
	}
}
`

const testAccClientGrantConfigCreate = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	client_id = auth0_client.my_client.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = ["create:foo" ]
	subject_type = "user"
	authorization_details_types = ["payment","shipping"]
}
`

const testAccClientGrantConfigUpdate = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	client_id = auth0_client.my_client.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ "create:foo" ]
}
`

const testAccClientGrantConfigUpdateAgain = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	client_id = auth0_client.my_client.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ ]
}
`

const testAccClientGrantConfigUpdateChangeClient = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client" "my_client_alt" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	name                 = "Acceptance Test - Client Grant Alt - {{.testName}}"
	custom_login_page_on = true
	is_first_party       = true
}

resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_client.my_client_alt ]

	client_id = auth0_client.my_client_alt.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ ]
}
`

const testAccAlreadyExistingGrantWillNotConflict = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client" "my_client_alt" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	name                 = "Acceptance Test - Client Grant Alt - {{.testName}}"
	custom_login_page_on = true
	is_first_party       = true
}

resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_client.my_client_alt ]

	client_id = auth0_client.my_client_alt.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ ]
}

resource "auth0_client_grant" "no_conflict_client_grant" {
	depends_on = [ auth0_client_grant.my_client_grant ]

	client_id = auth0_client.my_client_alt.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ ]
}
`

func TestAccClientGrant(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "audience", fmt.Sprintf("https://uat.tf.terraform-provider-auth0.com/client-grant/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "subject_type", "user"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "authorization_details_types.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.0", "create:foo"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigUpdateAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigUpdateChangeClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccAlreadyExistingGrantWillNotConflict, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.no_conflict_client_grant", "scopes.#", "0"),
				),
			},
		},
	})
}
