package connection_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccOktaConnection(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccConnectionOktaConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "name", fmt.Sprintf("Acceptance-Test-Okta-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "strategy", "okta"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "show_as_button", "true"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "client_id", "123456"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "client_secret", "123456"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "domain_aliases.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "domain_aliases.*", "example.com"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "domain_aliases.*", "api.example.com"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "issuer", "https://domain.okta.com"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "jwks_uri", "https://domain.okta.com/oauth2/v1/keys"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "token_endpoint", "https://domain.okta.com/oauth2/v1/token"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "userinfo_endpoint", "https://domain.okta.com/oauth2/v1/userinfo"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "authorization_endpoint", "https://domain.okta.com/oauth2/v1/authorize"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "scopes.#", "3"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "set_user_root_attributes", "on_each_login"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "non_persistent_attrs.*", "gender"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "non_persistent_attrs.*", "hair_color"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "upstream_params", `{"screen_name":{"alias":"login_hint"}}`),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "icon_url", "https://example.com/logo.svg"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccConnectionOktaConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "name", fmt.Sprintf("Acceptance-Test-Okta-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "strategy", "okta"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "show_as_button", "false"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "client_id", "123456"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "client_secret", "123456"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "domain_aliases.#", "1"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "domain_aliases.*", "example.com"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "issuer", "https://domain.okta.com"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "jwks_uri", "https://domain.okta.com/oauth2/v2/keys"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "token_endpoint", "https://domain.okta.com/oauth2/v2/token"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "userinfo_endpoint", "https://domain.okta.com/oauth2/v2/userinfo"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "authorization_endpoint", "https://domain.okta.com/oauth2/v2/authorize"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "scopes.*", "profile"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "set_user_root_attributes", "on_first_login"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_okta.okta", "non_persistent_attrs.*", "gender"),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "upstream_params", ""),
					resource.TestCheckResourceAttr("auth0_connection_okta.okta", "icon_url", "https://example.com/v2/logo.svg"),
				),
			},
		},
	})
}

const testAccConnectionOktaConfig = `
resource "auth0_connection_okta" "okta" {
	name           = "Acceptance-Test-Okta-{{.testName}}"
	display_name   = "Acceptance-Test-Okta-{{.testName}}"

	show_as_button = true

	client_id                = "123456"
	client_secret            = "123456"
	domain                   = "domain.okta.com"
	domain_aliases           = [ "example.com", "api.example.com" ]
	issuer                   = "https://domain.okta.com"
	jwks_uri                 = "https://domain.okta.com/oauth2/v1/keys"
	token_endpoint           = "https://domain.okta.com/oauth2/v1/token"
	userinfo_endpoint        = "https://domain.okta.com/oauth2/v1/userinfo"
	authorization_endpoint   = "https://domain.okta.com/oauth2/v1/authorize"
	scopes                   = [ "openid", "profile", "email" ]
	non_persistent_attrs     = [ "gender", "hair_color" ]
	set_user_root_attributes = "on_each_login"
	icon_url                 = "https://example.com/logo.svg"
	upstream_params = jsonencode({
		"screen_name": {
			"alias": "login_hint"
		}
	})
}
`

const testAccConnectionOktaConfigUpdate = `
resource "auth0_connection_okta" "okta" {
	name           = "Acceptance-Test-Okta-{{.testName}}"
	display_name   = "Acceptance-Test-Okta-{{.testName}}"

	show_as_button = false

	client_id                = "123456"
	client_secret            = "123456"
	domain                   = "domain.okta.com"
	domain_aliases           = [ "example.com" ]
	issuer                   = "https://domain.okta.com"
	jwks_uri                 = "https://domain.okta.com/oauth2/v2/keys"
	token_endpoint           = "https://domain.okta.com/oauth2/v2/token"
	userinfo_endpoint        = "https://domain.okta.com/oauth2/v2/userinfo"
	authorization_endpoint   = "https://domain.okta.com/oauth2/v2/authorize"
	scopes                   = [ "openid", "profile"]
	non_persistent_attrs     = [ "gender" ]
	set_user_root_attributes = "on_first_login"
	icon_url                 = "https://example.com/v2/logo.svg"
}
`
