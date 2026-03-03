package selfserviceprofile_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccSSOCustomText(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccSSOCustomTextEmptyBody,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_self_service_profile_custom_text.sso_custom_text", "sso_id"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "page", "get-started"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "language", "en"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "body", "{}"),
				),
			},
			{
				Config: testAccSSOCustomTextCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_self_service_profile_custom_text.sso_custom_text", "sso_id"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "page", "get-started"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "language", "en"),
					resource.TestCheckResourceAttr(
						"auth0_self_service_profile_custom_text.sso_custom_text",
						"body",
						"{\n    \"introduction\": \"Welcome! With only a few steps you'll be able to setup your new connection.\"\n}",
					),
				),
			},
			{
				Config: testAccSSOCustomTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_self_service_profile_custom_text.sso_custom_text", "sso_id"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "language", "en"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "page", "get-started"),
					resource.TestCheckResourceAttr(
						"auth0_self_service_profile_custom_text.sso_custom_text",
						"body",
						"{\n    \"introduction\": \"Welcome! This is an updated Text\"\n}",
					),
				),
			},
			{
				Config: testAccSSOCustomTextEmptyBody,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_self_service_profile_custom_text.sso_custom_text", "sso_id"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "language", "en"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "page", "get-started"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "body", "{}"),
				),
			},
		},
	})
}

func TestAccSSOCustomTextImport(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccSSOCustomTextCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_self_service_profile_custom_text.sso_custom_text", "sso_id"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "page", "get-started"),
					resource.TestCheckResourceAttr("auth0_self_service_profile_custom_text.sso_custom_text", "language", "en"),
				),
			},
			{
				Config:            testAccSSOCustomTextCreate,
				ResourceName:      "auth0_self_service_profile_custom_text.sso_custom_text",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					ssoID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_self_service_profile.my_self_service_profile", "id")
					require.NoError(t, err)

					lang, err := acctest.ExtractResourceAttributeFromState(state, "auth0_self_service_profile_custom_text.sso_custom_text", "language")
					require.NoError(t, err)

					page, err := acctest.ExtractResourceAttributeFromState(state, "auth0_self_service_profile_custom_text.sso_custom_text", "page")
					require.NoError(t, err)

					return ssoID + "::" + lang + "::" + page, nil
				},
			},
		},
	})
}

const givenSelfServiceProfile = `
resource "auth0_self_service_profile" "my_self_service_profile" {
	name = "my-sso-profile"
	description = "sample description"
	allowed_strategies = ["oidc", "samlp"]
	user_attributes	{
		name		= "sample-name-{{.testName}}"
		description = "sample-description"
		is_optional = true
	}
	branding {
		logo_url    = "https://mycompany.org/v2/logo.png"
		colors {
			primary = "#0059d6"
		}
	}
}
`

const testAccSSOCustomTextEmptyBody = givenSelfServiceProfile + `
resource "auth0_self_service_profile_custom_text" "sso_custom_text" {
  sso_id = auth0_self_service_profile.my_self_service_profile.id
  language = "en"
  page = "get-started"
  body = "{}"
}
`

const testAccSSOCustomTextCreate = givenSelfServiceProfile + `
resource "auth0_self_service_profile_custom_text" "sso_custom_text" {
  sso_id = auth0_self_service_profile.my_self_service_profile.id
  language = "en"
  page = "get-started"
  body = jsonencode(
    {
		"introduction": "Welcome! With only a few steps you'll be able to setup your new connection."
    }
  )
}
`

const testAccSSOCustomTextUpdate = givenSelfServiceProfile + `
resource "auth0_self_service_profile_custom_text" "sso_custom_text" {
  sso_id = auth0_self_service_profile.my_self_service_profile.id
  language = "en"
  page = "get-started"
  body = jsonencode(
    {
		"introduction": "Welcome! This is an updated Text"
    }
  )
}
`
