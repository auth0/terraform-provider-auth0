package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
)

const testAccDataGlobalClientConfig = `
%v
data auth0_global_client global {
}
`

func TestAccDataGlobalClient(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalClientConfigWithCustomLogin,
			},
			{
				Config: fmt.Sprintf(testAccDataGlobalClientConfig, testAccGlobalClientConfigWithCustomLogin),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_global_client.global", "custom_login_page", "<html>TEST123</html>"),
					resource.TestCheckResourceAttr("data.auth0_global_client.global", "custom_login_page_on", "true"),
					resource.TestCheckResourceAttrSet("data.auth0_global_client.global", "client_id"),
					resource.TestCheckResourceAttr("data.auth0_global_client.global", "app_type", ""),
					resource.TestCheckResourceAttr("data.auth0_global_client.global", "name", "All Applications"),
				),
			},
		},
	})
}
