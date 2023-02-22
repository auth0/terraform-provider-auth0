package branding_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccTenantAllowsUniversalLoginCustomization = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "auth.terraform-provider-auth0.com"
	type = "auth0_managed_certs"
}

`

const testAccTenantDisallowsUniversalLoginCustomization = `
resource "auth0_branding" "my_custom_domain" {
	logo_url = "https://mycompany.org/v1/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"

	universal_login {
		body = "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"
	}
}
`

const testAccBrandingConfigCreate = `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	logo_url = "https://mycompany.org/v1/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"
}
`

const testAccBrandingConfigUpdateAllFields = `
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
`

const testAccBrandingConfigUpdateAgain = `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	logo_url = "https://mycompany.org/v3/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"

	colors {
		primary = "#0059d6"
	}

	font {
		url = "https://mycompany.org/font/myfont.ttf"
	}

	universal_login {
		# Setting this to an empty string should
		# not trigger any API call, so the value
		# stays the same as the previous scenario.
		body = ""
	}
}
`

const testAccBrandingConfigReset = `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	logo_url = "https://mycompany.org/v1/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"
}
`

func TestAccBranding_WithNoCustomDomainsSet(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDisallowsUniversalLoginCustomization,
				ExpectError: regexp.MustCompile(
					"managing the universal login body through the 'auth0_branding' resource " +
						"requires at least one custom domain to be configured for the tenant",
				),
			},
		},
	})
}

func TestAccBranding(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccTenantAllowsUniversalLoginCustomization + testAccBrandingConfigCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v1/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.#", "0"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.#", "0"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.#", "0"),
				),
			},
			{
				Config: testAccTenantAllowsUniversalLoginCustomization + testAccBrandingConfigUpdateAllFields,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v2/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.primary", "#0059d6"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.page_background", "#000000"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.0.url", "https://mycompany.org/font/myfont.ttf"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.0.body", "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"),
				),
			},
			{
				Config: testAccTenantAllowsUniversalLoginCustomization + testAccBrandingConfigUpdateAgain,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v3/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.primary", "#0059d6"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.0.page_background", "#000000"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.0.url", "https://mycompany.org/font/myfont.ttf"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.0.body", "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"),
				),
			},
			{
				Config: testAccTenantAllowsUniversalLoginCustomization + testAccBrandingConfigReset,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v1/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.#", "0"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.#", "0"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.#", "0"),
				),
			},
		},
	})
}
