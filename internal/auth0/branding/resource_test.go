package branding_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenACustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "auth.terraform-provider-auth0.com"
	type   = "auth0_managed_certs"
}
`

const testAccTenantDisallowsUniversalLoginCustomizationWhenNoCustomDomainSet = `
resource "auth0_branding" "my_custom_domain" {
	logo_url    = "https://mycompany.org/v1/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"

	universal_login {
		body = "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"
	}
}
`

const testAccBrandingConfigCreate = testAccGivenACustomDomain + `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	logo_url    = "https://mycompany.org/v1/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"
}
`

const testAccBrandingConfigUpdateAllFields = testAccGivenACustomDomain + `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	logo_url    = "https://mycompany.org/v2/logo.png"
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

const testAccBrandingConfigThrowsAValidationErrorIfUniversalLoginBodyIsEmpty = testAccGivenACustomDomain + `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	logo_url    = "https://mycompany.org/v3/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"

	colors {
		primary = "#0059d6"
	}

	font {
		url = "https://mycompany.org/font/myfont.ttf"
	}

	universal_login {
		# Setting this to an empty string should trigger
		# a validation error as the API doesn't allow it.
		body = ""
	}
}
`

const testAccBrandingConfigRemovesUniversalLoginTemplate = testAccGivenACustomDomain + `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	logo_url    = "https://mycompany.org/v1/logo.png"
	favicon_url = "https://mycompany.org/favicon.ico"
}
`

const testAccBrandingConfigWithOnlyUniversalLogin = testAccGivenACustomDomain + `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	universal_login {
		body = "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"
	}
}
`

const testAccBrandingConfigReset = testAccGivenACustomDomain + `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]
}
`

func TestAccBranding(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccTenantDisallowsUniversalLoginCustomizationWhenNoCustomDomainSet,
				ExpectError: regexp.MustCompile(
					"managing the Universal Login body through the 'auth0_branding' resource " +
						"requires at least one custom domain to be configured for the tenant",
				),
			},
			{
				Config: testAccBrandingConfigCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v1/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.#", "0"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.#", "0"),
				),
			},
			{
				Config: testAccBrandingConfigUpdateAllFields,
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
				Config:      testAccBrandingConfigThrowsAValidationErrorIfUniversalLoginBodyIsEmpty,
				ExpectError: regexp.MustCompile("expected \"universal_login.0.body\" to contain a single auth0:head tag and at least one auth0:widget tag"),
			},
			{
				Config: testAccBrandingConfigRemovesUniversalLoginTemplate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v1/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.#", "0"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.#", "0"),
				),
			},
			{
				Config: testAccBrandingConfigWithOnlyUniversalLogin,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.0.body", "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"),
				),
			},
			{
				Config: testAccBrandingConfigReset,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "logo_url", "https://mycompany.org/v1/logo.png"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "favicon_url", "https://mycompany.org/favicon.ico"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "colors.#", "1"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "font.#", "0"),
					resource.TestCheckResourceAttr("auth0_branding.my_brand", "universal_login.#", "0"),
				),
			},
		},
	})
}
