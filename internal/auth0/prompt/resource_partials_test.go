package prompt_test

import (
	"context"
	"github.com/auth0/go-auth0/management"
	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

var (
	domain       = os.Getenv("AUTH0_DOMAIN")
	clientID     = os.Getenv("AUTH0_CLIENT_ID")
	clientSecret = os.Getenv("AUTH0_CLIENT_SECRET")
	manager, _   = management.New(domain, management.WithClientCredentials(context.Background(), clientID, clientSecret))
)

func TestAccPromptPartials(t *testing.T) {
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

const testAccGivenACustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "auth.terraform-provider-auth0.com"
	type   = "auth0_managed_certs"
}
`

const testGivenABrandingTemplate = `
resource "auth0_branding" "my_brand" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]

	universal_login {
		body = "<!DOCTYPE html><html><head>{%- auth0:head -%}</head><body>{%- auth0:widget -%}</body></html>"
	}
}
`

const testAccGivenPrerequisites = testAccGivenACustomDomain + testGivenABrandingTemplate

const testAccPromptPartialsCreate = testAccGivenPrerequisites + `
resource "auth0_prompt_partials" "prompt_partials" {
  depends_on = [ auth0_branding.my_brand ]

  prompt = "login"
  form_content_start = "<div>Test Header</div>"
}
`

const testAccPromptPartialsUpdate = testAccGivenPrerequisites + `
resource "auth0_prompt_partials" "prompt_partials" {
  depends_on = [ auth0_branding.my_brand ]

  prompt = "login"
  form_content_start = "<div>Updated Test Header</div>"
}
`
