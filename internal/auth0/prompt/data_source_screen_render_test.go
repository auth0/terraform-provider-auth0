package prompt_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPromptScreenRenderWithoutScreens = testAccGivenACustomDomain + testGivenABrandingTemplate + `
data "auth0_prompt_screen_renderer" "prompt_screen_render" {
	  prompt_type = "login-passwordless"
}
`
const testAccPromptScreenRenderInvalid = `
data "auth0_prompt_screen_renderer" "prompt_screen_render" {
	  prompt_type = "login-xxxxx"
      screen_name = "login-passwordless-email-code"
}
`

const testAccPromptScreenRenderData = `
resource "auth0_prompt_screen_renderer" "prompt_screen_render" {
	  prompt_type = "login-passwordless"
	  screen_name = "login-passwordless-email-code"
      rendering_mode = "advanced"
      context_configuration = [
        "branding.settings",
        "branding.themes.default",
        "client.logo_uri",
        "client.description",
        "organization.display_name",
        "organization.branding",
        "screen.texts",
        "tenant.name",
        "tenant.friendly_name",
        "tenant.enabled_locales",
        "untrusted_data.submitted_form_data",
        "untrusted_data.authorization_params.ui_locales",
        "untrusted_data.authorization_params.login_hint",
        "untrusted_data.authorization_params.screen_hint"
    ]
    head_tags = jsonencode([
       {
           attributes: {
               "async": true,
               "defer": true,
               "integrity": [
                   "sha512-v2CJ7UaYy4JwqLDIrZUI/4hqeoQieOmAZNXBeQyjo21dadnwR+8ZaIJVT8EE2iyI61OV8e6M8PP2/4hpQINQ/g=="
               ],
               "src": "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.1/jquery.min.js"
           },
           tag: "script"
       }
    ])


}

data "auth0_prompt_screen_renderer" "prompt_screen_render" {
  	  depends_on = [ auth0_prompt_screen_renderer.prompt_screen_render ]
	  prompt_type = auth0_prompt_screen_renderer.prompt_screen_render.prompt_type
	  screen_name = auth0_prompt_screen_renderer.prompt_screen_render.screen_name
}
`

func TestAccDataPromptScreenRender(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_prompt_screen_renderer" "prompt_screen_render" { }`,
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      testAccPromptScreenRenderWithoutScreens,
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      testAccPromptScreenRenderInvalid,
				ExpectError: regexp.MustCompile("expected prompt_type to be one of"),
			},
			{
				Config: testAccPromptScreenRenderData,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_render", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_render", "screen_name", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_render", "rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_render", "context_configuration.#", "14"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderer.prompt_screen_render", "prompt_type", "login-passwordless"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderer.prompt_screen_render", "screen_name", "login-passwordless-email-code"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderer.prompt_screen_render", "rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderer.prompt_screen_render", "context_configuration.#", "14"),
				),
			},
		},
	})
}
