package auth0

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccPromptCustomText(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccPromptCustomTextCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "prompt", "login"),
					resource.TestCheckResourceAttr("auth0_prompt_custom_text.prompt_custom_text", "language", "en"),
					resource.TestCheckResourceAttr(
						"auth0_prompt_custom_text.prompt_custom_text",
						"body",
						"{\n    \"login\": {\n        \"alertListTitle\": \"Alerts\",\n        \"buttonText\": \"Continue\",\n        \"emailPlaceholder\": \"Email address\"\n    }\n}",
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
						"{\n    \"login\": {\n        \"alertListTitle\": \"Alerts\",\n        \"buttonText\": \"Proceed\",\n        \"emailPlaceholder\": \"Email Address\"\n    }\n}",
					),
				),
			},
		},
	})
}

const testAccPromptCustomTextCreate = `
resource "auth0_prompt_custom_text" "prompt_custom_text" {
  prompt = "login"
  language = "en"
  body = jsonencode(
    {
      "login" : {
        "alertListTitle" : "Alerts",
        "buttonText" : "Continue",
        "emailPlaceholder" : "Email address"
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
        "emailPlaceholder" : "Email Address"
      }
    }
  )
}
`
