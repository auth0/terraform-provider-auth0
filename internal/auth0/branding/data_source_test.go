package branding_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceBranding = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "auth.terraform-provider-auth0.com"
	type = "auth0_managed_certs"
}

resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	logo_url = "https://mycompany.org/v2/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"

	colors {
		primary = "#0059d6"
		page_background = "#000000"
	}

	font {
		url = "https://mycompany.org/font/myfont.ttf"
	}

	universal_login {
		body = "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"
	}
}

data "auth0_branding" "test" {
	depends_on = [ auth0_branding.my_brand ]
}
`

func TestAccDataSourceBranding(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceBranding,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_branding.test", "logo_url", "https://mycompany.org/v2/logo.png"),
					resource.TestCheckResourceAttr("data.auth0_branding.test", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("data.auth0_branding.test", "colors.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding.test", "colors.0.primary", "#0059d6"),
					resource.TestCheckResourceAttr("data.auth0_branding.test", "colors.0.page_background", "#000000"),
					resource.TestCheckResourceAttr("data.auth0_branding.test", "font.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding.test", "font.0.url", "https://mycompany.org/font/myfont.ttf"),
					resource.TestCheckResourceAttr("data.auth0_branding.test", "universal_login.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_branding.test", "universal_login.0.body", "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"),
				),
			},
		},
	})
}
