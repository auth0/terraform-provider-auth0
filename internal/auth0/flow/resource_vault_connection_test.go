package flow_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testVaultConnectionCreateInvalidAppID = `
resource "auth0_flow_vault_connection" "my_vault_connection" {
    name = "test-v-connection"
	app_id = "INVALID"
}
`

const testVaultConnectionCreate = `
resource "auth0_flow_vault_connection" "my_vault_connection" {
    name = "test-v-connection"
	app_id = "HTTP"
}
`

const testVaultConnectionUpdate = `
resource "auth0_flow_vault_connection" "my_vault_connection" {
    name = "updated-test-v-connection"
	app_id = "HTTP"
}
`

func TestAccVaultConnection(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testVaultConnectionCreateInvalidAppID, t.Name()),
				ExpectError: regexp.MustCompile(`app_id to be one of`),
			},
			{
				Config: acctest.ParseTestName(testVaultConnectionCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_flow_vault_connection.my_vault_connection", "name", "test-v-connection"),
					resource.TestCheckResourceAttrSet("auth0_flow_vault_connection.my_vault_connection", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testVaultConnectionUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_flow_vault_connection.my_vault_connection", "name", "updated-test-v-connection"),
					resource.TestCheckResourceAttrSet("auth0_flow_vault_connection.my_vault_connection", "id"),
				),
			},
		},
	})
}
