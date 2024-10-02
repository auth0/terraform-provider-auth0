package prompt_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccPromptPartialsCreate = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_partials" "prompt_partials" {
  depends_on = [ auth0_branding.my_brand ]

  prompt = "login"

  form_content_start = "<div>Form Content Start</div>"
}
`

const testAccPromptPartialsUpdate = testAccGivenACustomDomain + testGivenABrandingTemplate + `
resource "auth0_prompt_partials" "prompt_partials" {
  depends_on = [ auth0_branding.my_brand ]

  prompt = "login"

  form_content_start      = "<div>Updated Form Content Start</div>"
  form_content_end        = "<div>Updated Form Content End</div>"
  form_footer_start       = "<div>Updated Footer Start</div>"
  form_footer_end         = "<div>Updated Footer End</div>"
  secondary_actions_start = "<div>Updated Secondary Actions Start</div>"
  secondary_actions_end   = "<div>Updated Secondary Actions End</div>"
}
`

func TestAccPromptPartials(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptPartialsCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_content_start", "<div>Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_content_end", ""),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_footer_start", ""),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_footer_end", ""),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "secondary_actions_start", ""),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "secondary_actions_end", ""),
				),
			},
			{
				Config: testAccPromptPartialsUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_content_start", "<div>Updated Form Content Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_content_end", "<div>Updated Form Content End</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_footer_start", "<div>Updated Footer Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "form_footer_end", "<div>Updated Footer End</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "secondary_actions_start", "<div>Updated Secondary Actions Start</div>"),
					resource.TestCheckResourceAttr("auth0_prompt_partials.prompt_partials", "secondary_actions_end", "<div>Updated Secondary Actions End</div>"),
				),
			},
		},
	})
}
