package client_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

const testAccGivenAClient = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - {{.testName}}"
	app_type = "non_interactive"
}
`

const testAccDataClientConfigByName = `
data "auth0_client" "test" {
	depends_on = [ auth0_client.my_client ]

	name = "Acceptance Test - {{.testName}}"
}
`

const testAccDataClientConfigByID = `
data "auth0_client" "test" {
	client_id = auth0_client.my_client.client_id
}
`

func TestAccDataClientByName(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccGivenAClient+testAccDataClientConfigByName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "client_id"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "signing_keys.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "name", fmt.Sprintf("Acceptance Test - %v", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_client.test", "app_type", "non_interactive"),
					resource.TestCheckNoResourceAttr("data.auth0_client.test", "client_secret_rotation_trigger"),
				),
			},
		},
	})
}

func TestAccDataClientById(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccGivenAClient+testAccDataClientConfigByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "id"),
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "name"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "signing_keys.#", "1"),
					resource.TestCheckNoResourceAttr("data.auth0_client.test", "client_secret_rotation_trigger"),
				),
			},
		},
	})
}
