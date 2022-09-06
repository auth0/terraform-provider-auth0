package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
)

const testAccPromptCreate = `
resource "auth0_prompt" "prompt" {
  universal_login_experience = "classic"
  identifier_first = false
  webauthn_platform_first_factor = false
}
`

const testAccPromptUpdate = `
resource "auth0_prompt" "prompt" {
  universal_login_experience = "new"
  identifier_first = true
  webauthn_platform_first_factor = false
}
`

const testAccPromptUpdateAgain = `
resource "auth0_prompt" "prompt" {
  universal_login_experience = "new"
  identifier_first = false
  webauthn_platform_first_factor = true
}
`

func TestAccPrompt(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "universal_login_experience", "classic"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "identifier_first", "false"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "webauthn_platform_first_factor", "false"),
				),
			},
			{
				Config: testAccPromptUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "universal_login_experience", "new"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "identifier_first", "true"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "webauthn_platform_first_factor", "false"),
				),
			},
			{
				Config: testAccPromptUpdateAgain,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "universal_login_experience", "new"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "identifier_first", "false"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "webauthn_platform_first_factor", "true"),
				),
			},
		},
	})
}
