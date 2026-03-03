package prompt_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPromptRenderingsDataSourceConfig = `
resource "auth0_prompt_screen_renderer" "login_prompt" {
  prompt_type    = "login-passwordless"
  screen_name    = "login-passwordless-email-code"
  rendering_mode = "advanced"
  context_configuration = [
    "branding.settings",
    "branding.themes.default",
    "client.logo_uri",
    "screen.texts",
    "tenant.name"
  ]
}

resource "auth0_prompt_screen_renderer" "signup_prompt" {
  prompt_type    = "signup-id"
  screen_name    = "signup-id"
  rendering_mode = "standard"
}

data "auth0_prompt_screen_renderers" "all_renderings" {
  depends_on = [
    auth0_prompt_screen_renderer.login_prompt,
    auth0_prompt_screen_renderer.signup_prompt
  ]
}

data "auth0_prompt_screen_renderers" "filtered_by_prompt" {
  depends_on = [
    auth0_prompt_screen_renderer.login_prompt,
    auth0_prompt_screen_renderer.signup_prompt
  ]
  prompt = "login-passwordless"
}

data "auth0_prompt_screen_renderers" "filtered_by_rendering_mode" {
  depends_on = [
    auth0_prompt_screen_renderer.login_prompt,
    auth0_prompt_screen_renderer.signup_prompt
  ]
  rendering_mode = "advanced"
}
`

func TestAccDataPromptRenderings(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					// Check all_renderings data source.
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.all_renderings", "renderings.#"),
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.all_renderings", "id"),

					// Check filtered_by_prompt data source.
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.filtered_by_prompt", "renderings.#"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.filtered_by_prompt", "prompt", "login-passwordless"),
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.filtered_by_prompt", "id"),

					// Check filtered_by_rendering_mode data source.
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.filtered_by_rendering_mode", "renderings.#"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.filtered_by_rendering_mode", "rendering_mode", "advanced"),
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.filtered_by_rendering_mode", "id"),

					// Verify rendering attributes.
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.all_renderings", "renderings.0.prompt"),
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.all_renderings", "renderings.0.screen"),
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.all_renderings", "renderings.0.rendering_mode"),
					// resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.all_renderings", "renderings.0.tenant"),
				),
			},
		},
	})
}

const testAccPromptRenderingsDataSourceWithWildcard = `
data "auth0_prompt_screen_renderers" "with_wildcard" {
  prompt = "login"
}
`

func TestAccDataPromptRenderingsWithWildcard(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsDataSourceWithWildcard,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.with_wildcard", "renderings.#"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.with_wildcard", "prompt", "login"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.with_wildcard", "renderings.#", "5"),
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.with_wildcard", "id"),
				),
			},
		},
	})
}

const testAccPromptRenderingsDataSourceWithMultipleFilters = `
resource "auth0_prompt_screen_renderer" "test_rendering" {
  prompt_type    = "login-id"
  screen_name    = "login-id"
  rendering_mode = "advanced"
  context_configuration = [
    "branding.settings",
    "screen.texts"
  ]
}

data "auth0_prompt_screen_renderers" "filtered_multiple" {
  depends_on = [auth0_prompt_screen_renderer.test_rendering]

  prompt         = "login-id"
  screen         = "login-id"
  rendering_mode = "advanced"
}
`

func TestAccDataPromptRenderingsWithMultipleFilters(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsDataSourceWithMultipleFilters,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.filtered_multiple", "renderings.#"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.filtered_multiple", "prompt", "login-id"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.filtered_multiple", "screen", "login-id"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.filtered_multiple", "rendering_mode", "advanced"),
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.filtered_multiple", "id"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.filtered_multiple", "renderings.0.prompt", "login-id"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.filtered_multiple", "renderings.0.screen", "login-id"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.filtered_multiple", "renderings.0.rendering_mode", "advanced"),
				),
			},
		},
	})
}

const testAccPromptRenderingsDataSourceEmpty = `
data "auth0_prompt_screen_renderers" "empty" {
  prompt = "device-flow"
  screen = "device-code-activation"
}
`

func TestAccDataPromptRenderingsEmpty(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsDataSourceEmpty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.empty", "id"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.empty", "prompt", "device-flow"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.empty", "screen", "device-code-activation"),
				),
			},
		},
	})
}

const testAccPromptRenderingsDataSourceByScreen = `
resource "auth0_prompt_screen_renderer" "signup" {
  prompt_type    = "signup-id"
  screen_name    = "signup-id"
  rendering_mode = "standard"
}

data "auth0_prompt_screen_renderers" "by_screen" {
  depends_on = [auth0_prompt_screen_renderer.signup]

  screen = "signup-id"
}
`

func TestAccDataPromptRenderingsByScreen(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptRenderingsDataSourceByScreen,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.by_screen", "renderings.#"),
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.by_screen", "screen", "signup-id"),
					resource.TestCheckResourceAttrSet("data.auth0_prompt_screen_renderers.by_screen", "id"),
					// Verify that all returned renderings have the correct screen.
					resource.TestCheckResourceAttr("data.auth0_prompt_screen_renderers.by_screen", "renderings.0.screen", "signup-id"),
				),
			},
		},
	})
}
