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

	testClientCreate = `
resource "auth0_client" "my_client-1" {
	name = "Acceptance-Test-Client-1-{{.testName}}"
}`
	testClientCreate2 = `
resource "auth0_client" "my_client-2" {
	name = "Acceptance-Test-Client-2-{{.testName}}"
}`

	testAccPromptScreenRenderWithoutSettings = `
resource "auth0_prompt_screen_renderer" "login-id" {
  prompt_type     = "login-id"
  screen_name =  "login-id"
}
`

	testAccPromptScreenRenderCreate = testAccPromptScreenRenderWithoutSettings + testClientCreate + testClientCreate2 + `
resource "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
    prompt_type                = "mfa-push"
    screen_name                = "mfa-push-enrollment-qr"
    rendering_mode             = "advanced"
    default_head_tags_disabled = false
    filters {
        match_type = "includes_any"
        clients = jsonencode([
            {
                id = "LeBGFyt7y2ZjvlBhqPBJwTn3dLoEhCGB"
            }
        ])
        organizations = jsonencode([
            {
                metadata = {
                    some_key = "some_value"
                },
            }
        ])
#         domains = jsonencode([
#             {
#                 id = "y8zHiIOI5UciKi6yh6yPAQ3FpxvghHFb"
#             }
#         ])
    }
    context_configuration = [
        "branding.settings",
        "branding.themes.default",
        "client.logo_uri",
        "client.description",
        "client.metadata.key",
        "organization.display_name",
        "organization.branding",
        "organization.metadata.key",
        "screen.texts",
        "tenant.name",
        "tenant.friendly_name",
        "tenant.enabled_locales",
        "untrusted_data.submitted_form_data",
        "untrusted_data.authorization_params.login_hint",
        "untrusted_data.authorization_params.screen_hint",
        "untrusted_data.authorization_params.ui_locales",
        "untrusted_data.authorization_params.ext-key",
    ]
    head_tags = jsonencode([
        {
            attributes : {
                "async" : true,
                "defer" : true,
                "integrity" : [
                    "sha512-v2CJ7UaYy4JwqLDIrZUI/4hqeoQieOmAZNXBeQyjo21dadnwR+8ZaIJVT8EE2iyI61OV8e6M8PP2/4hpQINQ/g=="
                ],
                "src" : "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.2/jquery.min.js"
            },
            tag : "script"
        }
    ])
}

`

	testAccPromptScreenRenderUpdate = testAccPromptScreenRenderWithoutSettings + testClientCreate + testClientCreate2 + `
resource "auth0_prompt_screen_renderer" "prompt_screen_renderer" {
    prompt_type                = "mfa-push"
    screen_name                = "mfa-push-enrollment-qr"
    rendering_mode             = "advanced"
    default_head_tags_disabled = false
    filters {
        match_type = "includes_any"
        clients = jsonencode([
            {
                id = "LeBGFyt7y2ZjvlBhqPBJwTn3dLoEhCGB"
            }
        ])
        organizations = jsonencode([
            {
                metadata = {
                    some_key = "some_value"
                },
            }
        ])
#         domains = jsonencode([
#             {
#                 id = "y8zHiIOI5UciKi6yh6yPAQ3FpxvghHFb"
#             }
#         ])
    }
    context_configuration = [
        "branding.settings",
        "branding.themes.default",
        "client.logo_uri",
        "client.description",
        "client.metadata.key",
        "organization.display_name",
        "organization.branding",
        "organization.metadata.key",
        "screen.texts",
        "tenant.name",
        "tenant.friendly_name",
        "tenant.enabled_locales",
        "untrusted_data.submitted_form_data",
        "untrusted_data.authorization_params.login_hint",
        "untrusted_data.authorization_params.screen_hint",
        "untrusted_data.authorization_params.ui_locales",
        "untrusted_data.authorization_params.ext-key",
    ]
    head_tags = jsonencode([
        {
            attributes : {
                "async" : true,
                "defer" : true,
                "integrity" : [
                    "sha512-v2CJ7UaYy4JwqLDIrZUI/4hqeoQieOmAZNXBeQyjo21dadnwR+8ZaIJVT8EE2iyI61OV8e6M8PP2/4hpQINQ/g=="
                ],
                "src" : "https://cdnjs.cloudflare.com/ajax/libs/jquery/3.7.2/jquery.min.js"
            },
            tag : "script"
        }
    ])
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
