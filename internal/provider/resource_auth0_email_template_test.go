package provider

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func init() {
	resource.AddTestSweepers("auth0_email_template", &resource.Sweeper{
		Name: "auth0_email_template",
		F: func(_ string) (err error) {
			api, err := Auth0()
			if err != nil {
				return
			}
			err = api.EmailTemplate.Update("welcome_email", &management.EmailTemplate{
				Enabled: auth0.Bool(false),
			})
			return
		},
	})
}

func TestAccEmailTemplate(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccEmailTemplateConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_email_template.my_email_template", "template", "welcome_email"),
					resource.TestCheckResourceAttr("auth0_email_template.my_email_template", "body", "<html><body><h1>Welcome!</h1></body></html>"),
					resource.TestCheckResourceAttr("auth0_email_template.my_email_template", "from", "welcome@example.com"),
					resource.TestCheckResourceAttr("auth0_email_template.my_email_template", "result_url", "https://example.com/welcome"),
					resource.TestCheckResourceAttr("auth0_email_template.my_email_template", "subject", "Welcome"),
					resource.TestCheckResourceAttr("auth0_email_template.my_email_template", "syntax", "liquid"),
					resource.TestCheckResourceAttr("auth0_email_template.my_email_template", "url_lifetime_in_seconds", "3600"),
					resource.TestCheckResourceAttr("auth0_email_template.my_email_template", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_email_template.my_email_template", "include_email_in_redirect", "false"),
				),
			},
		},
	})
}

const testAccEmailTemplateConfig = `
resource "auth0_email" "my_email_provider" {
	name = "ses"
	enabled = true
	default_from_address = "accounts@example.com"
	credentials {
		access_key_id = "AKIAXXXXXXXXXXXXXXXX"
		secret_access_key = "7e8c2148xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
		region = "us-east-1"
	}
}

resource "auth0_email_template" "my_email_template" {
	template = "welcome_email"
	body = "<html><body><h1>Welcome!</h1></body></html>"
	from = "welcome@example.com"
	result_url = "https://example.com/welcome"
	subject = "Welcome"
	syntax = "liquid"
	url_lifetime_in_seconds = 3600
	enabled = true
	include_email_in_redirect = false

	depends_on = ["auth0_email.my_email_provider"]
}
`
