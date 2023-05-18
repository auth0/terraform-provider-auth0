package client_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccThrowErrorWhenPrivateKeyJWT = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Client Credentials - {{.testName}}"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id = auth0_client.my_client.id

	authentication_method = "private_key_jwt"
}
`

const testAccCreateOneClientCredentialUsingPrivateKeyJWT = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Client Credentials - {{.testName}}"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id = auth0_client.my_client.id

	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			pem = <<EOF
%s
EOF
		}
	}
}
`

const testAccAddAnotherClientCredentialUsingPrivateKeyJWT = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - Client Credentials - {{.testName}}"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id = auth0_client.my_client.id

	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			pem = <<EOF
%s
EOF
		}
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			parse_expiry_from_cert = false
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem = <<EOF
%s
EOF
		}
	}
}
`

func TestClientAuthenticationMethods(t *testing.T) {
	credsCert1, err := os.ReadFile("./../../../test/data/creds-cert-1.pem")
	require.NoError(t, err)

	credsCert2, err := os.ReadFile("./../../../test/data/creds-cert-2.pem")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenPrivateKeyJWT, t.Name()),
				ExpectError: regexp.MustCompile("Client Credentials Missing"),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccCreateOneClientCredentialUsingPrivateKeyJWT, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "private_key_jwt"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.parse_expiry_from_cert", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.expires_at", "2033-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.updated_at"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccAddAnotherClientCredentialUsingPrivateKeyJWT, t.Name()), credsCert1, credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "private_key_jwt"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.#", "2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.parse_expiry_from_cert", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.expires_at", "2033-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.updated_at"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.name", "Testing Credentials 2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.parse_expiry_from_cert", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.expires_at", "2025-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.updated_at"),
				),
			},
		},
	})
}
