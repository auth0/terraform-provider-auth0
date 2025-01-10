package tokenexchangeprofile_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testTokenExchangeProfileCreate = `
resource "auth0_action" "my_action" {
	name = "Test Action {{.testName}}2"
	code = "exports.onExecutePostLogin = async (event, api) => {};"
	deploy = true
	supported_triggers {
		id      = "custom-token-exchange"
		version = "v1"
	}
}

resource "auth0_token_exchange_profile" "my_token_exchange_profile" {
	name = "my-token-exc-profile-{{.testName}}"
	subject_token_type = "https://acme.com/cis-token"
	action_id = auth0_action.my_action.id
	type = "custom_authentication"
}
`

const testTokenExchangeProfileUpdate = `
resource "auth0_action" "my_action" {
	name = "Test Action {{.testName}}2"
	code = "exports.onExecutePostLogin = async (event, api) => {};"
	deploy = true
	supported_triggers {
		id      = "custom-token-exchange"
		version = "v1"
	}
}

resource "auth0_token_exchange_profile" "my_token_exchange_profile" {
	name = "my-token-exc-profile-updated-{{.testName}}"
	subject_token_type = "https://acme.com/cis-token-updated"
	action_id = auth0_action.my_action.id
	type = "custom_authentication_updated"
}
`

func TestTokenExchangeProfile(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testTokenExchangeProfileCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "name", fmt.Sprintf("my-token-exc-profile-%s2", t.Name())),
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "subject_token_type", "https://acme.com/cis-token"),
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "type", "custom_authentication"),
					resource.TestCheckResourceAttrSet("auth0_token_exchange_profile.my_token_exchange_profile", "action_id"),
				),
			},
			{
				Config: acctest.ParseTestName(testTokenExchangeProfileUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "name", fmt.Sprintf("my-token-exc-profile-updated-%s2", t.Name())),
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "subject_token_type", "https://acme.com/cis-token-updated"),
					resource.TestCheckResourceAttr("auth0_token_exchange_profile.my_token_exchange_profile", "type", "custom_authentication_updated"),
					resource.TestCheckResourceAttrSet("auth0_token_exchange_profile.my_token_exchange_profile", "action_id"),
				),
			},
		},
	})
}
