package auth0

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/template"
)

func init() {
	resource.AddTestSweepers("auth0_connection", &resource.Sweeper{
		Name: "auth0_connection",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				connectionList, err := api.Connection.List(
					management.IncludeFields("id", "name"),
					management.Page(page),
				)
				if err != nil {
					return err
				}

				for _, connection := range connectionList.Connections {
					log.Printf("[DEBUG] ➝ %s", connection.GetName())

					if strings.Contains(connection.GetName(), "Test") {
						result = multierror.Append(
							result,
							api.Connection.Delete(connection.GetID()),
						)
						log.Printf("[DEBUG] ✗ %s", connection.GetName())
					}
				}
				if !connectionList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}

func TestAccConnection(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "is_domain_connection", "true"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "metadata.key1", "foo"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "metadata.key2", "bar"),
					resource.TestCheckNoResourceAttr("auth0_connection.my_connection", "show_as_button"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.password_policy", "fair"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.password_no_personal_info.0.enable", "true"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.password_dictionary.0.enable", "true"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.password_complexity_options.0.min_length", "6"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.enabled_database_customization", "false"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.brute_force_protection", "true"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.import_mode", "false"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.disable_signup", "false"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.requires_username", "true"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.validation.0.username.0.min", "10"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.validation.0.username.0.max", "40"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.custom_scripts.get_user", "myFunction"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.mfa.0.active", "true"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.mfa.0.return_enroll_settings", "true"),
					resource.TestCheckResourceAttrSet("auth0_connection.my_connection", "options.0.configuration.foo"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
			{
				Config: template.ParseTestName(testAccConnectionConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.brute_force_protection", "false"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.mfa.0.return_enroll_settings", "false"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.upstream_params", ""),
				),
			},
		},
	})
}

const testAccConnectionConfig = `
resource "auth0_connection" "my_connection" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	is_domain_connection = true
	strategy = "auth0"
	metadata = {
		key1 = "foo"
		key2 = "bar"
	}
	options {
		password_policy = "fair"
		password_history {
			enable = true
			size = 5
		}
		password_no_personal_info {
			enable = true
		}
		password_dictionary {
			enable = true
			dictionary = [ "password", "admin", "1234" ]
		}
		password_complexity_options {
			min_length = 6
		}
		validation {
			username {
				min = 10
				max = 40
			}
		}
		enabled_database_customization = false
		brute_force_protection = true
		import_mode = false
		requires_username = true
		disable_signup = false
		custom_scripts = {
			get_user = "myFunction"
		}
		configuration = {
			foo = "bar"
		}
		mfa {
			active                 = true
			return_enroll_settings = true
		}
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

const testAccConnectionConfigUpdate = `
resource "auth0_connection" "my_connection" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	is_domain_connection = true
	strategy = "auth0"
	metadata = {
		key1 = "foo"
		key2 = "bar"
	}
	options {
		password_policy = "fair"
		password_history {
			enable = true
			size = 5
		}
		password_no_personal_info {
			enable = true
		}
		enabled_database_customization = false
		brute_force_protection = false
		import_mode = false
		disable_signup = false
		requires_username = true
		custom_scripts = {
			get_user = "myFunction"
		}
		configuration = {
			foo = "bar"
		}
		mfa {
			active                 = true
			return_enroll_settings = false
		}
	}
}
`

func TestAccConnectionAD(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionADConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.ad", "name", fmt.Sprintf("Acceptance-Test-AD-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.ad", "strategy", "ad"),
					resource.TestCheckResourceAttr("auth0_connection.ad", "show_as_button", "true"),
					resource.TestCheckResourceAttr("auth0_connection.ad", "options.0.domain_aliases.#", "2"),
					resource.TestCheckResourceAttr("auth0_connection.ad", "options.0.tenant_domain", "example.com"),
					resource.TestCheckResourceAttr("auth0_connection.ad", "options.0.use_kerberos", "false"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.ad", "options.0.ips.*", "192.168.1.2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.ad", "options.0.ips.*", "192.168.1.1"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.ad", "options.0.domain_aliases.*", "example.com"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.ad", "options.0.domain_aliases.*", "api.example.com"),
					resource.TestCheckResourceAttr("auth0_connection.ad", "options.0.set_user_root_attributes", "on_each_login"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.ad", "options.0.non_persistent_attrs.*", "ethnicity"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.ad", "options.0.non_persistent_attrs.*", "gender"),
					resource.TestCheckResourceAttr("auth0_connection.ad", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
		},
	})
}

const testAccConnectionADConfig = `
resource "auth0_connection" "ad" {
	name = "Acceptance-Test-AD-{{.testName}}"
	strategy = "ad"
	show_as_button = true
	options {
		brute_force_protection = true
		tenant_domain = "example.com"
		domain_aliases = [
			"example.com",
			"api.example.com"
		]
		ips = [ "192.168.1.1", "192.168.1.2" ]
		set_user_root_attributes = "on_each_login"
		non_persistent_attrs = ["ethnicity","gender"]
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

func TestAccConnectionAzureAD(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionAzureADConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "name", fmt.Sprintf("Acceptance-Test-Azure-AD-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "strategy", "waad"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "show_as_button", "true"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "options.0.client_id", "123456"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "options.0.client_secret", "123456"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "options.0.tenant_domain", "example.onmicrosoft.com"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "options.0.domain", "example.onmicrosoft.com"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "options.0.domain_aliases.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.azure_ad", "options.0.domain_aliases.*", "example.com"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.azure_ad", "options.0.domain_aliases.*", "api.example.com"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "options.0.scopes.#", "3"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.azure_ad", "options.0.scopes.*", "basic_profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.azure_ad", "options.0.scopes.*", "ext_profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.azure_ad", "options.0.scopes.*", "ext_groups"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "options.0.set_user_root_attributes", "on_each_login"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "options.0.should_trust_email_verified_connection", "never_set_emails_as_verified"),
					resource.TestCheckResourceAttr("auth0_connection.azure_ad", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
		},
	})
}

const testAccConnectionAzureADConfig = `
resource "auth0_connection" "azure_ad" {
	name     = "Acceptance-Test-Azure-AD-{{.testName}}"
	strategy = "waad"
	show_as_button = true
	options {
		client_id     = "123456"
		client_secret = "123456"
		tenant_domain = "example.onmicrosoft.com"
		domain        = "example.onmicrosoft.com"
		domain_aliases = [
			"example.com",
			"api.example.com"
		]
		use_wsfed            = false
		waad_protocol        = "openid-connect"
		waad_common_endpoint = false
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
`

func TestAccConnectionOIDC(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionOIDCConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.oidc", "name", fmt.Sprintf("Acceptance-Test-OIDC-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "strategy", "oidc"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "show_as_button", "true"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.client_id", "123456"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.client_secret", "123456"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.domain_aliases.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.domain_aliases.*", "example.com"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.domain_aliases.*", "api.example.com"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.type", "back_channel"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.issuer", "https://api.login.yahoo.com"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.jwks_uri", "https://api.login.yahoo.com/openid/v1/certs"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.discovery_url", "https://api.login.yahoo.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.token_endpoint", "https://api.login.yahoo.com/oauth2/get_token"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.userinfo_endpoint", "https://api.login.yahoo.com/openid/v1/userinfo"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.authorization_endpoint", "https://api.login.yahoo.com/oauth2/request_auth"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.scopes.#", "3"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.set_user_root_attributes", "on_each_login"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.non_persistent_attrs.*", "gender"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.non_persistent_attrs.*", "hair_color"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
			{
				Config: template.ParseTestName(testAccConnectionOIDCConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.oidc", "show_as_button", "false"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.client_id", "1234567"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.client_secret", "1234567"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.domain_aliases.#", "1"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.domain_aliases.*", "example.com"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.type", "front_channel"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.issuer", "https://www.paypalobjects.com"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.jwks_uri", "https://api.paypal.com/v1/oauth2/certs"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.discovery_url", "https://www.paypalobjects.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.token_endpoint", "https://api.paypal.com/v1/oauth2/token"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.userinfo_endpoint", "https://api.paypal.com/v1/oauth2/token/userinfo"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.authorization_endpoint", "https://www.paypal.com/signin/authorize"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oidc", "options.0.scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.set_user_root_attributes", "on_first_login"),
					resource.TestCheckResourceAttr("auth0_connection.oidc", "options.0.upstream_params", ""),
				),
			},
		},
	})
}

const testAccConnectionOIDCConfig = `
resource "auth0_connection" "oidc" {
	name     = "Acceptance-Test-OIDC-{{.testName}}"
	display_name     = "Acceptance-Test-OIDC-{{.testName}}"
	strategy = "oidc"
	show_as_button = true
	options {
		client_id     = "123456"
		client_secret = "123456"
		domain_aliases = [
			"example.com",
			"api.example.com"
		]
		type                   = "back_channel"
		issuer                 = "https://api.login.yahoo.com"
		jwks_uri               = "https://api.login.yahoo.com/openid/v1/certs"
		discovery_url          = "https://api.login.yahoo.com/.well-known/openid-configuration"
		token_endpoint         = "https://api.login.yahoo.com/oauth2/get_token"
		userinfo_endpoint      = "https://api.login.yahoo.com/openid/v1/userinfo"
		authorization_endpoint = "https://api.login.yahoo.com/oauth2/request_auth"
		scopes                 = [ "openid", "email", "profile" ]
		set_user_root_attributes = "on_each_login"
		non_persistent_attrs = ["gender","hair_color"]
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

const testAccConnectionOIDCConfigUpdate = `
resource "auth0_connection" "oidc" {
	name     = "Acceptance-Test-OIDC-{{.testName}}"
	display_name     = "Acceptance-Test-OIDC-{{.testName}}"
	strategy = "oidc"
	show_as_button = false
	options {
		client_id     = "1234567"
		client_secret = "1234567"
		domain_aliases = [
			"example.com"
		]
		type                   = "front_channel"
		issuer                 = "https://www.paypalobjects.com"
		jwks_uri               = "https://api.paypal.com/v1/oauth2/certs"
		discovery_url          = "https://www.paypalobjects.com/.well-known/openid-configuration"
		token_endpoint         = "https://api.paypal.com/v1/oauth2/token"
		userinfo_endpoint      = "https://api.paypal.com/v1/oauth2/token/userinfo"
		authorization_endpoint = "https://www.paypal.com/signin/authorize"
		scopes                 = [ "openid", "email" ]
		set_user_root_attributes = "on_first_login"
	}
}
`

func TestAccConnectionOAuth2(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionOAuth2Config, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "name", fmt.Sprintf("Acceptance-Test-OAuth2-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "strategy", "oauth2"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.client_id", "123456"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.client_secret", "123456"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.token_endpoint", "https://api.login.yahoo.com/oauth2/get_token"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.authorization_endpoint", "https://api.login.yahoo.com/oauth2/request_auth"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.scopes.#", "3"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oauth2", "options.0.scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oauth2", "options.0.scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oauth2", "options.0.scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.scripts.fetchUserProfile", "function( { return callback(null) }"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.set_user_root_attributes", "on_each_login"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.icon_url", ""),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.pkce_enabled", "true"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
			{
				Config: template.ParseTestName(testAccConnectionOAuth2ConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.client_id", "1234567"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.client_secret", "1234567"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.token_endpoint", "https://api.paypal.com/v1/oauth2/token"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.authorization_endpoint", "https://www.paypal.com/signin/authorize"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oauth2", "options.0.scopes.*", "openid"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.oauth2", "options.0.scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.scripts.fetchUserProfile", "function( { return callback(null) }"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.set_user_root_attributes", "on_first_login"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.icon_url", "https://cdn.paypal.com/assets/logo.png"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.pkce_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_connection.oauth2", "options.0.upstream_params", ""),
				),
			},
		},
	})
}

const testAccConnectionOAuth2Config = `
resource "auth0_connection" "oauth2" {
	name     = "Acceptance-Test-OAuth2-{{.testName}}"
	strategy = "oauth2"
	is_domain_connection = false
	options {
		client_id     = "123456"
		client_secret = "123456"
		token_endpoint         = "https://api.login.yahoo.com/oauth2/get_token"
		authorization_endpoint = "https://api.login.yahoo.com/oauth2/request_auth"
		scopes = [ "openid", "email", "profile" ]
		set_user_root_attributes = "on_each_login"
		scripts = {
			fetchUserProfile= "function( { return callback(null) }"
		}
		pkce_enabled = true
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

const testAccConnectionOAuth2ConfigUpdate = `
resource "auth0_connection" "oauth2" {
	name     = "Acceptance-Test-OAuth2-{{.testName}}"
	strategy = "oauth2"
	is_domain_connection = false
	options {
		client_id     = "1234567"
		client_secret = "1234567"
		token_endpoint         = "https://api.paypal.com/v1/oauth2/token"
		authorization_endpoint = "https://www.paypal.com/signin/authorize"
		scopes = [ "openid", "email" ]
		set_user_root_attributes = "on_first_login"
		icon_url = "https://cdn.paypal.com/assets/logo.png"
		scripts = {
			fetchUserProfile= "function( { return callback(null) }"
		}
		pkce_enabled = false
	}
}
`

func TestAccConnectionWithEnabledClients(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionWithEnabledClientsConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "enabled_clients.#", "4"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.#", "1"), // Gets computed with defaults by the API.
				),
			},
		},
	})
}

const testAccConnectionWithEnabledClientsConfig = `
resource "auth0_client" "my_client_1" {
	name = "Application - Acceptance Test - 1 - {{.testName}}"
	description = "Test Applications Long Description"
	app_type = "non_interactive"
}

resource "auth0_client" "my_client_2" {
	depends_on = [auth0_client.my_client_1]
	name = "Application - Acceptance Test - 2 - {{.testName}}"
	description = "Test Applications Long Description"
	app_type = "non_interactive"
}

resource "auth0_client" "my_client_3" {
	depends_on = [auth0_client.my_client_2]
	name = "Application - Acceptance Test - 3 - {{.testName}}"
	description = "Test Applications Long Description"
	app_type = "non_interactive"
}

resource "auth0_client" "my_client_4" {
	depends_on = [auth0_client.my_client_3]
	name = "Application - Acceptance Test - 4 - {{.testName}}"
	description = "Test Applications Long Description"
	app_type = "non_interactive"
}

resource "auth0_connection" "my_connection" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	is_domain_connection = true
	strategy = "auth0"
	enabled_clients = [
		auth0_client.my_client_1.id,
		auth0_client.my_client_2.id,
		auth0_client.my_client_3.id,
		auth0_client.my_client_4.id,
	]
}
`

func TestAccConnectionSMS(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionSMSConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.sms", "name", fmt.Sprintf("Acceptance-Test-SMS-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.sms", "strategy", "sms"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.twilio_sid", "ABC123"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.twilio_token", "DEF456"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.totp.#", "1"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.totp.0.time_step", "300"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.totp.0.length", "6"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
		},
	})
}

const testAccConnectionSMSConfig = `
resource "auth0_connection" "sms" {
	name = "Acceptance-Test-SMS-{{.testName}}"
	is_domain_connection = false
	strategy = "sms"
	options {
		disable_signup = false
		name = "SMS OTP"
		twilio_sid = "ABC123"
		twilio_token = "DEF456"
		from = "+12345678"
		syntax = "md_with_macros"
		template = "@@password@@"
		messaging_service_sid = "GHI789"
		brute_force_protection = true
		totp {
			time_step = 300
			length = 6
		}
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

func TestAccConnectionCustomSMS(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionCustomSMSConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.sms", "name", fmt.Sprintf("Acceptance-Test-Custom-SMS-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.sms", "strategy", "sms"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.totp.#", "1"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.totp.0.time_step", "300"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.totp.0.length", "6"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.gateway_url", "https://somewhere.com/sms-gateway"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.gateway_authentication.#", "1"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.gateway_authentication.0.method", "bearer"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.gateway_authentication.0.subject", "test.us.auth0.com:sms"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.gateway_authentication.0.audience", "https://somewhere.com/sms-gateway"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.gateway_authentication.0.secret", "4e2680bb72ec2ae24836476dd37ed6c2"),
					resource.TestCheckResourceAttr("auth0_connection.sms", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
		},
	})
}

const testAccConnectionCustomSMSConfig = `
resource "auth0_connection" "sms" {
	name = "Acceptance-Test-Custom-SMS-{{.testName}}"
	is_domain_connection = false
	strategy = "sms"
	options {
		disable_signup = false
		name = "sms"
		from = "+12345678"
		syntax = "md_with_macros"
		template = "@@password@@"
		brute_force_protection = true
		totp {
			time_step = 300
			length = 6
		}
		provider = "sms_gateway"
		gateway_url = "https://somewhere.com/sms-gateway"
		gateway_authentication {
			method = "bearer"
			subject = "test.us.auth0.com:sms"
			audience = "https://somewhere.com/sms-gateway"
			secret = "4e2680bb72ec2ae24836476dd37ed6c2"
			secret_base64_encoded = false
		}
		forward_request_info = true
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

func TestAccConnectionEmail(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionEmailConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.email", "name", fmt.Sprintf("Acceptance-Test-Email-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.email", "strategy", "email"),
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.from", "Magic Password <password@example.com>"),
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.subject", "Sign in!"),
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.totp.#", "1"),
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.totp.0.time_step", "300"),
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.totp.0.length", "6"),
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
			{
				Config: template.ParseTestName(testAccConnectionEmailConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.totp.#", "1"),
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.totp.0.time_step", "360"),
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.totp.0.length", "4"),
					resource.TestCheckResourceAttr("auth0_connection.email", "options.0.upstream_params", ""),
				),
			},
		},
	})
}

const testAccConnectionEmailConfig = `
resource "auth0_connection" "email" {
	name = "Acceptance-Test-Email-{{.testName}}"
	is_domain_connection = false
	strategy = "email"
	options {
		disable_signup = false
		name = "Email OTP"
		from = "Magic Password <password@example.com>"
		subject = "Sign in!"
		syntax = "liquid"
		template = "<html><body><h1>Here's your password!</h1></body></html>"
		brute_force_protection = true
		totp {
			time_step = 300
			length = 6
		}
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}

`

const testAccConnectionEmailConfigUpdate = `
resource "auth0_connection" "email" {
	name = "Acceptance-Test-Email-{{.testName}}"
	is_domain_connection = false
	strategy = "email"
	options {
		disable_signup = false
		name = "Email OTP"
		from = "Magic Password <password@example.com>"
		subject = "Sign in!"
		syntax = "liquid"
		template = "<html><body><h1>Here's your password!</h1></body></html>"
		brute_force_protection = true
		totp {
			time_step = 360
			length = 4
		}
	}
}
`

func TestAccConnectionSalesforce(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionSalesforceConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.salesforce_community", "name", fmt.Sprintf("Acceptance-Test-Salesforce-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.salesforce_community", "strategy", "salesforce-community"),
					resource.TestCheckResourceAttr("auth0_connection.salesforce_community", "options.0.community_base_url", "https://salesforce.example.com"),
					resource.TestCheckResourceAttr("auth0_connection.salesforce_community", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
		},
	})
}

const testAccConnectionSalesforceConfig = `
resource "auth0_connection" "salesforce_community" {
	name = "Acceptance-Test-Salesforce-Connection-{{.testName}}"
	is_domain_connection = false
	strategy = "salesforce-community"
	options {
		client_id = "client-id"
		client_secret = "client-secret"
		community_base_url = "https://salesforce.example.com"
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

func TestAccConnectionGoogleOAuth2(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionGoogleOAuth2Config, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.google_oauth2", "name", fmt.Sprintf("Acceptance-Test-Google-OAuth2-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.google_oauth2", "strategy", "google-oauth2"),
					resource.TestCheckResourceAttr("auth0_connection.google_oauth2", "options.0.client_id", ""),
					resource.TestCheckResourceAttr("auth0_connection.google_oauth2", "options.0.client_secret", ""),
					resource.TestCheckResourceAttr("auth0_connection.google_oauth2", "options.0.allowed_audiences.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.google_oauth2", "options.0.allowed_audiences.*", "example.com"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.google_oauth2", "options.0.allowed_audiences.*", "api.example.com"),
					resource.TestCheckResourceAttr("auth0_connection.google_oauth2", "options.0.scopes.#", "4"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.google_oauth2", "options.0.scopes.*", "email"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.google_oauth2", "options.0.scopes.*", "profile"),
					resource.TestCheckResourceAttr("auth0_connection.google_oauth2", "options.0.set_user_root_attributes", "on_each_login"),
					resource.TestCheckResourceAttr("auth0_connection.google_oauth2", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
		},
	})
}

const testAccConnectionGoogleOAuth2Config = `
resource "auth0_connection" "google_oauth2" {
	name = "Acceptance-Test-Google-OAuth2-{{.testName}}"
	is_domain_connection = false
	strategy = "google-oauth2"
	options {
		client_id = ""
		client_secret = ""
		allowed_audiences = [ "example.com", "api.example.com" ]
		scopes = [ "email", "profile", "gmail", "youtube" ]
		set_user_root_attributes = "on_each_login"
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

func TestAccConnectionGoogleApps(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionGoogleApps, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "name", fmt.Sprintf("Acceptance-Test-Google-Apps-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "strategy", "google-apps"),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "show_as_button", "false"),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "options.0.client_id", ""),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "options.0.client_secret", ""),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "options.0.domain", "example.com"),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "options.0.tenant_domain", "example.com"),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "options.0.domain_aliases.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.google_apps", "options.0.domain_aliases.*", "example.com"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.google_apps", "options.0.domain_aliases.*", "api.example.com"),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "options.0.api_enable_users", "true"),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "options.0.scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.google_apps", "options.0.scopes.*", "ext_profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.google_apps", "options.0.scopes.*", "ext_groups"),
					resource.TestCheckResourceAttr("auth0_connection.google_apps", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
		},
	})
}

const testAccConnectionGoogleApps = `
resource "auth0_connection" "google_apps" {
	name = "Acceptance-Test-Google-Apps-{{.testName}}"
	is_domain_connection = false
	strategy = "google-apps"
	show_as_button = false
	options {
		client_id = ""
		client_secret = ""
		domain = "example.com"
		tenant_domain = "example.com"
		domain_aliases = [ "example.com", "api.example.com" ]
		api_enable_users = true
		scopes = [ "ext_profile", "ext_groups" ]
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

func TestAccConnectionFacebook(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionFacebookConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.facebook", "name", fmt.Sprintf("Acceptance-Test-Facebook-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "strategy", "facebook"),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "options.0.client_id", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "options.0.client_secret", "client_secret"),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "options.0.scopes.#", "4"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.facebook", "options.0.scopes.*", "public_profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.facebook", "options.0.scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
			{
				Config: template.ParseTestName(testAccConnectionFacebookConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.facebook", "name", fmt.Sprintf("Acceptance-Test-Facebook-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "strategy", "facebook"),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "options.0.client_id", "client_id_update"),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "options.0.client_secret", "client_secret_update"),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "options.0.scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.facebook", "options.0.scopes.*", "public_profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.facebook", "options.0.scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection.facebook", "options.0.upstream_params", ""),
				),
			},
		},
	})
}

const testAccConnectionFacebookConfig = `
resource "auth0_connection" "facebook" {
	name = "Acceptance-Test-Facebook-{{.testName}}"
	is_domain_connection = false
	strategy = "facebook"
	options {
		client_id = "client_id"
		client_secret = "client_secret"
		scopes = [ "public_profile", "email", "groups_access_member_info", "user_birthday" ]
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

const testAccConnectionFacebookConfigUpdate = `
resource "auth0_connection" "facebook" {
	name = "Acceptance-Test-Facebook-{{.testName}}"
	is_domain_connection = false
	strategy = "facebook"
	options {
		client_id = "client_id_update"
		client_secret = "client_secret_update"
		scopes = [ "public_profile", "email" ]
	}
}
`

func TestAccConnectionApple(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionAppleConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.apple", "name", fmt.Sprintf("Acceptance-Test-Apple-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.apple", "strategy", "apple"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.client_id", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.client_secret", "-----BEGIN PRIVATE KEY-----\nMIHBAgEAMA0GCSqGSIb3DQEBAQUABIGsMIGpAgEAAiEA3+luhVHxSJ8cv3VNzQDP\nEL6BPs7FjBq4oro0MWM+QJMCAwEAAQIgWbq6/pRK4/ZXV+ZTSj7zuxsWZuK5i3ET\nfR2TCEkZR3kCEQD2ElqDr/pY5aHA++9HioY9AhEA6PIxC1c/K3gJqu+K+EsfDwIQ\nG5MS8Y7Wzv9skOOqfKnZQQIQdG24vaZZ2GwiyOD5YKiLWQIQYNtrb3j0BWsT4LI+\nN9+l1g==\n-----END PRIVATE KEY-----"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.team_id", "team_id"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.key_id", "key_id"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.apple", "options.0.scopes.*", "name"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.apple", "options.0.scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.set_user_root_attributes", "on_each_login"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
			{
				Config: template.ParseTestName(testAccConnectionAppleConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.team_id", "team_id_update"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.key_id", "key_id_update"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.apple", "options.0.scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.set_user_root_attributes", "on_first_login"),
					resource.TestCheckResourceAttr("auth0_connection.apple", "options.0.upstream_params", ""),
				),
			},
		},
	})
}

const testAccConnectionAppleConfig = `
resource "auth0_connection" "apple" {
	name = "Acceptance-Test-Apple-{{.testName}}"
	is_domain_connection = false
	strategy = "apple"
	options {
		client_id = "client_id"
		client_secret = "-----BEGIN PRIVATE KEY-----\nMIHBAgEAMA0GCSqGSIb3DQEBAQUABIGsMIGpAgEAAiEA3+luhVHxSJ8cv3VNzQDP\nEL6BPs7FjBq4oro0MWM+QJMCAwEAAQIgWbq6/pRK4/ZXV+ZTSj7zuxsWZuK5i3ET\nfR2TCEkZR3kCEQD2ElqDr/pY5aHA++9HioY9AhEA6PIxC1c/K3gJqu+K+EsfDwIQ\nG5MS8Y7Wzv9skOOqfKnZQQIQdG24vaZZ2GwiyOD5YKiLWQIQYNtrb3j0BWsT4LI+\nN9+l1g==\n-----END PRIVATE KEY-----"
		team_id = "team_id"
		key_id = "key_id"
		scopes = ["email", "name"]
		set_user_root_attributes = "on_each_login"
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

const testAccConnectionAppleConfigUpdate = `
resource "auth0_connection" "apple" {
	name = "Acceptance-Test-Apple-{{.testName}}"
	is_domain_connection = false
	strategy = "apple"
	options {
		client_id = "client_id"
		client_secret = "-----BEGIN PRIVATE KEY-----\nMIHBAgEAMA0GCSqGSIb3DQEBAQUABIGsMIGpAgEAAiEA3+luhVHxSJ8cv3VNzQDP\nEL6BPs7FjBq4oro0MWM+QJMCAwEAAQIgWbq6/pRK4/ZXV+ZTSj7zuxsWZuK5i3ET\nfR2TCEkZR3kCEQD2ElqDr/pY5aHA++9HioY9AhEA6PIxC1c/K3gJqu+K+EsfDwIQ\nG5MS8Y7Wzv9skOOqfKnZQQIQdG24vaZZ2GwiyOD5YKiLWQIQYNtrb3j0BWsT4LI+\nN9+l1g==\n-----END PRIVATE KEY-----"
		team_id = "team_id_update"
		key_id = "key_id_update"
		scopes = ["email"]
		set_user_root_attributes = "on_first_login"
	}
}
`

func TestAccConnectionLinkedin(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionLinkedinConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "name", fmt.Sprintf("Acceptance-Test-Linkedin-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "strategy", "linkedin"),
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "options.0.client_id", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "options.0.client_secret", "client_secret"),
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "options.0.strategy_version", "2"),
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "options.0.scopes.#", "3"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.linkedin", "options.0.scopes.*", "basic_profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.linkedin", "options.0.scopes.*", "email"),
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
			{
				Config: template.ParseTestName(testAccConnectionLinkedinConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "options.0.client_id", "client_id_update"),
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "options.0.client_secret", "client_secret_update"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.linkedin", "options.0.scopes.*", "basic_profile"),
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "options.0.scopes.#", "2"),
					resource.TestCheckResourceAttr("auth0_connection.linkedin", "options.0.upstream_params", ""),
				),
			},
		},
	})
}

const testAccConnectionLinkedinConfig = `
resource "auth0_connection" "linkedin" {
	name = "Acceptance-Test-Linkedin-{{.testName}}"
	is_domain_connection = false
	strategy = "linkedin"
	options {
		client_id = "client_id"
		client_secret = "client_secret"
		strategy_version = 2
		scopes = [ "basic_profile", "profile", "email" ]
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

const testAccConnectionLinkedinConfigUpdate = `
resource "auth0_connection" "linkedin" {
	name = "Acceptance-Test-Linkedin-{{.testName}}"
	is_domain_connection = false
	strategy = "linkedin"
	options {
		client_id = "client_id_update"
		client_secret = "client_secret_update"
		strategy_version = 2
		scopes = [ "basic_profile", "profile" ]
	}
}
`

func TestAccConnectionGitHub(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionGitHubConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.github", "name", fmt.Sprintf("Acceptance-Test-GitHub-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.github", "strategy", "github"),
					resource.TestCheckResourceAttr("auth0_connection.github", "options.0.client_id", "client-id"),
					resource.TestCheckResourceAttr("auth0_connection.github", "options.0.client_secret", "client-secret"),
					resource.TestCheckResourceAttr("auth0_connection.github", "options.0.scopes.#", "20"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "email"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "follow"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "read_repo_hook"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "admin_public_key"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "write_public_key"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "write_repo_hook"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "write_org"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "read_user"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "admin_repo_hook"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "admin_org"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "repo"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "repo_status"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "read_org"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "gist"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "repo_deployment"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "public_repo"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "notifications"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "delete_repo"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.github", "options.0.scopes.*", "read_public_key"),
					resource.TestCheckResourceAttr("auth0_connection.github", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
		},
	})
}

const testAccConnectionGitHubConfig = `
resource "auth0_connection" "github" {
	name = "Acceptance-Test-GitHub-{{.testName}}"
	strategy = "github"
	options {
		client_id = "client-id"
		client_secret = "client-secret"
		scopes = [ "email", "profile", "read_user", "follow", "public_repo", "repo", "repo_deployment", "repo_status",
				   "delete_repo", "notifications", "gist", "read_repo_hook", "write_repo_hook", "admin_repo_hook",
				   "read_org", "admin_org", "read_public_key", "write_public_key", "admin_public_key", "write_org"
		]
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

func TestAccConnectionWindowslive(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionWindowsliveConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "name", fmt.Sprintf("Acceptance-Test-Windowslive-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "strategy", "windowslive"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.client_id", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.client_secret", "client_secret"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.strategy_version", "2"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.windowslive", "options.0.scopes.*", "signin"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.windowslive", "options.0.scopes.*", "graph_user"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
			{
				Config: template.ParseTestName(testAccConnectionWindowsliveConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "name", fmt.Sprintf("Acceptance-Test-Windowslive-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "strategy", "windowslive"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.client_id", "client_id_update"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.client_secret", "client_secret_update"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.strategy_version", "2"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.scopes.#", "1"),
					resource.TestCheckTypeSetElemAttr("auth0_connection.windowslive", "options.0.scopes.*", "signin"),
					resource.TestCheckResourceAttr("auth0_connection.windowslive", "options.0.upstream_params", ""),
				),
			},
		},
	})
}

const testAccConnectionWindowsliveConfig = `
resource "auth0_connection" "windowslive" {
	name = "Acceptance-Test-Windowslive-{{.testName}}"
	is_domain_connection = false
	strategy = "windowslive"
	options {
		client_id = "client_id"
		client_secret = "client_secret"
		strategy_version = 2
		scopes = ["signin", "graph_user"]
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
	}
}
`

const testAccConnectionWindowsliveConfigUpdate = `
resource "auth0_connection" "windowslive" {
	name = "Acceptance-Test-Windowslive-{{.testName}}"
	is_domain_connection = false
	strategy = "windowslive"
	options {
		client_id = "client_id_update"
		client_secret = "client_secret_update"
		strategy_version = 2
		scopes = ["signin"]
	}
}
`

func TestAccConnectionConfiguration(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccConnectionConfigurationCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.configuration.%", "2"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.configuration.foo", "xxx"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.configuration.bar", "zzz"),
				),
			},
			{
				Config: template.ParseTestName(testAccConnectionConfigurationUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.configuration.%", "3"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.configuration.foo", "xxx"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.configuration.bar", "yyy"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.configuration.baz", "zzz"),
				),
			},
		},
	})
}

const testAccConnectionConfigurationCreate = `
resource "auth0_connection" "my_connection" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	is_domain_connection = true
	strategy = "auth0"
	options {
		brute_force_protection = true
		configuration = {
			foo = "xxx"
			bar = "zzz"
		}
	}
}
`

const testAccConnectionConfigurationUpdate = `
resource "auth0_connection" "my_connection" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	is_domain_connection = true
	strategy = "auth0"
	options {
		brute_force_protection = true
		configuration = {
			foo = "xxx"
			bar = "yyy"
			baz = "zzz"
		}
	}
}
`

func TestConnectionInstanceStateUpgradeV0(t *testing.T) {
	for _, tt := range []struct {
		name            string
		version         interface{}
		versionExpected int
	}{
		{
			name:            "Empty",
			version:         "",
			versionExpected: 0,
		},
		{
			name:            "Zero",
			version:         "0",
			versionExpected: 0,
		},
		{
			name:            "NonZero",
			version:         "123",
			versionExpected: 123,
		},
		{
			name:            "Invalid",
			version:         "foo",
			versionExpected: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := map[string]interface{}{
				"options": []interface{}{
					map[string]interface{}{"strategy_version": tt.version},
				},
			}

			actual, err := connectionSchemaUpgradeV0(context.Background(), state, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}

			expected := map[string]interface{}{
				"options": []interface{}{
					map[string]interface{}{"strategy_version": tt.versionExpected},
				},
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
			}
		})
	}
}

func TestConnectionInstanceStateUpgradeV1(t *testing.T) {
	for _, tt := range []struct {
		name               string
		validation         map[string]string
		validationExpected []map[string][]interface{}
	}{
		{
			name: "Only Min",
			validation: map[string]string{
				"min": "5",
			},
			validationExpected: []map[string][]interface{}{
				{
					"username": []interface{}{
						map[string]string{
							"min": "5",
						},
					},
				},
			},
		},
		{
			name: "Min and Max",
			validation: map[string]string{
				"min": "5",
				"max": "10",
			},
			validationExpected: []map[string][]interface{}{
				{
					"username": []interface{}{
						map[string]string{
							"min": "5",
							"max": "10",
						},
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := map[string]interface{}{
				"options": []interface{}{
					map[string]interface{}{"validation": tt.validation},
				},
			}

			actual, err := connectionSchemaUpgradeV1(context.Background(), state, nil)
			if err != nil {
				t.Fatalf("error migrating state: %s", err)
			}

			expected := map[string]interface{}{
				"options": []interface{}{
					map[string]interface{}{"validation": tt.validationExpected},
				},
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
			}
		})
	}
}

func TestAccConnectionSAML(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testConnectionSAMLConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "name", fmt.Sprintf("Acceptance-Test-SAML-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "display_name", fmt.Sprintf("Acceptance-Test-SAML-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "strategy", "samlp"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "show_as_button", "false"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.#", "1"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.sign_out_endpoint", "https://saml.provider/sign_out"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.disable_sign_out", "false"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.entity_id", ""),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.idp_initiated.0.client_authorize_query", "type=code&timeout=30"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.fields_map", "{\"email\":[\"emailaddress\",\"nameidentifier\"],\"family_name\":\"surname\",\"name\":[\"name\",\"nameidentifier\"]}"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.metadata_url", ""),
					resource.TestCheckResourceAttrSet("auth0_connection.my_connection", "options.0.metadata_xml"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.signing_key.#", "1"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.signing_key.0.cert", "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.signing_key.0.key", "-----BEGIN PRIVATE KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
				),
			},
			{
				Config: template.ParseTestName(testConnectionSAMLConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "show_as_button", "true"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.#", "1"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.idp_initiated.0.client_authorize_query", "type=code&timeout=60"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.sign_out_endpoint", "https://saml.provider/sign_out"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.disable_sign_out", "true"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.entity_id", "example"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.fields_map", "{\"email\":[\"emailaddress\",\"nameidentifier\"],\"family_name\":\"appelido\",\"name\":[\"name\"]}"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.metadata_url", "https://raw.githubusercontent.com/auth0/terraform-provider-auth0/a51c2f52877c26a00e7a3e67ca56aff00be18762/auth0/testdata/saml_metadata.xml"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.metadata_xml", ""),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.signing_key.#", "0"),
					resource.TestCheckResourceAttr("auth0_connection.my_connection", "options.0.upstream_params", ""),
				),
			},
		},
	})
}

const testConnectionSAMLConfigCreate = `
resource "auth0_connection" "my_connection" {
	name = "Acceptance-Test-SAML-{{.testName}}"
	display_name = "Acceptance-Test-SAML-{{.testName}}"
	strategy = "samlp"
	show_as_button = false
	options {
		signing_cert = <<EOF
-----BEGIN CERTIFICATE-----
MIID6TCCA1ICAQEwDQYJKoZIhvcNAQEFBQAwgYsxCzAJBgNVBAYTAlVTMRMwEQYD
VQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRQwEgYDVQQK
EwtHb29nbGUgSW5jLjEMMAoGA1UECxMDRW5nMQwwCgYDVQQDEwNhZ2wxHTAbBgkq
hkiG9w0BCQEWDmFnbEBnb29nbGUuY29tMB4XDTA5MDkwOTIyMDU0M1oXDTEwMDkw
OTIyMDU0M1owajELMAkGA1UEBhMCQVUxEzARBgNVBAgTClNvbWUtU3RhdGUxITAf
BgNVBAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDEjMCEGA1UEAxMaZXVyb3Bh
LnNmby5jb3JwLmdvb2dsZS5jb20wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIK
AoICAQC6pgYt7/EibBDumASF+S0qvqdL/f+nouJw2T1Qc8GmXF/iiUcrsgzh/Fd8
pDhz/T96Qg9IyR4ztuc2MXrmPra+zAuSf5bevFReSqvpIt8Duv0HbDbcqs/XKPfB
uMDe+of7a9GCywvAZ4ZUJcp0thqD9fKTTjUWOBzHY1uNE4RitrhmJCrbBGXbJ249
bvgmb7jgdInH2PU7PT55hujvOoIsQW2osXBFRur4pF1wmVh4W4lTLD6pjfIMUcML
ICHEXEN73PDic8KS3EtNYCwoIld+tpIBjE1QOb1KOyuJBNW6Esw9ALZn7stWdYcE
qAwvv20egN2tEXqj7Q4/1ccyPZc3PQgC3FJ8Be2mtllM+80qf4dAaQ/fWvCtOrQ5
pnfe9juQvCo8Y0VGlFcrSys/MzSg9LJ/24jZVgzQved/Qupsp89wVidwIzjt+WdS
fyWfH0/v1aQLvu5cMYuW//C0W2nlYziL5blETntM8My2ybNARy3ICHxCBv2RNtPI
WQVm+E9/W5rwh2IJR4DHn2LHwUVmT/hHNTdBLl5Uhwr4Wc7JhE7AVqb14pVNz1lr
5jxsp//ncIwftb7mZQ3DF03Yna+jJhpzx8CQoeLT6aQCHyzmH68MrHHT4MALPyUs
Pomjn71GNTtDeWAXibjCgdL6iHACCF6Htbl0zGlG0OAK+bdn0QIDAQABMA0GCSqG
SIb3DQEBBQUAA4GBAOKnQDtqBV24vVqvesL5dnmyFpFPXBn3WdFfwD6DzEb21UVG
5krmJiu+ViipORJPGMkgoL6BjU21XI95VQbun5P8vvg8Z+FnFsvRFY3e1CCzAVQY
ZsUkLw2I7zI/dNlWdB8Xp7v+3w9sX5N3J/WuJ1KOO5m26kRlHQo7EzT3974g
-----END CERTIFICATE-----
EOF
		signing_key {
			key = "-----BEGIN PRIVATE KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"
       		cert = "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"
		}
		sign_in_endpoint = "https://saml.provider/sign_in"
		sign_out_endpoint = "https://saml.provider/sign_out"
		disable_sign_out = false
		user_id_attribute = "https://saml.provider/imi/ns/identity-200810"
		tenant_domain = "example.com"
		domain_aliases = ["example.com", "example.coz"]
		protocol_binding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
		request_template = "<samlp:AuthnRequest xmlns:samlp=\"urn:oasis:names:tc:SAML:2.0:protocol\"\n@@AssertServiceURLAndDestination@@\n    ID=\"@@ID@@\"\n    IssueInstant=\"@@IssueInstant@@\"\n    ProtocolBinding=\"@@ProtocolBinding@@\" Version=\"2.0\">\n    <saml:Issuer xmlns:saml=\"urn:oasis:names:tc:SAML:2.0:assertion\">@@Issuer@@</saml:Issuer>\n</samlp:AuthnRequest>"
		signature_algorithm = "rsa-sha256"
		digest_algorithm = "sha256"
		icon_url = "https://example.com/logo.svg"
		fields_map = jsonencode({
			"name": ["name", "nameidentifier"]
			"email": ["emailaddress", "nameidentifier"]
			"family_name": "surname"
		})
		upstream_params = jsonencode({
			"screen_name": {
				"alias": "login_hint"
			}
		})
		idp_initiated {
			client_id = "client_id"
			client_protocol = "samlp"
			client_authorize_query = "type=code&timeout=30"
		}
		metadata_xml = <<EOF
<?xml version="1.0"?>
<md:EntityDescriptor xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" entityID="https://example.com">
  <md:IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://saml.provider/sign_out"/>
    <md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://saml.provider/sign_in"/>
  </md:IDPSSODescriptor>
</md:EntityDescriptor>
EOF
	}
}
`

const testConnectionSAMLConfigUpdate = `
resource "auth0_connection" "my_connection" {
	name = "Acceptance-Test-SAML-{{.testName}}"
	display_name = "Acceptance-Test-SAML-{{.testName}}"
	strategy = "samlp"
	show_as_button = true
	options {
		signing_cert = <<EOF
-----BEGIN CERTIFICATE-----
MIID6TCCA1ICAQEwDQYJKoZIhvcNAQEFBQAwgYsxCzAJBgNVBAYTAlVTMRMwEQYD
VQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRQwEgYDVQQK
EwtHb29nbGUgSW5jLjEMMAoGA1UECxMDRW5nMQwwCgYDVQQDEwNhZ2wxHTAbBgkq
hkiG9w0BCQEWDmFnbEBnb29nbGUuY29tMB4XDTA5MDkwOTIyMDU0M1oXDTEwMDkw
OTIyMDU0M1owajELMAkGA1UEBhMCQVUxEzARBgNVBAgTClNvbWUtU3RhdGUxITAf
BgNVBAoTGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDEjMCEGA1UEAxMaZXVyb3Bh
LnNmby5jb3JwLmdvb2dsZS5jb20wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIK
AoICAQC6pgYt7/EibBDumASF+S0qvqdL/f+nouJw2T1Qc8GmXF/iiUcrsgzh/Fd8
pDhz/T96Qg9IyR4ztuc2MXrmPra+zAuSf5bevFReSqvpIt8Duv0HbDbcqs/XKPfB
uMDe+of7a9GCywvAZ4ZUJcp0thqD9fKTTjUWOBzHY1uNE4RitrhmJCrbBGXbJ249
bvgmb7jgdInH2PU7PT55hujvOoIsQW2osXBFRur4pF1wmVh4W4lTLD6pjfIMUcML
ICHEXEN73PDic8KS3EtNYCwoIld+tpIBjE1QOb1KOyuJBNW6Esw9ALZn7stWdYcE
qAwvv20egN2tEXqj7Q4/1ccyPZc3PQgC3FJ8Be2mtllM+80qf4dAaQ/fWvCtOrQ5
pnfe9juQvCo8Y0VGlFcrSys/MzSg9LJ/24jZVgzQved/Qupsp89wVidwIzjt+WdS
fyWfH0/v1aQLvu5cMYuW//C0W2nlYziL5blETntM8My2ybNARy3ICHxCBv2RNtPI
WQVm+E9/W5rwh2IJR4DHn2LHwUVmT/hHNTdBLl5Uhwr4Wc7JhE7AVqb14pVNz1lr
5jxsp//ncIwftb7mZQ3DF03Yna+jJhpzx8CQoeLT6aQCHyzmH68MrHHT4MALPyUs
Pomjn71GNTtDeWAXibjCgdL6iHACCF6Htbl0zGlG0OAK+bdn0QIDAQABMA0GCSqG
SIb3DQEBBQUAA4GBAOKnQDtqBV24vVqvesL5dnmyFpFPXBn3WdFfwD6DzEb21UVG
5krmJiu+ViipORJPGMkgoL6BjU21XI95VQbun5P8vvg8Z+FnFsvRFY3e1CCzAVQY
ZsUkLw2I7zI/dNlWdB8Xp7v+3w9sX5N3J/WuJ1KOO5m26kRlHQo7EzT3974g
-----END CERTIFICATE-----
EOF
		sign_in_endpoint = "https://saml.provider/sign_in"
		sign_out_endpoint = "https://saml.provider/sign_out"
		disable_sign_out = true
		tenant_domain = "example.com"
		domain_aliases = ["example.com", "example.coz"]
		protocol_binding = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
		signature_algorithm = "rsa-sha256"
		digest_algorithm = "sha256"
		entity_id = "example"
		fields_map = jsonencode({
			"name": ["name"]
			"email": ["emailaddress", "nameidentifier"]
			"family_name": "appelido"
		})
		idp_initiated {
			client_id = "client_id"
			client_protocol = "samlp"
			client_authorize_query = "type=code&timeout=60"
		}
		metadata_url = "https://raw.githubusercontent.com/auth0/terraform-provider-auth0/a51c2f52877c26a00e7a3e67ca56aff00be18762/auth0/testdata/saml_metadata.xml"
	}
}
`
