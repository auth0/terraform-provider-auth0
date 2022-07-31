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

func TestAccClient(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_token_endpoint_ip_header_trusted", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "token_endpoint_auth_method", "client_secret_post"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "allowed_clients.0", "https://allowed.example.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.leeway", "42"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.token_lifetime", "424242"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.rotation_type", "rotating"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.expiration_type", "expiring"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_token_lifetime", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.infinite_idle_token_lifetime", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.idle_token_lifetime", "3600"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.client_email", "john.doe@example.com"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.firebase.lifetime_in_seconds", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.audience", "https://example.com/saml"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.map_identities", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.name_identifier_format", "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.signing_cert", "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_metadata.foo", "zoo"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "initiate_login_uri", "https://example.com/login"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "signing_keys.#", "1"), // checks that signing_keys is set, and it includes 1 element
				),
			},
			{
				Config: template.ParseTestName(testAccClientConfigWithoutAddonsWithSAMLPLogout, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.logout.%", "2"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.logout.callback", "http://example.com/callback"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.logout.slo_enabled", "true"),
				),
			},
		},
	})
}

const testAccClientConfig = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - {{.testName}}"
  description = "Test Application Long Description"
  app_type = "non_interactive"
  custom_login_page_on = true
  is_first_party = true
  is_token_endpoint_ip_header_trusted = true
  token_endpoint_auth_method = "client_secret_post"
  oidc_conformant = true
  callbacks = [ "https://example.com/callback" ]
  allowed_origins = [ "https://example.com" ]
  allowed_clients = [ "https://allowed.example.com" ]
  grant_types = [ "authorization_code", "http://auth0.com/oauth/grant-type/password-realm", "implicit", "password", "refresh_token" ]
  organization_usage = "deny"
  organization_require_behavior = "no_prompt"
  allowed_logout_urls = [ "https://example.com" ]
  web_origins = [ "https://example.com" ]
  jwt_configuration {
    lifetime_in_seconds = 300
    secret_encoded = true
    alg = "RS256"
    scopes = {
      foo = "bar"
    }
  }
  client_metadata = {
    foo = "zoo"
  }
  addons {
    firebase = {
      client_email = "john.doe@example.com"
      lifetime_in_seconds = 1
      private_key = "wer"
      private_key_id = "qwreerwerwe"
    }
    samlp {
      audience = "https://example.com/saml"
      mappings = {
        email = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
        name = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
      }
      create_upn_claim = false
      passthrough_claims_with_no_mapping = false
      map_unknown_claims_as_is = false
      map_identities = false
      name_identifier_format = "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
      name_identifier_probes = [
        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      ]
	  signing_cert = "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"
    }
  }
  refresh_token {
    leeway = 42
    token_lifetime = 424242
    rotation_type = "rotating"
    expiration_type = "expiring"
    infinite_token_lifetime = true
    infinite_idle_token_lifetime = false
    idle_token_lifetime = 3600
  }
  mobile {
    ios {
      team_id = "9JA89QQLNQ"
      app_bundle_identifier = "com.my.bundle.id"
    }
  }
  initiate_login_uri = "https://example.com/login"
}
`

const testAccClientConfigWithoutAddonsWithSAMLPLogout = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - {{.testName}}"
  description = "Test Application Long Description"
  app_type = "non_interactive"
  custom_login_page_on = true
  is_first_party = true
  is_token_endpoint_ip_header_trusted = true
  token_endpoint_auth_method = "client_secret_post"
  oidc_conformant = true
  callbacks = [ "https://example.com/callback" ]
  allowed_origins = [ "https://example.com" ]
  allowed_clients = [ "https://allowed.example.com" ]
  grant_types = [ "authorization_code", "http://auth0.com/oauth/grant-type/password-realm", "implicit", "password", "refresh_token" ]
  organization_usage = "deny"
  organization_require_behavior = "no_prompt"
  allowed_logout_urls = [ "https://example.com" ]
  web_origins = [ "https://example.com" ]
  jwt_configuration {
    lifetime_in_seconds = 300
    secret_encoded = true
    alg = "RS256"
    scopes = {
      foo = "bar"
    }
  }
  client_metadata = {
    foo = "zoo"
  }
  addons {
    firebase = {
      client_email = "john.doe@example.com"
      lifetime_in_seconds = 1
      private_key = "wer"
      private_key_id = "qwreerwerwe"
    }
    samlp {
      audience = "https://example.com/saml"
      mappings = {
        email = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
        name = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
      }
      create_upn_claim = false
      passthrough_claims_with_no_mapping = false
      map_unknown_claims_as_is = false
      map_identities = false
      name_identifier_format = "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
      name_identifier_probes = [
        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      ]
      logout = {
        callback = "http://example.com/callback"
        slo_enabled = true
      }
	  signing_cert = "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"
    }
  }
  refresh_token {
    leeway = 42
    token_lifetime = 424242
    rotation_type = "rotating"
    expiration_type = "expiring"
    infinite_token_lifetime = true
    infinite_idle_token_lifetime = false
    idle_token_lifetime = 3600
  }
  mobile {
    ios {
      team_id = "9JA89QQLNQ"
      app_bundle_identifier = "com.my.bundle.id"
    }
  }
  initiate_login_uri = "https://example.com/login"
}
`

func TestAccClientZeroValueCheck(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Zero Value Check - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_first_party", "false"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_first_party", "true"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientConfigUpdateAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "is_first_party", "false"),
				),
			},
		},
	})
}

const testAccClientConfigCreate = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - Zero Value Check - {{.testName}}"
  is_first_party = false
}
`

const testAccClientConfigUpdate = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - Zero Value Check - {{.testName}}"
  is_first_party = true
}
`

const testAccClientConfigUpdateAgain = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - Zero Value Check - {{.testName}}"
  is_first_party = false
}
`

func TestAccClientRotateSecret(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfigRotateSecret, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Rotate Secret - %s", t.Name())),
				),
			},
			{
				Config: template.ParseTestName(testAccClientConfigRotateSecretUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_secret_rotation_trigger.triggered_at", "2018-01-02T23:12:01Z"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "client_secret_rotation_trigger.triggered_by", "alex"),
				),
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
    triggered_at = "2018-01-02T23:12:01Z"
    triggered_by = "alex"
  }
}
`

func TestAccClientInitiateLoginUri(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config:      template.ParseTestName(testAccClientConfigInitiateLoginURIHTTP, t.Name()),
				ExpectError: regexp.MustCompile("to have a url with schema"),
			},
			{
				Config:      template.ParseTestName(testAccClientConfigInitiateLoginURIFragment, t.Name()),
				ExpectError: regexp.MustCompile("to have a url with an empty fragment"),
			},
		},
	})
}

const testAccClientConfigInitiateLoginURIHTTP = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - Initiate Login URI - {{.testName}}"
  initiate_login_uri = "http://example.com/login"
}
`

const testAccClientConfigInitiateLoginURIFragment = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - Initiate Login URI - {{.testName}}"
  initiate_login_uri = "https://example.com/login#fragment"
}
`

func TestAccClientJwtScopes(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfigJwtScopes, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.secret_encoded", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.lifetime_in_seconds", "300"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.%", "0"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientConfigJwtScopesUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.lifetime_in_seconds", "300"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.%", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.scopes.foo", "bar"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.secret_encoded", "true"),
				),
			},
		},
	})
}

const testAccClientConfigJwtScopes = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - JWT Scopes - {{.testName}}"
  jwt_configuration {
    lifetime_in_seconds = 300
    secret_encoded = true
    alg = "RS256"
    scopes = {}
  }
}
`

const testAccClientConfigJwtScopesUpdate = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - JWT Scopes - {{.testName}}"
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

func TestAccClientMobile(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfigMobile, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.app_package_name", "com.example"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.0", "DE:AD:BE:EF"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.0.enabled", "false"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientConfigMobileUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.app_package_name", "com.example"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "mobile.0.android.0.sha256_cert_fingerprints.#", "0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.apple.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "native_social_login.0.facebook.0.enabled", "true"),
				),
			},
			{
				// This just makes sure that you can change the type (where native_social_login cannot be set)
				Config: template.ParseTestName(testAccClientConfigMobileUpdateNonMobile, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
				),
			},
		},
	})
}

const testAccClientConfigMobile = `
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

const testAccClientConfigMobileUpdate = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - Mobile - {{.testName}}"
  app_type = "native"
  mobile {
    android {
      app_package_name = "com.example"
      sha256_cert_fingerprints = []
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

const testAccClientConfigMobileUpdateNonMobile = `

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

func TestAccClientMobileValidationError(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config:      template.ParseTestName(testAccClientConfigMobileUpdateError, t.Name()),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
		},
	})
}

const testAccClientConfigMobileUpdateError = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - Mobile - {{.testName}}"
  mobile {
    android {
      # nothing specified, should throw validation error
    }
  }
}
`

func TestAccClientRefreshTokenApplied(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientConfigWithRefreshToken, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Refresh Token - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.rotation_type", "non-rotating"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "refresh_token.0.expiration_type", "non-expiring"),

					resource.TestCheckResourceAttrSet("auth0_client.my_client", "refresh_token.0.token_lifetime"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "refresh_token.0.infinite_token_lifetime"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "refresh_token.0.infinite_idle_token_lifetime"),
					resource.TestCheckResourceAttrSet("auth0_client.my_client", "refresh_token.0.idle_token_lifetime"),
				),
			},
		},
	})
}

const testAccClientConfigWithRefreshToken = `
resource "auth0_client" "my_client" {
	name                       = "Acceptance Test - Refresh Token - {{.testName}}"
	app_type                   = "spa"

	refresh_token {
	  rotation_type   = "non-rotating"
	  expiration_type = "non-expiring"
	  # Intentionally not setting leeway, token_lifetime, infinite_token_lifetime, infinite_idle_token_lifetime, idle_token_lifetime because those get inferred by Auth0 defaults
	}
  }
`

func TestAccClientSSOIntegration(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccClientSSOIntegrationCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.audience", "http://tableau-server-test.domain.eu.com/audience"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.recipient", "http://tableau-server-test.domain.eu.com/recipient"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.destination", "http://tableau-server-test.domain.eu.com/destination"),
				),
			},
			{
				Config: template.ParseTestName(testAccClientSSOIntegrationUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "sso_integration"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.audience", "http://tableau-server-test.domain.eu.com/audience-different"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.recipient", "http://tableau-server-test.domain.eu.com/recipient-different"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "addons.0.samlp.0.destination", "http://tableau-server-test.domain.eu.com/destination"),
				),
			},
		},
	})
}

const testAccClientSSOIntegrationCreate = `
resource "auth0_client" "my_client" {
  name = "Acceptance Test - SSO Integration - {{.testName}}"
  app_type = "sso_integration"
  addons{
	samlp {
		audience= "http://tableau-server-test.domain.eu.com/audience"
		destination= "http://tableau-server-test.domain.eu.com/destination"
		digest_algorithm= "sha256"
		lifetime_in_seconds= 3600
		mappings= {
			email= "username"
		}
		name_identifier_format= "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
		passthrough_claims_with_no_mapping= false
		recipient= "http://tableau-server-test.domain.eu.com/recipient"
	}
  }
}
`

const testAccClientSSOIntegrationUpdate = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - SSO Integration - {{.testName}}"
	app_type = "sso_integration"
	addons{
	  samlp {
		audience= "http://tableau-server-test.domain.eu.com/audience-different"
		destination= "http://tableau-server-test.domain.eu.com/destination"
		digest_algorithm= "sha256"
		lifetime_in_seconds= 3600
		mappings= {
			email= "username"
		}
		name_identifier_format= "urn:oasis:names:tc:SAML:2.0:attrname-format:basic"
		passthrough_claims_with_no_mapping= false
		recipient= "http://tableau-server-test.domain.eu.com/recipient-different"
	  }
	}
  }
`
