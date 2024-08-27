package prompt_test

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

const testGivenABrandingTemplate = `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	universal_login {
		body = "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"
	}
}
`

const testAccPromptPartialsCreate = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_partials" "prompt_partials" {
  depends_on = [ auth0_branding.my_brand ]

  prompt = "login"

  form_content_start = "<div>Form Content Start</div>"
}
`

const testAccPromptPartialsUpdate = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_partials" "prompt_partials" {
  depends_on = [ auth0_branding.my_brand ]

  prompt = "login"

  form_content_start      = "<div>Updated Form Content Start</div>"
  form_content_end        = "<div>Updated Form Content End</div>"
  form_footer_start       = "<div>Updated Footer Start</div>"
  form_footer_end         = "<div>Updated Footer End</div>"
  secondary_actions_start = "<div>Updated Secondary Actions Start</div>"
  secondary_actions_end   = "<div>Updated Secondary Actions End</div>"
}
`
const testAccPromptPartialsWithScreenName = testAccGivenACustomDomain + testGivenABrandingTemplate + `

resource "auth0_prompt_partials" "prompt_partials_with_screen_name" {
  depends_on = [ auth0_branding.my_brand ]

	prompt = "login-passwordless"
	screen_name = "login-passwordless-email-code"
	form_content_start = "<div>Form Content Start</div>"
}
`
const testAccPromptPartialsWithScreenNameUpdate = testAccGivenACustomDomain + testGivenABrandingTemplate + `

resource "auth0_prompt_partials" "prompt_partials_with_screen_name" {
  depends_on = [ auth0_branding.my_brand ]

	prompt = "login-passwordless"
	screen_name = "login-passwordless-sms-otp"
	form_content_start = "<div>Form Content Start</div>"
}
`

const testAccPromptPartialsWithInvalidScreenName = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_partials" "prompt_partials_with_screen_name" {
  depends_on = [ auth0_branding.my_brand ]

	prompt = "login-passwordless"
	screen_name = "invalid-screen-name"
	form_content_start = "<div>Form Content Start</div>"
}
`

func TestAccPromptPartials(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptPartialsCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_content_start", "<div>Form Content Start</div>"),
				),
			},
			{
				Config: testAccPromptPartialsUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_content_start", "<div>Updated Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_content_end", "<div>Updated Form Content End</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_footer_start", "<div>Updated Footer Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_footer_end", "<div>Updated Footer End</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "secondary_actions_start", "<div>Updated Secondary Actions Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "secondary_actions_end", "<div>Updated Secondary Actions End</div>"),
				),
			},
			{
				Config: testAccPromptPartialsWithScreenName,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials_with_screen_name", "prompt", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials_with_screen_name", "screen_name", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials_with_screen_name", "form_content_start", "<div>Form Content Start</div>"),
				),
			},
			{
				Config: testAccPromptPartialsWithScreenNameUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials_with_screen_name", "prompt", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials_with_screen_name", "screen_name", "login-passwordless-sms-otp"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials_with_screen_name", "form_content_start", "<div>Form Content Start</div>"),
				),
			},
			{
				Config:      testAccPromptPartialsWithInvalidScreenName,
				ExpectError: regexp.MustCompile(`Invalid screen 'invalid-screen-name' in prompt 'login-passwordless'`),
			},
		},
	})
}
