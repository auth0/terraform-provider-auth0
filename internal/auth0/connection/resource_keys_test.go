package connection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccKeysRotationForOIDCConnection = `
resource "auth0_client" "my_client" {
  name     = "Acceptance Test {{.testName}}"
  app_type = "non_interactive"
}

resource "auth0_connection" "oidc" {
    name     = "OIDC-Connection-{{.testName}}"
    strategy = "oidc"
    options {
        client_id                     = auth0_client.my_client.id
        scopes                        = ["ext_nested_groups","openid"]
        issuer                        = "https://example.com"
        authorization_endpoint        = "https://example.com"
        jwks_uri                      = "https://example.com/jwks"
        type                          = "front_channel"
        discovery_url                 = "https://www.paypalobjects.com/.well-known/openid-configuration"
        token_endpoint_auth_method    = "private_key_jwt"
        token_endpoint_auth_signing_alg = "RS256"
    }
}

resource "auth0_connection_keys" "my_keys"{
    connection_id = auth0_connection.oidc.id
    triggers = {
        version = "1"
        date   = "2023-10-01T00:00:00Z"
    }
}

data "auth0_connection_keys" "my_keys" {
	depends_on = [auth0_connection_keys.my_keys]
	connection_id = auth0_connection.oidc.id
}
`

func TestAccConnectionRotateDataKeys(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccKeysRotationForOIDCConnection, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_connection_keys.my_keys", "algorithm"),
					resource.TestCheckResourceAttrSet("auth0_connection_keys.my_keys", "cert"),
					resource.TestCheckResourceAttrSet("auth0_connection_keys.my_keys", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_keys.my_keys", "current"),
					resource.TestCheckResourceAttrSet("auth0_connection_keys.my_keys", "fingerprint"),
					resource.TestCheckResourceAttrSet("auth0_connection_keys.my_keys", "key_use"),
					resource.TestCheckResourceAttrSet("auth0_connection_keys.my_keys", "kid"),
					resource.TestCheckResourceAttrSet("auth0_connection_keys.my_keys", "pkcs"),
					resource.TestCheckResourceAttr("auth0_connection_keys.my_keys", "triggers.version", "1"),
					resource.TestCheckResourceAttr("data.auth0_connection_keys.my_keys", "keys.#", "3"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.algorithm"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.cert"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.connection_id"),
					resource.TestCheckResourceAttr("data.auth0_connection_keys.my_keys", "keys.0.current", "true"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.current_since"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.fingerprint"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.key_use"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.kid"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.pkcs"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.subject_dn"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.0.thumbprint"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.1.algorithm"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.1.cert"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.1.connection_id"),
					resource.TestCheckResourceAttr("data.auth0_connection_keys.my_keys", "keys.1.current", "false"),
					resource.TestCheckResourceAttr("data.auth0_connection_keys.my_keys", "keys.1.current_since", ""),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.1.fingerprint"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.1.key_use"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.1.kid"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.1.pkcs"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.1.subject_dn"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_keys.my_keys", "keys.1.thumbprint"),
				),
			},
		},
	})
}
