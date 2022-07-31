package provider

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccGlobalClient(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClientConfigWithCustomLogin,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_global_client.global", "client_id"),
					resource.TestCheckResourceAttrSet("auth0_global_client.global", "client_secret"),
					resource.TestCheckResourceAttr("auth0_global_client.global", "custom_login_page", "<html>TEST123</html>"),
					resource.TestCheckResourceAttr("auth0_global_client.global", "custom_login_page_on", "true"),
				),
			},
			{
				Config: testAccGlobalClientConfigEmpty,
				Check: resource.ComposeTestCheckFunc(
					func(state *terraform.State) error {
						for _, m := range state.Modules {
							if len(m.Resources) > 0 {
								if _, ok := m.Resources["auth0_global_client.global"]; ok {
									return errors.New("auth0_global_client.global exists when it should have been removed")
								}
							}
						}
						return nil
					},
				),
			},
			{
				Config: testAccGlobalClientConfigDefault,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_global_client.global", "custom_login_page", "<html>TEST123</html>"),
					resource.TestCheckResourceAttr("auth0_global_client.global", "custom_login_page_on", "true"),
				),
			},

			{
				Config: testAccGlobalClientConfigNoCustomLogin,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_global_client.global", "custom_login_page_on", "false"),
				),
			},
		},
	})
}

const testAccGlobalClientConfigEmpty = `
`

const testAccGlobalClientConfigDefault = `
resource "auth0_global_client" "global" {
}
`

const testAccGlobalClientConfigWithCustomLogin = `
resource "auth0_global_client" "global" {
    custom_login_page = "<html>TEST123</html>"
    custom_login_page_on = true
}
`

const testAccGlobalClientConfigNoCustomLogin = `
resource "auth0_global_client" "global" {
    custom_login_page_on = false
}
`
