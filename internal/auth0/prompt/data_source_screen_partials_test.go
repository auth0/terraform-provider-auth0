package prompt_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPromptScreenPartialsInvalid = `
data "auth0_prompt_screen_partials" "prompt_screen_partials" {
	  prompt_type = "login-xxxxx"
}
`

const testAccPromptScreenPartialsWithoutScreens = testAccGivenACustomDomain + testGivenABrandingTemplate + `
data "auth0_prompt_screen_partials" "prompt_screen_partials" {
	  prompt_type = "login-passwordless"
}
`

const testAccPromptScreenPartialsData = testAccGivenACustomDomain + testGivenABrandingTemplate + `
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

data "auth0_prompt_screen_partials" "prompt_screen_partials" {
  	  depends_on = [ auth0_prompt_screen_partials.prompt_screen_partials ]
	  prompt_type = auth0_prompt_screen_partials.prompt_screen_partials.prompt_type
}
`

func TestAccDataPromptScreenPartials(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_prompt_screen_partials" "prompt_screen_partials" { }`,
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      testAccPromptScreenPartialsInvalid,
				ExpectError: regexp.MustCompile("expected prompt_type to be one of"),
			},
			{
				Config: testAccPromptScreenPartialsWithoutScreens,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.prompt_screen_partials", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.#", "0"),
				),
			},
			{
				Config: testAccPromptScreenPartialsData,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.screen_name", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.insertion_points.0.form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.insertion_points.0.form_content_end", "<div>Form Content End</div>"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.prompt_screen_partials", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.screen_name", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.insertion_points.0.form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.prompt_screen_partials", "screen_partials.0.insertion_points.0.form_content_end", "<div>Form Content End</div>"),
				),
			},
		},
	})
}
