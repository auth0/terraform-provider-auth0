package flow_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAGivenVaultConnection = `
resource "auth0_flow_vault_connection" "my_connection" {
    name = "test-vault-connection"
	app_id = "HTTP"
}
`

const testDataResourceVaultConnectionWithoutID = testAGivenVaultConnection + `
data "auth0_flow_vault_connection" "my_connection" {
	depends_on = [resource.auth0_flow_vault_connection.my_connection]
}`

const testDataResourceVaultConnectionWithInvalidID = testAGivenVaultConnection + `
data "auth0_flow_vault_connection" "my_connection" {
	depends_on = [resource.auth0_flow_vault_connection.my_connection]
	id = "ac_5Cy8N47rNvNixgNBShK8fK"
}`

const testDataResourceVaultConnectionWithValidID = testAGivenVaultConnection + `
data "auth0_flow_vault_connection" "my_connection" {
	depends_on = [resource.auth0_flow_vault_connection.my_connection]
	id = resource.auth0_flow_vault_connection.my_connection.id
}`

func TestAccVaultConnectionDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testDataResourceVaultConnectionWithoutID, t.Name()),
				ExpectError: regexp.MustCompile("The argument \"id\" is required, but no definition was found."),
			},
			{
				Config: acctest.ParseTestName(testDataResourceVaultConnectionWithInvalidID, t.Name()),
				ExpectError: regexp.MustCompile(
					`Error: 404 Not Found`,
				),
			},
			{
				Config: acctest.ParseTestName(testDataResourceVaultConnectionWithValidID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_flow_vault_connection.my_connection", "name"),
				),
			},
		},
	})
}
