package client_test

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

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

func TestAccGlobalClient(t *testing.T) {
	acctest.Test(t, resource.TestCase{
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
