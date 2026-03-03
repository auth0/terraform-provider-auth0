package connection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccSCIMTokenGivenAConnection = `
resource "auth0_connection" "azure_ad" {
	name     = "Acceptance-Test-Azure-AD-{{.testName}}"
	strategy = "waad"
	show_as_button = true
	options {
		identity_api 	 = "microsoft-identity-platform-v2.0"
		client_id        = "123456"
		client_secret    = "123456"
		strategy_version = 2
		tenant_domain    = "example.onmicrosoft.com"
		domain           = "example.onmicrosoft.com"
		domain_aliases = [
			"example.com",
			"api.example.com"
		]
		use_wsfed            = false
		waad_protocol        = "openid-connect"
		waad_common_endpoint = false
		user_id_attribute    = "oid"
		api_enable_users     = true
		scopes               = [
			"basic_profile",
			"ext_groups",
			"ext_profile"
		]
		set_user_root_attributes = "on_each_login"
		should_trust_email_verified_connection = "never_set_emails_as_verified"
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}


resource "auth0_connection_scim_configuration" "my_scim_config" {
	connection_id = auth0_connection.azure_ad.id
}

resource "auth0_connection_scim_token" "my_scim_token" {
	connection_id = auth0_connection.azure_ad.id
	scopes = [
		"post:users",
		"get:users"
	]
	depends_on = [auth0_connection_scim_configuration.my_scim_config]
}
`

func TestAccSCIMToken(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccSCIMTokenGivenAConnection, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_connection.azure_ad", "id", "auth0_connection_scim_token.my_scim_token", "connection_id"),
					resource.TestCheckResourceAttr("auth0_connection_scim_token.my_scim_token", "scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_scim_token.my_scim_token", "scopes.*", "post:users"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_scim_token.my_scim_token", "scopes.*", "get:users"),
				),
			},
		},
	})
}
