package tokenexchangeprofile_test

import (
	"fmt"
	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"regexp"
	"testing"
)

const testAGivenTokenExchangeProfile = `
resource "auth0_action" "my_action" {
    name = "test-action"
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

resource "auth0_token_exchange_profile" "my_token_exchange_profile" {
    name = "token-prof-{{.testName}}"
    subject_token_type = "https://acme.com/cis-token"
    action_id = auth0_action.my_action.id
    type = "custom_authentication"
}

`

const testDataResourceWithoutID = testAGivenTokenExchangeProfile + `
data "auth0_token_exchange_profile" "my_profile" {
	depends_on = [ auth0_token_exchange_profile.my_token_exchange_profile ]
}`

const testDataResourceWithValidID = testAGivenTokenExchangeProfile + `
data "auth0_token_exchange_profile" "my_profile" {
	depends_on = [ auth0_token_exchange_profile.my_token_exchange_profile ]
    id = auth0_token_exchange_profile.my_token_exchange_profile.id
}`

const testDataResourceWithInvalidID = testAGivenTokenExchangeProfile + `
data "auth0_token_exchange_profile" "my_profile" {
	depends_on = [ auth0_token_exchange_profile.my_token_exchange_profile ]
    id = "tep_Tnvl88SKv98TkMmr"
}
`

func TestTokenExchangeDataSourceResourceRequiredId(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testDataResourceWithoutID,
				ExpectError: regexp.MustCompile("The argument \"id\" is required, but no definition was found."),
			},
		},
	})
}

func TestTokenExchangeDataSourceResource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testDataResourceWithInvalidID, t.Name()),
				ExpectError: regexp.MustCompile(
					`Error: 404 Not Found`,
				),
			},
			{
				Config: acctest.ParseTestName(testDataResourceWithValidID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_token_exchange_profile.my_profile", "name", fmt.Sprintf("token-prof-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_token_exchange_profile.my_profile", "subject_token_type", "https://acme.com/cis-token"),
					resource.TestCheckResourceAttr("data.auth0_token_exchange_profile.my_profile", "type", "custom_authentication"),
					resource.TestCheckResourceAttrSet("data.auth0_token_exchange_profile.my_profile", "action_id"),
				),
			},
		},
	})
}
