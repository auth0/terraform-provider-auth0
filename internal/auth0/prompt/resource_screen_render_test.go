package prompt_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const (
	testAccPromptScreenRenderWithoutScreensInfo = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_screen_renderer" "prompt_screen_render" {
	  prompt_type = "login-passwordless"
}
`

	testAccPromptScreenRenderWithoutPromptsInfo = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_screen_renderer" "prompt_screen_render" {
	  screen_name = "login-passwordless"
}
`

	testAccPromptScreenRenderInvalidInfo = `
resource "auth0_prompt_screen_renderer" "prompt_screen_render" {
	  prompt_type = "login-xxxxx"
      screen_name = "login-passwordless-email-code"
}
`

	testAccPromptScreenRenderWithoutSettings = `
resource "auth0_prompt_screen_renderer" "login-id" {
  prompt_type     = "login-id"
  screen_name =  "login-id"
}
`

	testAccPromptScreenRenderCreate = testAccPromptScreenRenderWithoutSettings + `
resource "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
  prompt_type     = "login-password"
  screen_name     =  "login-password"
  rendering_mode = "advanced"

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

	filters {
        match_type = "includes_any"

        clients {
            id = "fCYih2bqL9fCVAi37DOnx9OxKbTkAsUs"

            metadata = {
                key1 = "value1"
            }
        }
    }

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
		"untrusted_data.authorization_params.login_hint",
		"untrusted_data.authorization_params.screen_hint",
		"untrusted_data.authorization_params.ui_locales",
		"organization.metadata.key",
  ]
}
`

	testAccPromptScreenRenderUpdate = `
resource "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
  prompt_type     = "login-password"
  screen_name     =  "login-password"
  rendering_mode = "advanced"

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
		"untrusted_data.authorization_params.login_hint",
		"untrusted_data.authorization_params.ui_locales",
  ]
}
`

	testAccPromptScreenRenderDelete = testAccPromptScreenRenderWithoutSettings

	testAccPromptScreenRenderDataAfterDelete = `
data "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
	  prompt_type = "login-password"
      screen_name = "login-password"
}
`
)

func TestAccPromptScreenSettings(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccPromptScreenRenderWithoutScreensInfo,
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      testAccPromptScreenRenderWithoutPromptsInfo,
				ExpectError: regexp.MustCompile("Error: Missing required argument"),
			},
			{
				Config:      testAccPromptScreenRenderInvalidInfo,
				ExpectError: regexp.MustCompile("expected prompt_type to be one of"),
			},
			{
				Config: testAccPromptScreenRenderWithoutSettings,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.login-id", "prompt_type", "login-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.login-id", "screen_name", "login-id"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderer.login-id", "id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.login-id", "rendering_mode", "standard"),
				),
			},
			{
				Config: testAccPromptScreenRenderCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "prompt_type", "login-password"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "screen_name", "login-password"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderer.prompt_screen_renderer", "id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "context_configuration.#", "15"),
				),
			},
			{
				Config: testAccPromptScreenRenderUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "prompt_type", "login-password"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "screen_name", "login-password"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderer.prompt_screen_renderer", "id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "context_configuration.#", "11"),
				),
			},
			{
				Config: testAccPromptScreenRenderDelete,
			},
			{
				Config: testAccPromptScreenRenderDataAfterDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderer.prompt_screen_renderer", "prompt_type", "login-password"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderer.prompt_screen_renderer", "rendering_mode", "standard"),
				),
			},
		},
	})
}

func TestNew(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptScreenRenderCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.login-id", "prompt_type", "login-id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.login-id", "screen_name", "login-id"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderer.login-id", "id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.login-id", "rendering_mode", "standard"),
				),
			},
		},
	})
}
