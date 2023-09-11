package prompt_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccPromptCustomText(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccPromptCustomTextEmptyBody,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "language", "en"),
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "body", "{}"),
				),
			},
			{
				Config: testAccPromptCustomTextCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "language", "en"),
					resource.TestCheckResourceAttr(
						"auth0_prompt_custom_text.prompt_custom_text",
						"body",
						"{\n    \"login\": {\n        \"alertListTitle\": \"Alerts\",\n        \"buttonText\": \"Continue\",\n        \"emailPlaceholder\": \"Email address\",\n        \"title\": \"Welcome to ${companyName}\"\n    }\n}",
					),
				),
			},
			{
				Config: testAccPromptCustomTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "language", "en"),
					resource.TestCheckResourceAttr(
						"auth0_prompt_custom_text.prompt_custom_text",
						"body",
						"{\n    \"login\": {\n        \"alertListTitle\": \"Alerts\",\n        \"buttonText\": \"Proceed\",\n        \"emailPlaceholder\": \"Email Address\",\n        \"title\": \"Welcome to ${companyName}\"\n    }\n}",
					),
				),
			},
			{
				Config: testAccPromptCustomTextEmptyBody,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "language", "en"),
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "body", "{}"),
				),
			},
		},
	})
}

const testAccPromptCustomTextEmptyBody = `
resource "auth0_prompt_custom_text" "prompt_custom_text" {
  prompt = "login"
  language = "en"
  body = "{}"
}
`

const testAccPromptCustomTextCreate = `
resource "auth0_prompt_custom_text" "prompt_custom_text" {
  prompt = "login"
  language = "en"
  body = jsonencode(
    {
      "login" : {
		"alertListTitle" : "Alerts",
		"buttonText" : "Continue",
		"emailPlaceholder" : "Email address",
		"title" : "Welcome to $${companyName}"
      }
    }
  )
}
`

const testAccPromptCustomTextUpdate = `
resource "auth0_prompt_custom_text" "prompt_custom_text" {
  prompt = "login"
  language = "en"
  body = jsonencode(
    {
      "login" : {
		"alertListTitle" : "Alerts",
		"buttonText" : "Proceed",
		"emailPlaceholder" : "Email Address",
		"title" : "Welcome to $${companyName}"
      }
    }
  )
}
`
