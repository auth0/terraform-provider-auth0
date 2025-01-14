package tokenexchangeprofile_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const givenACustomTokenAction = `
resource "auth0_action" "my_action" {
	name = "{{.testName}}-Action"
	code = <<-EOT
		exports.onContinuePostLogin = async (event, api) => {
			console.log("foo")
		};"
		EOT
	deploy = true
	supported_triggers {
		id      = "custom-token-exchange"
		version = "v1"
	}
}
`

const testTokenExchangeProfileCreate = givenACustomTokenAction + `
resource "auth0_token_exchange_profile" "my_token_exchange_profile" {
	name = "token-prof-{{.testName}}"
	subject_token_type = "https://acme.com/cis-token"
	action_id = auth0_action.my_action.id
	type = "custom_authentication"
}
`

const testTokenExchangeProfileUpdate = givenACustomTokenAction + `
resource "auth0_token_exchange_profile" "my_token_exchange_profile" {
	name = "token-prof-updated-{{.testName}}"
	subject_token_type = "https://acme.com/cis-token-updated"
	action_id = auth0_action.my_action.id
	type = "custom_authentication"
}
`

func TestTokenExchangeProfile(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testTokenExchangeProfileCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "name", fmt.Sprintf("token-prof-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "subject_token_type", "https://acme.com/cis-token"),
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "type", "custom_authentication"),
					resource.TestCheckResourceAttrSet("auth0_token_exchange_profile.my_token_exchange_profile", "action_id"),
				),
			},
			{
				Config: acctest.ParseTestName(testTokenExchangeProfileUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "name", fmt.Sprintf("token-prof-updated-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "subject_token_type", "https://acme.com/cis-token-updated"),
				),
			},
			{
				Config: acctest.ParseTestName(givenACustomTokenAction, t.Name()),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}
