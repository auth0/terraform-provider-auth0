package prompt_test

import (
	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
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

const testAccPromptScreenPartialsCreate = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_screen_partials" "prompt_screen_partials" {
  depends_on = [ auth0_branding.my_brand ]
  prompt_type = "login-passwordless"
  screen_partials {
	screen_name = "login-passwordless-email-code"
	insertion_points {
		form_content_start = "<div>Form Content Start</div>"
		form_content_end = "<div>Form Content End</div>"
	}
  }

  screen_partials {
	screen_name = "login-passwordless-sms-otp"
	insertion_points {
		form_content_start = "<div>Form Content Start</div>"
		form_content_end = "<div>Form Content End</div>"
	}
  }

}
`

const testAccPromptScreenPartialsUpdate = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_screen_partials" "prompt_screen_partials" {
  depends_on = [ auth0_branding.my_brand ]
  prompt_type = "login-passwordless"
  screen_partials {
	screen_name = "login-passwordless-email-code"
	insertion_points {
		form_content_start = "<div>Form Content Start</div>"
		form_content_end = "<div>Form Content End</div>"
	}
  }
}
`

const testAccPromptScreenPartialsDelete = testAccGivenACustomDomain + testGivenABrandingTemplate + `
data "auth0_prompt_screen_partials" "prompt_screen_partials" {
  depends_on = [ auth0_branding.my_brand ]
  prompt_type = "login-passwordless"
}
`

func TestAccPromptScreenPartials(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptScreenPartialsCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.screen_name", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.insertion_points.0.form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.insertion_points.0.form_content_end", "<div>Form Content End</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.1.screen_name", "login-passwordless-sms-otp"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.1.insertion_points.0.form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.1.insertion_points.0.form_content_end", "<div>Form Content End</div>"),
				),
			},
			{
				Config: testAccPromptScreenPartialsUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.screen_name", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.insertion_points.0.form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.insertion_points.0.form_content_end", "<div>Form Content End</div>"),
					resource.TestCheckNoResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.1"),
				),
			},
			{
				Config: testAccPromptScreenPartialsDelete,
			},
		},
	})
}
