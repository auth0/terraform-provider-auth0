package auth0

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/random"
)

const testAccDataClientConfigByName = `
%v
data auth0_client test {
  name = "Acceptance Test - {{.random}}"
}
`

const testAccDataClientConfigById = `
%v
data auth0_client test {
  client_id = auth0_client.my_client.client_id
}
`

func TestAccDataClientByName(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"auth0": Provider(),
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: random.Template(testAccClientConfig, rand), // must initialize resource before reading with data source
			},
			{
				Config: random.Template(fmt.Sprintf(testAccDataClientConfigByName, testAccClientConfig), rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "client_id"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "signing_keys.#", "1"), // checks that signing_keys is set, and it includes 1 element
					resource.TestCheckResourceAttr("data.auth0_client.test", "name", fmt.Sprintf("Acceptance Test - %v", rand)),
					resource.TestCheckResourceAttr("data.auth0_client.test", "app_type", "non_interactive"), // Arbitrary property selection
					resource.TestCheckNoResourceAttr("data.auth0_client.test", "client_secret_rotation_trigger"),
					resource.TestCheckNoResourceAttr("data.auth0_client.test", "client_secret"),
				),
			},
		},
	})
}

func TestAccDataClientById(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"auth0": Provider(),
		},
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: random.Template(testAccClientConfig, rand),
			},
			{
				Config: random.Template(fmt.Sprintf(testAccDataClientConfigById, testAccClientConfig), rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "id"),
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "name"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "signing_keys.#", "1"), // checks that signing_keys is set, and it includes 1 element
					resource.TestCheckNoResourceAttr("data.auth0_client.test", "client_secret_rotation_trigger"),
					resource.TestCheckNoResourceAttr("data.auth0_client.test", "client_secret"),
				),
			},
		},
	})
}
