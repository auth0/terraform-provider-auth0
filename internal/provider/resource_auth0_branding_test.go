package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccBranding(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccBrandingConfigCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v1/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
				),
			},
			{
				Config: testAccBrandingConfigUpdateColors,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v1/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.primary", "#0059d6"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.page_background", "#000000"),
				),
			},
			{
				Config: testAccBrandingConfigUpdateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v2/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.primary", "#ffa629"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.page_background", "#ffffff"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.0.url", "https://mycompany.org/font/myfont.ttf"),
				),
			},
		},
	})
}

const testAccBrandingConfigCreate = `
resource "auth0_branding" "my_brand" {
	logo_url = "https://mycompany.org/v1/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"
}
`

const testAccBrandingConfigUpdateColors = `
resource "auth0_branding" "my_brand" {
	logo_url = "https://mycompany.org/v1/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"
	colors {
		primary = "#0059d6"
		page_background = "#000000"
	}
}
`

const testAccBrandingConfigUpdateFull = `
resource "auth0_branding" "my_brand" {
	logo_url = "https://mycompany.org/v2/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"
	colors {
		primary = "#ffa629"
		page_background = "#ffffff"
	}
	font {
		url = "https://mycompany.org/font/myfont.ttf"
	}
}
`
