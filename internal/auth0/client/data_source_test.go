package client_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAClient = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - {{.testName}}"
	app_type = "non_interactive"
	description = "Description for {{.testName}}"
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
	depends_on = [ auth0_client.my_client ]

	client_id = auth0_client.my_client.client_id
}
`

func TestAccDataClientByName(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccGivenAClient, t.Name()),
			},
			{
				Config: acctest.ParseTestName(testAccGivenAClient+testAccDataClientConfigByName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "client_id"),
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "client_secret"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "name", fmt.Sprintf("Acceptance Test - %v", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_client.test", "app_type", "non_interactive"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "description", "Description for TestAccDataClientByName"),
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
				Config: acctest.ParseTestName(testAccGivenAClient+testAccDataClientConfigByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "name", fmt.Sprintf("Acceptance Test - %v", t.Name())),
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "client_secret"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "app_type", "non_interactive"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "description", "Description for TestAccDataClientById"),
				),
			},
		},
	})
}

const testAccDataSourceClientNonexistentID = `
data "auth0_client" "test" {
	client_id = "this-client-does-not-exist"
}
`

func TestAccDataSourceClientNonexistentID(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceClientNonexistentID, t.Name()),
				ExpectError: regexp.MustCompile(
					` 404 Not Found: The client does not exist`,
				),
			},
		},
	})
}
