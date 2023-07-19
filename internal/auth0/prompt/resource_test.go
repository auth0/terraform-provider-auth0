package prompt_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPromptEmpty = `
resource "auth0_prompt" "prompt" {
	identifier_first = false # Required by API to include at least one property
}
`

const testAccPromptCreate = `
resource "auth0_prompt" "prompt" {
	universal_login_experience     = "classic"
	identifier_first               = false
	webauthn_platform_first_factor = false
}
`

const testAccPromptUpdate = `
resource "auth0_prompt" "prompt" {
	universal_login_experience     = "new"
	identifier_first               = true
	webauthn_platform_first_factor = false
}
`

const testAccPromptUpdateAgain = `
resource "auth0_prompt" "prompt" {
	universal_login_experience     = "new"
	identifier_first               = false
	webauthn_platform_first_factor = true
}
`

func TestAccPrompt(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptEmpty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_prompt.prompt", "universal_login_experience"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "identifier_first", "false"),
					resource.TestCheckResourceAttrSet("auth0_prompt.prompt", "webauthn_platform_first_factor"),
				),
			},
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
			{
				Config: testAccPromptEmpty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "universal_login_experience", "new"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "identifier_first", "false"),
					resource.TestCheckResourceAttr("auth0_prompt.prompt", "webauthn_platform_first_factor", "true"),
				),
			},
		},
	})
}
