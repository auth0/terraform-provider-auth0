package prompt_test

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPromptScreenPartialCreate = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_screen_partial" "login_passwordless_email_code" {
  	depends_on = [ auth0_branding.my_brand ]
  	prompt_type = "login-passwordless"
	screen_name = "login-passwordless-email-code"
	insertion_points {
		form_content_start = "<div>Form Content Start</div>"
		form_content_end = "<div>Form Content End</div>"
	}
}
`
const testAccPromptScreenPartialCreate2 = `
resource "auth0_prompt_screen_partial" "login_passwordless_sms_otp" {
	  	depends_on = [ auth0_branding.my_brand ]
	  	prompt_type = "login-passwordless"
		screen_name = "login-passwordless-sms-otp"
		insertion_points {
			form_content_start = "<div>Form Content Start</div>"
			form_content_end = "<div>Form Content End</div>"
		}
}
`
const testAccPromptScreenPartialUpdate = testAccPromptScreenPartialCreate + testAccPromptScreenPartialCreate2

const testAccPromptScreenPartialDelete = testAccGivenACustomDomain + testGivenABrandingTemplate + testAccPromptScreenPartialCreate2

const testAccPromptScreenPartialData = `
data "auth0_prompt_screen_partials" "login_passwordless" {
	prompt_type = "login-passwordless"
}
`

func TestAccPromptScreenPartial(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptScreenPartialCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_email_code", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_email_code", "screen_name", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_email_code", "insertion_points.0.form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_email_code", "insertion_points.0.form_content_end", "<div>Form Content End</div>"),
				),
			},
			{
				Config: testAccPromptScreenPartialUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_email_code", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_email_code", "screen_name", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_email_code", "insertion_points.0.form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_email_code", "insertion_points.0.form_content_end", "<div>Form Content End</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_sms_otp", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_sms_otp", "screen_name", "login-passwordless-sms-otp"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_sms_otp", "insertion_points.0.form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_partial.login_passwordless_sms_otp", "insertion_points.0.form_content_end", "<div>Form Content End</div>"),
				),
			},
			{
				Config: testAccPromptScreenPartialDelete,
			},
			{
				Config: testAccPromptScreenPartialData,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.login_passwordless", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.login_passwordless", "screen_partials.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.login_passwordless", "screen_partials.0.screen_name", "login-passwordless-sms-otp"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.login_passwordless", "screen_partials.0.insertion_points.0.form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_partials.login_passwordless", "screen_partials.0.insertion_points.0.form_content_end", "<div>Form Content End</div>"),
					resource.TestCheckNoResourceAttr("data.auth0_prompt_screen_partials.login_passwordless", "screen_partials.1"),
				),
			},
		},
	})
}
