package provider

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

func init() {
	resource.AddTestSweepers("auth0_client", &resource.Sweeper{
		Name: "auth0_client",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				clientList, err := api.Client.List(management.Page(page))
				if err != nil {
					return err
				}

				for _, client := range clientList.Clients {
					log.Printf("[DEBUG] ➝ %s", client.GetName())

					if strings.Contains(client.GetName(), "Test") {
						result = multierror.Append(
							result,
							api.Client.Delete(client.GetClientID()),
						)
						log.Printf("[DEBUG] ✗ %s", client.GetName())
					}
				}
				if !clientList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}

const testAccClientValidationOnInitiateLoginURIWithHTTP = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Initiate Login URI - {{.testName}}"
	initiate_login_uri = "http://example.com/login"
}
`

const testAccClientValidationOnInitiateLoginURIWithFragment = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Initiate Login URI - {{.testName}}"
	initiate_login_uri = "https://example.com/login#fragment"
}
`

func TestAccClientInitiateLoginUriValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(nil),
		Steps: []resource.TestStep{
			{
				Config:      template.ParseTestName(testAccClientValidationOnInitiateLoginURIWithHTTP, t.Name()),
				ExpectError: regexp.MustCompile("to have a url with schema"),
			},
			{
				Config:      template.ParseTestName(testAccClientValidationOnInitiateLoginURIWithFragment, t.Name()),
				ExpectError: regexp.MustCompile("to have a url with an empty fragment"),
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
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfigRotateSecret, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Rotate Secret - %s", t.Name())),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "client_secret_rotation_trigger"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientConfigRotateSecretUpdate, t.Name()),
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
	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(nil),
		Steps: []resource.TestStep{
			{
				Config:      template.ParseTestName(testAccClientValidationOnMobile, t.Name()),
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
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccCreateMobileClient, t.Name()),
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
				Config: template.ParseTestName(testAccUpdateMobileClient, t.Name()),
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
				Config: template.ParseTestName(testAccUpdateMobileClientAgainByRemovingSomeFields, t.Name()),
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
				Config: template.ParseTestName(testAccChangeMobileClientToM2M, t.Name()),
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
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccCreateClientWithRefreshToken, t.Name()),
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
				Config: template.ParseTestName(testAccUpdateClientWithRefreshToken, t.Name()),
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
				Config: template.ParseTestName(testAccUpdateClientWithRefreshTokenWhenRemovedFromConfig, t.Name()),
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
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccCreateClientWithJWTConfiguration, t.Name()),
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
				Config: template.ParseTestName(testAccUpdateClientWithJWTConfiguration, t.Name()),
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
				Config: template.ParseTestName(testAccUpdateClientWithJWTConfigurationEmpty, t.Name()),
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
				Config: template.ParseTestName(testAccUpdateClientWithJWTConfigurationRemoved, t.Name()),
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
	cross_origin_auth = false
	client_aliases = [ "https://example.com/audience" ]
	callbacks = [ "https://example.com/callback" ]
	allowed_origins = [ "https://example.com" ]
	allowed_clients = [ "https://allowed.example.com" ]
	grant_types = [ "authorization_code", "http://auth0.com/oauth/grant-type/password-realm", "implicit", "password", "refresh_token" ]
	allowed_logout_urls = [ "https://example.com" ]
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
	initiate_login_uri = "https://example.com/login-uri"
	logo_uri = "https://another-example.com/logoUri"
	organization_require_behavior = "no_prompt"
	organization_usage = "deny"
	sso = true
	sso_disabled = true
	custom_login_page_on = true
	is_first_party = true
	is_token_endpoint_ip_header_trusted = true
	oidc_conformant = true
	cross_origin_auth = true
	client_aliases = [ ]
	callbacks = [ ]
	allowed_origins = [ ]
	allowed_clients = [ ]
	grant_types = [ ]
	allowed_logout_urls = [ ]
	web_origins = [ ]
	client_metadata = {}
}
`

func TestAccClient(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfigCreateWithOnlyRequiredFields, t.Name()),
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
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "0"),
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
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "client_aliases"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "callbacks"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "allowed_logout_urls"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "allowed_origins"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "allowed_clients"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "web_origins"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "encryption_key"),
					resource.TestCheckNoResourceAttr("auth0_client.my_client", "client_metadata"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientConfigUpdateAllFields, t.Name()),
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
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "0"),
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
				),
			},
			{
				Config: template.ParseTestName(testAccClientConfigUpdateSomeFieldsToEmpty, t.Name()),
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
					resource.TestCheckResourceAttr("auth0_client.my_client", "initiate_login_uri", "https://example.com/login-uri"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "logo_uri", "https://another-example.com/logoUri"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "organization_require_behavior", "no_prompt"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "organization_usage", "deny"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "sso", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "sso_disabled", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "custom_login_page_on", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_first_party", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_token_endpoint_ip_header_trusted", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "oidc_conformant", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "cross_origin_auth", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "0"),
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
				),
			},
		},
	})
}

const testAccCreateClientWithAddons = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		firebase = {
			client_email = "john.doe@example.com"
			lifetime_in_seconds = 1
			private_key = "wer"
			private_key_id = "qwreerwerwe"
		}

		samlp {
			issuer = "https://tableau-server-test.domain.eu.com/api/v1"
			audience = "https://tableau-server-test.domain.eu.com/audience-different"
			destination = "https://tableau-server-test.domain.eu.com/destination"
			digest_algorithm = "sha256"
			lifetime_in_seconds = 3600
			name_identifier_format = "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
			name_identifier_probes = [
				"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
			]
			create_upn_claim = false
			passthrough_claims_with_no_mapping = false
			map_unknown_claims_as_is = false
			map_identities = false
			recipient = "https://tableau-server-test.domain.eu.com/recipient-different"
			signing_cert = "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"
			mappings = {
				email = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
				name = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
			}
			logout = {
				callback = "https://example.com/callback"
				slo_enabled = true
			}
		}
	}
}
`

const testAccCreateClientWithAddonsAndEmptyFields = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	addons {
		firebase = {
			client_email = "john.doe@example.com"
			lifetime_in_seconds = 1
			private_key = "wer"
			private_key_id = "qwreerwerwe"
		}

		samlp {
			issuer = "https://tableau-server-test.domain.eu.com/api/v3"
			audience = "https://tableau-server-test.domain.eu.com/audience-different"
			destination = "https://tableau-server-test.domain.eu.com/destination"
			digest_algorithm = "sha256"
			lifetime_in_seconds = 3600
			name_identifier_format = "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
			name_identifier_probes = []
			create_upn_claim = false
			passthrough_claims_with_no_mapping = false
			map_unknown_claims_as_is = false
			map_identities = false
			recipient = "https://tableau-server-test.domain.eu.com/recipient-different"
			signing_cert = "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"
			mappings = {}
			logout = {}
		}
	}
}
`

const testAccCreateClientWithAddonsRemovedFromConfig = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"

	# Unfortunately we can't set firebase and
	# samlp addons set above, to empty.
	# This is because we don't have properly
	# defined structs for them in the Go SDK
	# and neither here in the terraform provider.
}
`

func TestAccClientSSOIntegrationWithSAML(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccCreateClientWithAddons, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.client_email", "john.doe@example.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.lifetime_in_seconds", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.private_key", "wer"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.private_key_id", "qwreerwerwe"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.issuer", "https://tableau-server-test.domain.eu.com/api/v1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.audience", "https://tableau-server-test.domain.eu.com/audience-different"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.destination", "https://tableau-server-test.domain.eu.com/destination"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.digest_algorithm", "sha256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.lifetime_in_seconds", "3600"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.name_identifier_format", "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.name_identifier_probes.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.name_identifier_probes.0", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.create_upn_claim", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.passthrough_claims_with_no_mapping", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.map_unknown_claims_as_is", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.map_identities", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.recipient", "https://tableau-server-test.domain.eu.com/recipient-different"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.signing_cert", "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.mappings.%", "2"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.mappings.email", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.mappings.name", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.logout.%", "2"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.logout.callback", "https://example.com/callback"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.logout.slo_enabled", "true"),
				),
			},
			{
				Config: template.ParseTestName(testAccCreateClientWithAddonsAndEmptyFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.client_email", "john.doe@example.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.lifetime_in_seconds", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.private_key", "wer"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.private_key_id", "qwreerwerwe"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.issuer", "https://tableau-server-test.domain.eu.com/api/v3"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.audience", "https://tableau-server-test.domain.eu.com/audience-different"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.destination", "https://tableau-server-test.domain.eu.com/destination"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.digest_algorithm", "sha256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.lifetime_in_seconds", "3600"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.name_identifier_format", "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.name_identifier_probes.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.create_upn_claim", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.passthrough_claims_with_no_mapping", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.map_unknown_claims_as_is", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.map_identities", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.recipient", "https://tableau-server-test.domain.eu.com/recipient-different"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.signing_cert", "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.mappings.%", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.logout.%", "0"),
				),
			},
			{
				Config: template.ParseTestName(testAccCreateClientWithAddonsRemovedFromConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - SSO Integration - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.%", "4"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.#", "1"),
				),
			},
		},
	})
}

func TestAccClientMetadataBehavior(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(`
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
				Config: template.ParseTestName(`
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
				Config: template.ParseTestName(`
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
				Config: template.ParseTestName(`
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
