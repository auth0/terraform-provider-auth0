package prompt_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const (
	testAccPromptScreenRenderWithoutSettings = `
resource "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
  prompt_type     = "logout"
  screen_name =  "logout"
}
`

	testAccPromptScreenRenderCreate = `
resource "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
  prompt_type     = "logout"
  screen_name     =  "logout"
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
        "tenant.enabled_locales",
        "untrusted_data.submitted_form_data",
        "untrusted_data.authorization_params.ui_locales",
        "untrusted_data.authorization_params.login_hint",
        "untrusted_data.authorization_params.screen_hint",
        "user.organizations"
  ]
}
`

	testAccPromptScreenRenderUpdate = `
resource "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
  prompt_type     = "logout"
  screen_name     =  "logout"
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
        "tenant.enabled_locales",
  ]
}
`

	testAccPromptScreenRenderDelete = `
data "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
	  prompt_type = "logout"
     screen_name = "logout"
}`
	testAccPromptScreenRenderDataAfterDelete = testAccPromptScreenPartialsDelete + `
data "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
	  prompt_type = "logout"
      screen_name = "logout"
}
`
)

func TestAccPromptScreenSettings(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptScreenRenderWithoutSettings,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "prompt_type", "logout"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "screen_name", "logout"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderer.prompt_screen_renderer", "id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "rendering_mode", "standard"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "context_configuration.#", "0"),
				),
			},
			{
				Config: testAccPromptScreenRenderCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "prompt_type", "logout"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "screen_name", "logout"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderer.prompt_screen_renderer", "id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "context_configuration.#", "15"),
				),
			},
			{
				Config: testAccPromptScreenRenderUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "prompt_type", "logout"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "screen_name", "logout"),
					resource.TestCheckResourceAttrSet("auth0_prompt_screen_renderer.prompt_screen_renderer", "id"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "rendering_mode", "advanced"),
					resource.TestCheckResourceAttr("auth0_prompt_screen_renderer.prompt_screen_renderer", "context_configuration.#", "10"),
				),
			},
			{
				Config: testAccPromptScreenRenderDelete,
			},
			{
				Config: testAccPromptScreenRenderDataAfterDelete,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderer.prompt_screen_renderer", "prompt_type", "logout"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderer.prompt_screen_renderer", "rendering_mode", "standard"),
				),
			},
		},
	})
}
