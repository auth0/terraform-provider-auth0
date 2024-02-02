package prompt_test

import (
	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"testing"
)

func TestPromptPartials(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptPartialsCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0prompt_partials.prompt_partials", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0prompt_partials.prompt_partials", "form_content_start", "<div>Test Header</div>"),
				),
			},
			{
				Config: testAccPromptPartialsUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0prompt_partials.prompt_partials", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0prompt_partials.prompt_partials", "form_content_start", "<div>Updated Test Header</div>"),
				),
			},
		},
	})
}

const testAccPromptPartialsCreate = `
resource "auth0_prompt_partials" "prompt_partials" {
  prompt = "login"
  form_content_start = "<div>Test Header</div>"
}
`

const testAccPromptPartialsUpdate = `
resource "auth0_prompt_partials" "prompt_partials" {
  prompt = "login"
  form_content_start = "<div>Updated Test Header</div>"
}
`
