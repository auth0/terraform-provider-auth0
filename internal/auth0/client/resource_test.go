package client_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccClientValidationOnInitiateLoginURIWithHTTP = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Initiate Login URI - {{.testName}}"
	initiate_login_uri = "http://example.com/login"
}
`

func TestAccClientInitiateLoginUriValidation(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccClientValidationOnInitiateLoginURIWithHTTP, t.Name()),
				ExpectError: regexp.MustCompile("to have a url with schema"),
			},
		},
	})
}

const testAccClientConfigRotateSecret = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Rotate Secret - {{.testName}}"
}
`

const testAccClientConfigRotateSecretUpdate = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Rotate Secret - {{.testName}}"

	client_secret_rotation_trigger = {
		triggered_at = "2021-10-01T23:12:01Z"
		triggered_by = "dx-cdt"
	}
}
`

func TestAccClientRotateSecret(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccClientConfigRotateSecret, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Rotate Secret - %s", t.Name())),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "client_secret_rotation_trigger"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientConfigRotateSecretUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Rotate Secret - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_secret_rotation_trigger.triggered_at", "2021-10-01T23:12:01Z"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_secret_rotation_trigger.triggered_by", "dx-cdt"),
				),
			},
		},
	})
}

const testAccClientValidationOnMobile = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Mobile - {{.testName}}"
	mobile {
		android {
			# nothing specified, should throw validation error
		}
	}
}
`

func TestAccClientMobileValidationError(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccClientValidationOnMobile, t.Name()),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
		},
	})
}

const testAccCreateMobileClient = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Mobile - {{.testName}}"
	app_type = "native"

	mobile {
		android {
			app_package_name = "com.example"
			sha256_cert_fingerprints = ["DE:AD:BE:EF"]
		}

		ios {
			team_id = "9JA89QQLNQ"
			app_bundle_identifier = "com.my.bundle.id"
		}
	}

	native_social_login {
		apple {
			enabled = true
		}

		facebook {
			enabled = false
		}
	}
}
`

const testAccUpdateMobileClient = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Mobile - {{.testName}}"
	app_type = "native"

	mobile {
		android {
			app_package_name = "com.example"
			sha256_cert_fingerprints = ["DE:AD:BE:EF", "CA:DE:FF:AA"]
		}

		ios {
			team_id = "1111111111"
			app_bundle_identifier = "com.my.auth0.bundle"
		}
	}

	native_social_login {
		apple {
			enabled = false
		}

		facebook {
			enabled = true
		}
	}
}
`

const testAccUpdateMobileClientAgainByRemovingSomeFields = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Mobile - {{.testName}}"
	app_type = "native"

	mobile {
		android {
			app_package_name = "com.example"
			sha256_cert_fingerprints = ["DE:AD:BE:EF", "CA:DE:FF:AA"]
		}
	}

	native_social_login {
		facebook {
			enabled = false
		}
	}
}
`

const testAccChangeMobileClientToM2M = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Mobile - {{.testName}}"
	app_type = "non_interactive"

	native_social_login {
		apple {
			enabled = false
		}

		facebook {
			enabled = false
		}
	}
}
`

func TestAccClientMobile(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateMobileClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Mobile - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "native"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.app_package_name", "com.example"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.0", "DE:AD:BE:EF"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.0.team_id", "9JA89QQLNQ"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.0.app_bundle_identifier", "com.my.bundle.id"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.0.enabled", "false"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateMobileClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Mobile - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "native"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.app_package_name", "com.example"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.#", "2"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.0", "DE:AD:BE:EF"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.1", "CA:DE:FF:AA"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.0.team_id", "1111111111"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.0.app_bundle_identifier", "com.my.auth0.bundle"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.0.enabled", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateMobileClientAgainByRemovingSomeFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Mobile - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "native"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.app_package_name", "com.example"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.#", "2"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.0", "DE:AD:BE:EF"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.1", "CA:DE:FF:AA"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.0.team_id", "1111111111"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.0.app_bundle_identifier", "com.my.auth0.bundle"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.0.enabled", "false"),
				),
			},
			{
				// This just makes sure that we can change the app type.
				//
				// To note also that we can't reset mobile to empty.
				// We need a different approach or wait until the API behaves differently.
				Config: acctest.ParseTestName(testAccChangeMobileClientToM2M, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Mobile - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.app_package_name", "com.example"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.#", "2"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.0", "DE:AD:BE:EF"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.1", "CA:DE:FF:AA"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.0.team_id", "1111111111"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.ios.0.app_bundle_identifier", "com.my.auth0.bundle"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.0.enabled", "false"),
				),
			},
		},
	})
}

const testAccCreateClientWithRefreshToken = `
resource "auth0_client" "my_client" {
	name      = "Acceptance Test - Refresh Token - {{.testName}}"
	app_type  = "spa"

	refresh_token {
		rotation_type   = "non-rotating"
		expiration_type = "non-expiring"

		# Intentionally not setting leeway,
		# token_lifetime, infinite_token_lifetime,
		# infinite_idle_token_lifetime,
		# idle_token_lifetime because those get
		# inferred by Auth0 defaults.
	}
}
`

const testAccUpdateClientWithRefreshToken = `
resource "auth0_client" "my_client" {
	name      = "Acceptance Test - Refresh Token - {{.testName}}"
	app_type  = "spa"

	refresh_token {
		rotation_type   = "non-rotating"
		expiration_type = "non-expiring"
		leeway = 60
		token_lifetime = 256000
		infinite_token_lifetime = true
		infinite_idle_token_lifetime = true
		idle_token_lifetime = 128000
	}
}
`

const testAccUpdateClientWithRefreshTokenWhenRemovedFromConfig = `
resource "auth0_client" "my_client" {
	name      = "Acceptance Test - Refresh Token - {{.testName}}"
	app_type  = "spa"
}
`

func TestAccClientRefreshToken(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateClientWithRefreshToken, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Refresh Token - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "spa"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.rotation_type", "non-rotating"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.expiration_type", "non-expiring"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "refresh_token.0.leeway"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "refresh_token.0.token_lifetime"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "refresh_token.0.infinite_token_lifetime"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "refresh_token.0.infinite_idle_token_lifetime"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "refresh_token.0.idle_token_lifetime"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithRefreshToken, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Refresh Token - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "spa"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.rotation_type", "non-rotating"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.expiration_type", "non-expiring"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.leeway", "60"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.token_lifetime", "256000"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_idle_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.idle_token_lifetime", "128000"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithRefreshTokenWhenRemovedFromConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Refresh Token - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "spa"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.rotation_type", "non-rotating"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.expiration_type", "non-expiring"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.leeway", "60"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.token_lifetime", "256000"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_idle_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.idle_token_lifetime", "128000"),
				),
			},
		},
	})
}

const testAccCreateClientWithJWTConfiguration = `
resource "auth0_client" "my_client" {
	name      = "Acceptance Test - JWT Config - {{.testName}}"
	app_type  = "non_interactive"

	jwt_configuration {}
}
`

const testAccUpdateClientWithJWTConfiguration = `
resource "auth0_client" "my_client" {
	name      = "Acceptance Test - JWT Config - {{.testName}}"
	app_type  = "non_interactive"

	jwt_configuration {
		lifetime_in_seconds = 300
		secret_encoded = true
		alg = "RS256"
		scopes = {
			foo = "bar"
		}
	}
}
`

const testAccUpdateClientWithJWTConfigurationEmpty = `
resource "auth0_client" "my_client" {
	name      = "Acceptance Test - JWT Config - {{.testName}}"
	app_type  = "non_interactive"

	jwt_configuration {
		lifetime_in_seconds = 1
		secret_encoded = false
		alg = "RS256"
		scopes = {}
	}
}
`

const testAccUpdateClientWithJWTConfigurationRemoved = `
resource "auth0_client" "my_client" {
	name      = "Acceptance Test - JWT Config - {{.testName}}"
	app_type  = "non_interactive"
}
`

func TestAccClientJWTConfiguration(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateClientWithJWTConfiguration, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - JWT Config - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.lifetime_in_seconds", "36000"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.%", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.secret_encoded", "false"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithJWTConfiguration, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - JWT Config - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.lifetime_in_seconds", "300"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.secret_encoded", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.%", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.foo", "bar"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithJWTConfigurationEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - JWT Config - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.lifetime_in_seconds", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.secret_encoded", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.%", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithJWTConfigurationRemoved, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - JWT Config - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.lifetime_in_seconds", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.secret_encoded", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.%", "0"),
				),
			},
		},
	})
}

const testAccClientConfigCreateWithOnlyRequiredFields = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - {{.testName}}"
}
`

const testAccClientConfigUpdateAllFields = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - {{.testName}}"
	app_type = "non_interactive"
	description = "Test Application Long Description"
	cross_origin_loc = "https://example.com/cross-origin-loc"
	custom_login_page = "test"
	form_template = "test"
	token_endpoint_auth_method = "client_secret_post"
	initiate_login_uri = "https://example.com/login"
	logo_uri = "https://example.com/logoUri"
	organization_require_behavior = "no_prompt"
	organization_usage = "deny"
	sso = false
	sso_disabled = false
	custom_login_page_on = true
	is_first_party = true
	is_token_endpoint_ip_header_trusted = true
	oidc_conformant = true
	client_aliases = [ "https://example.com/audience" ]
	callbacks = [ "https://example.com/callback" ]
	allowed_origins = [ "https://example.com" ]
	allowed_clients = [ "https://allowed.example.com" ]
	grant_types = [ "authorization_code", "http://auth0.com/oauth/grant-type/password-realm", "implicit", "password", "refresh_token" ]
	allowed_logout_urls = [ "https://example.com" ]
	oidc_backchannel_logout_urls = [ "https://example.com/oidc-logout" ]
	web_origins = [ "https://example.com" ]
	client_metadata = {
		foo = "zoo"
	}
}
`

const testAccClientConfigUpdateSomeFieldsToEmpty = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - {{.testName}}"
	app_type = "non_interactive"
	description = ""
	cross_origin_loc = "https://example.com/cross-origin-loc"
	custom_login_page = ""
	form_template = ""
	token_endpoint_auth_method = "client_secret_post"
	initiate_login_uri = ""
	logo_uri = "https://another-example.com/logoUri"
	organization_require_behavior = "no_prompt"
	organization_usage = "deny"
	sso = true
	sso_disabled = true
	custom_login_page_on = true
	is_first_party = true
	is_token_endpoint_ip_header_trusted = true
	oidc_conformant = true
	client_aliases = [ ]
	callbacks = [ ]
	allowed_origins = [ ]
	allowed_clients = [ ]
	grant_types = [ ]
	allowed_logout_urls = [ ]
	web_origins = [ ]
	client_metadata = {}
	oidc_backchannel_logout_urls = []
}
`

func TestAccClient(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccClientConfigCreateWithOnlyRequiredFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "client_id"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "description", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "cross_origin_loc", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "custom_login_page", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "form_template", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "token_endpoint_auth_method", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "initiate_login_uri", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "logo_uri", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "organization_require_behavior", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "organization_usage", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "sso", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "sso_disabled", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "custom_login_page_on", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_first_party", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_token_endpoint_ip_header_trusted", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "oidc_conformant", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "cross_origin_auth", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.#", "0"),

					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "signing_keys.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.#", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.0", "authorization_code"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.1", "implicit"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.2", "refresh_token"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.3", "client_credentials"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.lifetime_in_seconds", "36000"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.%", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.secret_encoded", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.%", "7"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.expiration_type", "non-expiring"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.idle_token_lifetime", "1296000"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_idle_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.leeway", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.rotation_type", "non-rotating"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.token_lifetime", "2592000"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "client_secret_rotation_trigger"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_aliases.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "callbacks.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_logout_urls.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_origins.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_clients.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "web_origins.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "encryption_key.%", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.%", "0"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "oidc_backchannel_logout_urls"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientConfigUpdateAllFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "client_id"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "description", "Test Application Long Description"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "cross_origin_loc", "https://example.com/cross-origin-loc"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "custom_login_page", "test"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "form_template", "test"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "token_endpoint_auth_method", "client_secret_post"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "initiate_login_uri", "https://example.com/login"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "logo_uri", "https://example.com/logoUri"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "organization_require_behavior", "no_prompt"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "organization_usage", "deny"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "sso", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "sso_disabled", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "custom_login_page_on", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_first_party", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_token_endpoint_ip_header_trusted", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "oidc_conformant", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "cross_origin_auth", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.#", "0"),

					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "signing_keys.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.#", "5"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.0", "authorization_code"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.1", "http://auth0.com/oauth/grant-type/password-realm"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.2", "implicit"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.3", "password"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.4", "refresh_token"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.lifetime_in_seconds", "36000"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.%", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.secret_encoded", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.%", "7"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.expiration_type", "non-expiring"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.idle_token_lifetime", "1296000"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_idle_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.leeway", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.rotation_type", "non-rotating"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.token_lifetime", "2592000"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "client_secret_rotation_trigger"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_aliases.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_aliases.0", "https://example.com/audience"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "callbacks.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "callbacks.0", "https://example.com/callback"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_logout_urls.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_logout_urls.0", "https://example.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_origins.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_origins.0", "https://example.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_clients.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_clients.0", "https://allowed.example.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "web_origins.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "web_origins.0", "https://example.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.%", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.foo", "zoo"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "encryption_key.%", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "oidc_backchannel_logout_urls.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientConfigUpdateSomeFieldsToEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "client_id"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "description", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "cross_origin_loc", "https://example.com/cross-origin-loc"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "custom_login_page", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "form_template", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "token_endpoint_auth_method", "client_secret_post"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "initiate_login_uri", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "logo_uri", "https://another-example.com/logoUri"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "organization_require_behavior", "no_prompt"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "organization_usage", "deny"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "sso", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "sso_disabled", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "custom_login_page_on", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_first_party", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_token_endpoint_ip_header_trusted", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "oidc_conformant", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "cross_origin_auth", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.#", "0"),

					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "signing_keys.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "grant_types.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", ""),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.lifetime_in_seconds", "36000"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.%", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.secret_encoded", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.%", "7"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.expiration_type", "non-expiring"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.idle_token_lifetime", "1296000"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_idle_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.leeway", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.rotation_type", "non-rotating"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.token_lifetime", "2592000"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "client_secret_rotation_trigger"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_aliases.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "callbacks.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_logout_urls.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_origins.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_clients.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "web_origins.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.%", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "encryption_key.%", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "oidc_backchannel_logout_urls.#", "0"),
				),
			},
		},
	})
}

const testAccCreateClientWithAddonsAWS = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		aws {
			principal           = "arn:aws:iam::010616021751:saml-provider/idpname"
			role                = "arn:aws:iam::010616021751:role/foo"
			lifetime_in_seconds = 32000
		}
	}
}
`

const testAccUpdateClientWithAddonsAzureBlob = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		azure_blob {
			account_name       = "acmeorg"
			storage_access_key = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa=="
			container_name     = "my-container"
			blob_name          = "my-blob"
			expiration         = 10
			signed_identifier  = "id123"
			blob_read          = true
			blob_write         = true
			blob_delete        = true
			container_read     = true
			container_write    = true
			container_delete   = true
			container_list     = true
		}
	}
}
`

const testAccUpdateClientWithAddonsAzureSB = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		azure_sb {
			namespace    = "acmeorg"
			sas_key_name = "my-policy"
			sas_key      = "my-key"
			entity_path  = "my-queue"
			expiration   = 10
		}
	}
}
`

const testAccUpdateClientWithAddonsRMS = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		rms {
			url = "https://example.com"
		}
	}
}
`

const testAccUpdateClientWithAddonsMSCRM = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		mscrm {
			url = "https://example.com"
		}
	}
}
`

const testAccUpdateClientWithAddonsSlack = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		slack {
			team = "acmeorg"
		}
	}
}
`

const testAccUpdateClientWithAddonsSentry = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		sentry {
			org_slug = "acmeorg"
			base_url = ""
		}
	}
}
`

const testAccUpdateClientWithAddonsEchoSign = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		echosign {
			domain = "acmeorg"
		}
	}
}
`

const testAccUpdateClientWithAddonsEgnyte = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		egnyte {
			domain = "acmeorg"
		}
	}
}
`

const testAccUpdateClientWithAddonsFirebase = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		firebase {
			secret              = "secret"
			private_key_id      = "private-key-id"
			private_key         = "private-key"
			client_email        = "service-account"
			lifetime_in_seconds = 7200
		}
	}
}
`

const testAccUpdateClientWithAddonsNewRelic = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		newrelic {
			account = "123456"
		}
	}
}
`

const testAccUpdateClientWithAddonsOffice365 = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		office365 {
			domain     = "acmeorg"
			connection = "Username-Password-Authentication"
		}
	}
}
`

const testAccUpdateClientWithAddonsSalesforce = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		salesforce {
			entity_id = "https://acme-org.com"
		}

		salesforce_api {
			client_id             = "client-id"
			principal             = "principal"
			community_name        = "community-name"
			community_url_section = "community-url-section"
		}

		salesforce_sandbox_api {
			client_id             = "client-id"
			principal             = "principal"
			community_name        = "community-name"
			community_url_section = "community-url-section"
		}
	}
}
`

const testAccUpdateClientWithAddonsLayer = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		layer {
			provider_id = "provider-id"
			key_id      = "key-id"
			private_key = "private-key"
			principal   = "principal"
			expiration  = 10
		}
	}
}
`

func TestAccClientAddons(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateClientWithAddonsAWS, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.aws.0.principal", "arn:aws:iam::010616021751:saml-provider/idpname"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.aws.0.role", "arn:aws:iam::010616021751:role/foo"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.aws.0.lifetime_in_seconds", "32000"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsAzureBlob, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.account_name", "acmeorg"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.storage_access_key", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa=="),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.container_name", "my-container"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.blob_name", "my-blob"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.expiration", "10"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.signed_identifier", "id123"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.blob_read", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.blob_write", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.blob_delete", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.container_read", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.container_write", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.container_delete", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_blob.0.container_list", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsAzureSB, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_sb.0.namespace", "acmeorg"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_sb.0.sas_key_name", "my-policy"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_sb.0.sas_key", "my-key"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_sb.0.entity_path", "my-queue"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.azure_sb.0.expiration", "10"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsRMS, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.rms.0.url", "https://example.com"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsMSCRM, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.mscrm.0.url", "https://example.com"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsSlack, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.slack.0.team", "acmeorg"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsSentry, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.sentry.0.org_slug", "acmeorg"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.sentry.0.base_url", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsEchoSign, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.echosign.0.domain", "acmeorg"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsEgnyte, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.egnyte.0.domain", "acmeorg"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsFirebase, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.0.secret", "secret"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.0.private_key_id", "private-key-id"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.0.private_key", "private-key"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.0.client_email", "service-account"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.0.lifetime_in_seconds", "7200"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsNewRelic, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.newrelic.0.account", "123456"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsOffice365, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.office365.0.domain", "acmeorg"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.office365.0.connection", "Username-Password-Authentication"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsSalesforce, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce.0.entity_id", "https://acme-org.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce_api.0.client_id", "client-id"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce_api.0.principal", "principal"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce_api.0.community_name", "community-name"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce_api.0.community_url_section", "community-url-section"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce.0.entity_id", "https://acme-org.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce_sandbox_api.0.client_id", "client-id"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce_sandbox_api.0.principal", "principal"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce_sandbox_api.0.community_name", "community-name"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.salesforce_sandbox_api.0.community_url_section", "community-url-section"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateClientWithAddonsLayer, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.layer.0.provider_id", "provider-id"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.layer.0.key_id", "key-id"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.layer.0.private_key", "private-key"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.layer.0.principal", "principal"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.layer.0.expiration", "10"),
				),
			},
		},
	})
}

func TestAccClientMetadataBehavior(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(`
					resource "auth0_client" "my_client" {
						name = "Acceptance Test - Metadata - {{.testName}}"
						client_metadata = {
							foo = "zoo"
							bar = "baz"
						}
					}`, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Metadata - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.%", "2"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.foo", "zoo"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.bar", "baz"),
				),
			},
			{
				Config: acctest.ParseTestName(`
					resource "auth0_client" "my_client" {
						name = "Acceptance Test - Metadata - {{.testName}}"
						client_metadata = {
							foo = "newZooButOldFoo"
							newBar = "newBaz"
						}
					}`, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Metadata - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.%", "2"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.foo", "newZooButOldFoo"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.newBar", "newBaz"),
				),
			},
			{
				Config: acctest.ParseTestName(`
					resource "auth0_client" "my_client" {
						name = "Acceptance Test - Metadata - {{.testName}}"
						client_metadata = {
							bar = "baz"
						}
					}`, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Metadata - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.%", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.bar", "baz"),
				),
			},
			{
				Config: acctest.ParseTestName(`
					resource "auth0_client" "my_client" {
						name = "Acceptance Test - Metadata - {{.testName}}"
						client_metadata = { }
					}`, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Metadata - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.%", "0"),
				),
			},
		},
	})
}
