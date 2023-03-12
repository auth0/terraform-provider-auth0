package client_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

const testAccClientGrantAuxConfig = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Client Grant - {{.testName}}"
	custom_login_page_on = true
	is_first_party = true
}

resource "auth0_resource_server" "my_resource_server" {
	name = "Acceptance Test - Client Grant - {{.testName}}"
	identifier = "https://uat.tf.terraform-provider-auth0.com/client-grant/{{.testName}}"
	scopes {
		value = "create:foo"
		description = "Create foos"
	}
	scopes {
		value = "create:bar"
		description = "Create bars"
	}
}
`

const testAccClientGrantConfigCreate = testAccClientGrantAuxConfig + `
resource "auth0_client_grant" "my_client_grant" {
	client_id = "${auth0_client.my_client.id}"
	audience = "${auth0_resource_server.my_resource_server.identifier}"
	scope = [ ]
}
`

const testAccClientGrantConfigUpdate = testAccClientGrantAuxConfig + `
resource "auth0_client_grant" "my_client_grant" {
	client_id = "${auth0_client.my_client.id}"
	audience = "${auth0_resource_server.my_resource_server.identifier}"
	scope = [ "create:foo" ]
}
`

const testAccClientGrantConfigUpdateAgain = testAccClientGrantAuxConfig + `
resource "auth0_client_grant" "my_client_grant" {
	client_id = "${auth0_client.my_client.id}"
	audience = "${auth0_resource_server.my_resource_server.identifier}"
	scope = [ ]
}
`

const testAccClientGrantConfigUpdateChangeClient = testAccClientGrantAuxConfig + `
resource "auth0_client" "my_client_alt" {
	name = "Acceptance Test - Client Grant Alt - {{.testName}}"
	custom_login_page_on = true
	is_first_party = true
}

resource "auth0_client_grant" "my_client_grant" {
	client_id = "${auth0_client.my_client_alt.id}"
	audience = "${auth0_resource_server.my_resource_server.identifier}"
	scope = [ ]
}
`

const testAccAlreadyExistingGrantWillNotConflict = testAccClientGrantAuxConfig + `
resource "auth0_client" "my_client_alt" {
	name = "Acceptance Test - Client Grant Alt - {{.testName}}"
	custom_login_page_on = true
	is_first_party = true
}

resource "auth0_client_grant" "my_client_grant" {
	client_id = "${auth0_client.my_client_alt.id}"
	audience = "${auth0_resource_server.my_resource_server.identifier}"
	scope = [ ]
}

resource "auth0_client_grant" "no_conflict_client_grant" {
	depends_on = [ auth0_client_grant.my_client_grant ]

	client_id = "${auth0_client.my_client_alt.id}"
	audience = "${auth0_resource_server.my_resource_server.identifier}"
	scope = [ ]
}
`

func TestAccClientGrant(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientGrantConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "audience", fmt.Sprintf("https://uat.tf.terraform-provider-auth0.com/client-grant/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scope.#", "0"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientGrantConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scope.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scope.0", "create:foo"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientGrantConfigUpdateAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scope.#", "0"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientGrantConfigUpdateChangeClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scope.#", "0"),
				),
			},
			{
				Config: template.ParseTestName(testAccAlreadyExistingGrantWillNotConflict, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.no_conflict_client_grant", "scope.#", "0"),
				),
			},
		},
	})
}
