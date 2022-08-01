package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/template"
)

const testAccDataClientConfigByName = `
%v
data auth0_client test {
  name = "Acceptance Test - {{.testName}}"
}
`

const testAccDataClientConfigByID = `
%v
data auth0_client test {
  client_id = auth0_client.my_client.client_id
}
`

func TestAccDataClientByName(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories:         testProviders(httpRecorder),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
				), // check that the client got created correctly before using the data source
			},
			{
				Config: template.ParseTestName(fmt.Sprintf(testAccDataClientConfigByName, testAccClientConfig), t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "client_id"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "signing_keys.#", "1"), // checks that signing_keys is set, and it includes 1 element
					resource.TestCheckResourceAttr("data.auth0_client.test", "name", fmt.Sprintf("Acceptance Test - %v", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_client.test", "app_type", "non_interactive"), // Arbitrary property selection
					resource.TestCheckNoResourceAttr("data.auth0_client.test", "client_secret_rotation_trigger"),
				),
			},
		},
	})
}

func TestAccDataClientById(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories:         testProviders(httpRecorder),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %v", t.Name())),
				), // check that the client got created correctly before using the data source
			},
			{
				Config: template.ParseTestName(fmt.Sprintf(testAccDataClientConfigByID, testAccClientConfig), t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "id"),
					resource.TestCheckResourceAttrSet("data.auth0_client.test", "name"),
					resource.TestCheckResourceAttr("data.auth0_client.test", "signing_keys.#", "1"), // checks that signing_keys is set, and it includes 1 element
					resource.TestCheckNoResourceAttr("data.auth0_client.test", "client_secret_rotation_trigger"),
				),
			},
		},
	})
}
