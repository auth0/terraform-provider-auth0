package auth0

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccBranding(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"auth0": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: TestAccBrandingConfigCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v1/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.primary", "#0059d6"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.page_background", "#000000"),
				),
			},
			{
				Config: TestAccBrandingConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v2/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.primary", "#ffa629"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.page_background", "#ffffff"),
				),
			},
		},
	})
}

const TestAccBrandingConfigCreate = `
resource "auth0_branding" "my_brand" {
	logo_url = "https://mycompany.org/v1/logo.png"
	colors {
		primary = "#0059d6"
		page_background = "#000000"
	}
}
`

const TestAccBrandingConfigUpdate = `
resource "auth0_branding" "my_brand" {
	logo_url = "https://mycompany.org/v2/logo.png"
	colors {
		primary = "#ffa629"
		page_background = "#ffffff"
	}
}
`
